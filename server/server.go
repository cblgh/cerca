package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"cerca/crypto"
	"cerca/database"
	"cerca/defaults"
	cercaHTML "cerca/html"
	"cerca/i18n"
	"cerca/limiter"
	"cerca/server/session"
	"cerca/types"
	"cerca/util"

	"github.com/carlmjohnson/requests"
	"github.com/cblgh/plain/rss"
)

/* TODO (2022-01-03): include csrf token via gorilla, or w/e, when rendering */

type TemplateData struct {
	Data       interface{}
	QuickNav   bool
	LoggedIn   bool 
	IsAdmin    bool 
	HasRSS     bool
	LoggedInID int
	ForumName  string
	Title      string
}

type PasswordResetData struct {
	Action   string
	Username string
	Payload  string
}

type ChangePasswordData struct {
	Action  string
}

type IndexData struct {
	Threads []database.Thread
}

type GenericMessageData struct {
	Title       string
	Message     string
	LinkMessage string
	Link        string
	LinkText    string
}

type RegisterData struct {
	VerificationCode         string
	ErrorMessage             string
	Rules                    template.HTML
	VerificationInstructions template.HTML
	ConductLink              string
}

type LoginData struct {
	FailedAttempt bool
}

type ThreadData struct {
	Title     string
	Posts     []database.Post
	ThreadURL string
}

type EditPostData struct {
	Title   string
	Content string
}

type RequestHandler struct {
	db         *database.DB
	session    *session.Session
	allowlist  []string // allowlist of domains valid for forum registration
	files      map[string][]byte
	config     types.Config
	translator i18n.Translator
	templates  *template.Template
	rssFeed    string
}

var developing bool

func dump(err error) {
	if developing {
		fmt.Println(err)
	}
}

type RateLimitingWare struct {
	limiter *limiter.TimedRateLimiter
}

func NewRateLimitingWare(routes []string) *RateLimitingWare {
	ware := RateLimitingWare{}
	// refresh one access every 15 minutes. forget about the requester after 24h of non-activity
	ware.limiter = limiter.NewTimedRateLimiter(routes, 15*time.Minute, 24*time.Hour)
	// allow 15 requests at once, then
	ware.limiter.SetBurstAllowance(25)
	return &ware
}

func (ware *RateLimitingWare) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		portIndex := strings.LastIndex(req.RemoteAddr, ":")
		ip := req.RemoteAddr[:portIndex]
		// specific fix in case of using a reverse proxy setup
		if address, exists := req.Header["X-Real-Ip"]; ip == "127.0.0.1" && exists {
			ip = address[0]
		}
		// rate limiting likely not working as intended on server;
		// set a x-real-ip header: https://docs.nginx.com/nginx/admin-guide/web-server/reverse-proxy/
		if !developing && ip == "127.0.0.1" {
			next.ServeHTTP(res, req)
			return
		}
		err := ware.limiter.BlockUntilAllowed(ip, req.URL.String(), req.Context())
		if err != nil {
			err = util.Eout(err, "RateLimitingWare")
			dump(err)
			http.Error(res, "An error occured", http.StatusInternalServerError)
			return
		}
		next.ServeHTTP(res, req)
	})
}

// returns true if logged in, and the userid of the logged in user.
// returns false (and userid set to -1) if not logged in
func (h RequestHandler) IsLoggedIn(req *http.Request) (bool, int) {
	ed := util.Describe("IsLoggedIn")
	userid, err := h.session.Get(req)
	err = ed.Eout(err, "getting userid from session cookie")
	if err != nil {
		dump(err)
		return false, -1
	}

	// make sure the user from the cookie actually exists
	userExists, err := h.db.CheckUserExists(userid)
	if err != nil {
		dump(ed.Eout(err, "check userid in db"))
		return false, -1
	} else if !userExists {
		return false, -1
	}
	return true, userid
}

// establish closure over config + translator so that it's present in templates during render
func generateTemplates(config types.Config, translator i18n.Translator) (*template.Template, error) {
	// only read logo contents once when generating
	logo, err := os.ReadFile(config.Documents.LogoPath)
	util.Check(err, "generate-template: dump logo")
	templateFuncs := template.FuncMap{
		"dumpLogo": func() template.HTML {
			return template.HTML(logo)
		},
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDateTimeRFC3339": func(t time.Time) string {
			return t.Format(time.RFC3339Nano)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateTimeRelative": util.RelativeTime,
		"formatDateRelative": func(t time.Time) string {
			diff := time.Since(t)
			if diff < time.Hour*24 {
				return "today"
			} else if diff >= time.Hour*24 && diff < time.Hour*48 {
				return "yesterday"
			}
			return t.Format("2006-01-02")
		},
		"translate": func(key string) string {
			return translator.Translate(key)
		},
		"translateWithData": func(key string) string {
			data := struct {
				Name string
				Link string
			}{
				Name: config.Community.Name,
				Link: config.Community.ConductLink,
			}
			return translator.TranslateWithData(key, i18n.TranslationData{data})
		},
		"capitalize": util.Capitalize,
		"markup":     util.Markup,
		"tohtml": func(s string) template.HTML {
			// use of this function is risky cause it interprets the passed in string and renders it as unescaped html.
			// can allow for attacks!
			//
			// advice: only use on strings that come statically from within cerca code, never on titles that may contain user-submitted data
			// :)
			return (template.HTML)(s)
		},
	}
	views := []string{
		"about",
		"about-template",
		"footer",
		"generic-message",
		"head",
		"edit-post",
		"index",
		"login",
		"login-component",
		"new-thread",
		"register",
		"register-success",
		"thread",
		"admin",
		"admins-list",
		"admin-add-user",
		"moderation-log",
		"password-reset",
		"change-password",
		"change-password-success",
	}

	rootTemplate := template.New("root")

	for _, view := range views {
		newTemplate, err := rootTemplate.Funcs(templateFuncs).ParseFS(cercaHTML.Templates, fmt.Sprintf("%s.html", view))
		if err != nil {
			return nil, fmt.Errorf("could not get files: %w", err)
		}
		rootTemplate = newTemplate
	}

	return rootTemplate, nil
}

func (h RequestHandler) renderView(res http.ResponseWriter, viewName string, data TemplateData) {
	if data.Title == "" {
		data.Title = strings.ReplaceAll(viewName, "-", " ")
	}

	if h.config.Community.Name != "" {
		data.ForumName = h.config.Community.Name
	}
	if data.ForumName == "" {
		data.ForumName = "Forum"
	}

	view := fmt.Sprintf("%s.html", viewName)
	if err := h.templates.ExecuteTemplate(res, view, data); err != nil {
		if errors.Is(err, syscall.EPIPE) {
			fmt.Println("recovering from broken pipe")
			return
		} else {
			util.Check(err, "rendering %q view", view)
		}
	}
}

func (h RequestHandler) renderGenericMessage(res http.ResponseWriter, req *http.Request, incomingData GenericMessageData) {
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)
	data := TemplateData{
		Data: incomingData,
		// the following two fields are defaults that usually are not set and which are cumbersome to set each time since
		// they don't really matter / vary across invocations
		HasRSS: h.config.RSS.URL != "",
		LoggedIn: loggedIn,
		IsAdmin: isAdmin,
	}
	h.renderView(res, "generic-message", data)
	return
}

func (h *RequestHandler) ThreadRoute(res http.ResponseWriter, req *http.Request) {
	threadid, ok := util.GetURLPortion(req, 2)
	loggedIn, userid := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)

	if !ok {
		title := h.translator.Translate("ErrThread404")
		data := GenericMessageData{
			Title:   title,
			Message: h.translator.Translate("ErrThread404Message"),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	if req.Method == "POST" && loggedIn {
		// handle POST (=> add a reply, then show the thread)
		content := req.PostFormValue("content")
		// TODO (2022-01-09): make sure rendered content won't be empty after sanitizing:
		// * run sanitize step && strings.TrimSpace and check length **before** doing AddPost
		// TODO(2022-01-09): send errors back to thread's posting view
		_ = h.db.AddPost(content, threadid, userid)
		// we want to effectively redirect to <#posts+1> to mark the thread as read in the thread index
		// TODO(2022-01-30): find a solution for either:
		// * scrolling to thread bottom (and maintaining the same slug, important for visited state in browser)
		// * passing data to signal "your post was successfully added" (w/o impacting visited state / url)
		posts := h.db.GetThread(threadid)
		newSlug := util.GetThreadSlug(threadid, posts[0].ThreadTitle, len(posts))
		// update the rss feed
		h.rssFeed = GenerateRSS(h.db, h.config)
		http.Redirect(res, req, newSlug, http.StatusFound)
		return
	}
	// TODO (2022-01-07):
	// * handle error
	thread := h.db.GetThread(threadid)
	data := ThreadData{Posts: thread, ThreadURL: req.URL.Path}
	view := TemplateData{Data: &data, IsAdmin: isAdmin, QuickNav: loggedIn, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, LoggedInID: userid}
	if len(thread) > 0 {
		data.Title = thread[0].ThreadTitle
		view.Title = data.Title
	}
	h.renderView(res, "thread", view)
}

func (h RequestHandler) ErrorRoute(res http.ResponseWriter, req *http.Request, status int) {
	title := h.translator.Translate("ErrGeneric404")
	data := GenericMessageData{
		Title:   title,
		Message: fmt.Sprintf(h.translator.Translate("ErrGeneric404Message"), status),
	}
	h.renderGenericMessage(res, req, data)
}

func (h RequestHandler) IndexRoute(res http.ResponseWriter, req *http.Request) {
	// handle 404
	if req.URL.Path != "/" {
		h.ErrorRoute(res, req, http.StatusNotFound)
		return
	}
	loggedIn, _ := h.IsLoggedIn(req)
	var mostRecentPost bool
	isAdmin, _ := h.IsAdmin(req)

	params := req.URL.Query()
	if q, exists := params["sort"]; exists {
		sortby := q[0]
		mostRecentPost = sortby == "posts"
	}
	// show index listing
	threads := h.db.ListThreads(mostRecentPost)
	view := TemplateData{Data: IndexData{threads}, IsAdmin: isAdmin, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: h.translator.Translate("Threads")}
	h.renderView(res, "index", view)
}

func IndexRedirect(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/", http.StatusSeeOther)
}

const rfc822RSS = "Mon, 02 Jan 2006 15:04:05 -0700"

func joinPath(host, upath string) string {
	host = strings.TrimSuffix(host, "/")
	upath = strings.TrimPrefix(upath, "/")
	return fmt.Sprintf("%s/%s", host, upath)
}

func GenerateRSS(db *database.DB, config types.Config) string {
	if config.RSS.URL == "" {
		return "feed not configured"
	}
	// TODO (2022-12-08): augment ListThreads to choose getting author of latest post or thread creator (currently latest
	// post always)
	threads := db.ListThreads(true)
	entries := make([]string, len(threads))
	for i, t := range threads {
		fulltime := t.Publish.Format(rfc822RSS)
		date := t.Publish.Format("2006-01-02")
		posturl := joinPath(config.RSS.URL, fmt.Sprintf("%s#%d", t.Slug, t.PostID))
		entry := rss.OutputRSSItem(fulltime, t.Title, fmt.Sprintf("[%s] %s posted", date, t.Author), posturl)
		entries[i] = entry
	}
	feedName := config.RSS.Name
	if feedName == "" {
		feedName = config.Community.Name
	}
	feed := rss.OutputRSS(feedName, config.RSS.URL, config.RSS.Description, entries)
	return feed
}

func (h *RequestHandler) RSSRoute(res http.ResponseWriter, req *http.Request) {
	// error if feed not configured (e.g. config.RSS.URL not set)
	if h.config.RSS.URL == "" {
		http.Error(res, "Feed Not Configured", http.StatusNotFound)
		return
	}
	res.Header().Set("Content-Type", "application/xml")
	res.Write([]byte(h.rssFeed))
}

func (h RequestHandler) LogoutRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	if loggedIn {
		h.session.Delete(res, req)
	}
	IndexRedirect(res, req)
}

func (h RequestHandler) LoginRoute(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("LoginRoute")
	loggedIn, _ := h.IsLoggedIn(req)
	switch req.Method {
	case "GET":
		h.renderView(res, "login", TemplateData{Data: LoginData{}, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: h.translator.Translate("Login")})
	case "POST":
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		// * hash received password and compare to stored hash
		passwordHash, userid, err := h.db.GetPasswordHash(username)
		// make sure user exists
		if err = ed.Eout(err, "getting password hash and uid"); err == nil && !crypto.ValidatePasswordHash(password, passwordHash) {
			err = errors.New("incorrect password")
		}
		if err != nil {
			fmt.Println(err)
			h.renderView(res, "login", TemplateData{Data: LoginData{FailedAttempt: true}, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: h.translator.Translate("Login")})
			return
		}
		// save user id in cookie
		err = h.session.Save(req, res, userid)
		ed.Check(err, "saving session cookie")
		IndexRedirect(res, req)
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

// downloads the content at the verification link and compares it to the verification code. returns true if the verification link content contains the verification code somewhere
func hasVerificationCode(link, verification string) bool {
	var linkBody string
	err := requests.
		URL(link).
		ToString(&linkBody).
		Fetch(context.Background())
	if err != nil {
		fmt.Println(util.Eout(err, "HasVerificationCode"))
		return false
	}

	return strings.Contains(strings.TrimSpace(linkBody), strings.TrimSpace(verification))
}

func (h RequestHandler) handleChangePassword(res http.ResponseWriter, req *http.Request) {
	// TODO (2022-10-24): add translations for change password view
	title := h.translator.Translate("ChangePassword")
	renderErr := func(errFmt string, args ...interface{}) {
		errMessage := fmt.Sprintf(errFmt, args...)
		fmt.Println(errMessage)
		data := GenericMessageData{
			Title:    title,
			Message:  errMessage,
			Link:     "/reset",
			LinkText: h.translator.Translate("GoBack"),
		}
		h.renderGenericMessage(res, req, data)
	}
	_, uid := h.IsLoggedIn(req)

	ed := util.Describe("change password")
	switch req.Method {
	case "GET":
		switch req.URL.Path {
		default:
			h.renderView(res, "change-password", TemplateData{HasRSS: h.config.RSS.URL != "", LoggedIn: true, Data: ChangePasswordData{Action: "/reset/submit"}})
		}
	case "POST":
		switch req.URL.Path {
		case "/reset/submit":
			oldPassword := req.PostFormValue("password-old")
			newPassword := req.PostFormValue("password-new")

			// check that the submitted, old password is valid
			username, err := h.db.GetUsername(uid)
			if err != nil {
				dump(ed.Eout(err, "get username"))
				return
			}

			pwhashOld, _, err := h.db.GetPasswordHash(username)
			if err != nil {
				dump(ed.Eout(err, "get old password hash"))
				return
			}

			oldPasswordValid := crypto.ValidatePasswordHash(oldPassword, pwhashOld)
			if !oldPasswordValid {
				renderErr("old password did not match what was in database; not changing password")
				return
			}

			// let's set the new password in the database. first, hash it
			pwhashNew, err := crypto.HashPassword(newPassword)
			if err != nil {
				dump(ed.Eout(err, "hash new password"))
				return
			}
			// then save the hash
			h.db.UpdateUserPasswordHash(uid, pwhashNew)
			// render a success message & show a link to the login page :')
			h.renderView(res, "change-password-success", TemplateData{HasRSS: h.config.RSS.URL != "", LoggedIn: true, Data: ChangePasswordData{}})
		default:
			fmt.Printf("unsupported POST route (%s), redirecting to /\n", req.URL.Path)
			IndexRedirect(res, req)
		}
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

func (h RequestHandler) ResetPasswordRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	title := util.Capitalize(h.translator.Translate("PasswordReset"))

	// the user was logged in, let them change their password themselves :)
	if loggedIn {
		h.handleChangePassword(res, req)
		return
	}

	renderPlaceholder := func(errFmt string, args ...interface{}) {
		errMessage := fmt.Sprintf(errFmt, args...)
		fmt.Println(errMessage)
		data := GenericMessageData{
			Title:    title,
			Message:  errMessage,
			Link:     "/",
			LinkText: h.translator.Translate("GoBack"),
		}
		h.renderView(res, "generic-message", TemplateData{Data: data, Title: title})
	}
	renderPlaceholder("Password reset under construction: please contact admin if you need help resetting yr pw :)")
	return
}

func (h RequestHandler) RegisterRoute(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("register route")
	loggedIn, _ := h.IsLoggedIn(req)
	if loggedIn {
		// TODO (2022-09-20): translate
		data := GenericMessageData{
			Title:       util.Capitalize(h.translator.Translate("Register")),
			Message:     h.translator.Translate("RegisterMessage"),
			Link:        "/",
			LinkMessage: h.translator.Translate("RegisterLinkMessage"),
			LinkText:    h.translator.Translate("Index"),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	rules := util.Markup(string(h.files["rules"]))
	verification := util.Markup(string(h.files["verification-instructions"]))
	conduct := h.config.Community.ConductLink
	var verificationCode string
	renderErr := func(errFmt string, args ...interface{}) {
		errMessage := fmt.Sprintf(errFmt, args...)
		fmt.Println(errMessage)
		h.renderView(res, "register", TemplateData{Data: RegisterData{verificationCode, errMessage, rules, verification, conduct}})
	}

	var err error
	switch req.Method {
	case "GET":
		// try to get the verification code from the session (useful in case someone refreshed the page)
		verificationCode, err = h.session.GetVerificationCode(req)
		// we had an error getting the verification code, generate a code and set it on the session
		if err != nil {
			prefix := util.VerificationPrefix(h.config.Community.Name)
			verificationCode = fmt.Sprintf("%s%06d\n", prefix, crypto.GenerateVerificationCode())
			err = h.session.SaveVerificationCode(req, res, verificationCode)
			if err != nil {
				renderErr("Had troubles setting the verification code on session")
				return
			}
		}
		h.renderView(res, "register", TemplateData{Data: RegisterData{verificationCode, "", rules, verification, conduct}})
	case "POST":
		verificationCode, err = h.session.GetVerificationCode(req)
		if err != nil {
			renderErr("There was no verification record for this browser session; missing data to compare against verification link content")
			return
		}
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		var verificationLink string
		// skip verification code during dev registering
		if !developing {
			// read verification code from form
			verificationLink = req.PostFormValue("verificationlink")
			// fmt.Printf("user: %s, verilink: %s\n", username, verificationLink)
			u, err := url.Parse(verificationLink)
			if err != nil {
				renderErr("Had troubles parsing the verification link, are you sure it was a proper url?")
				return
			}
			// check verification link domain against allowlist
			if !util.Contains(h.allowlist, u.Host) {
				fmt.Println(h.allowlist, u.Host, util.Contains(h.allowlist, u.Host))
				renderErr("Verification link's host (%s) is not in the allowlist", u.Host)
				return
			}

			// parse out verification code from verification link and compare against verification code in session
			has := hasVerificationCode(verificationLink, verificationCode)
			if !has {
				if !developing {
					renderErr("Verification code from link (%s) does not match", verificationLink)
					return
				}
			}
		}
		// make sure username is not registered already
		var exists bool
		if exists, err = h.db.CheckUsernameExists(username); err != nil {
			renderErr("Database had a problem when checking username")
			return
		} else if exists {
			renderErr("Username %s appears to already exist, please pick another name", username)
			return
		}
		var hash string
		if hash, err = crypto.HashPassword(password); err != nil {
			fmt.Println(ed.Eout(err, "hash password"))
			renderErr("Database had a problem when hashing password")
			return
		}
		var userID int
		if userID, err = h.db.CreateUser(username, hash); err != nil {
			renderErr("Error in db when creating user")
			return
		}
		// log the new user in
		h.session.Save(req, res, userID)
		// log where the registration is coming from, in the case of indirect invites && for curiosity
		err = h.db.AddRegistration(userID, verificationLink)
		if err = ed.Eout(err, "add registration"); err != nil {
			dump(err)
		}
		h.renderView(res, "register-success", TemplateData{HasRSS: h.config.RSS.URL != "", LoggedIn: true, Title: h.translator.Translate("RegisterSuccess")})
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

// purely an example route; intentionally unused :)
func (h RequestHandler) GenericRoute(res http.ResponseWriter, req *http.Request) {
	data := GenericMessageData{
		Title:       "GenericTitle",
		Message:     "Generic message",
		Link:        "/",
		LinkMessage: "Generic link messsage",
		LinkText:    "with link",
	}
	h.renderGenericMessage(res, req, data)
}

func (h RequestHandler) AboutRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	input := util.Markup(string(h.files["about"]))
	h.renderView(res, "about-template", TemplateData{Data: input, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: h.translator.Translate("About")})
}

func (h RequestHandler) RobotsRoute(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "User-agent: *\nDisallow: /")
}

func (h *RequestHandler) NewThreadRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	switch req.Method {
	// Handle GET (=> want to start a new thread)
	case "GET":
		// TODO (2022-09-20): translate
		if !loggedIn {
			title := h.translator.Translate("NotLoggedIn")
			data := GenericMessageData{
				Title:       title,
				Message:     h.translator.Translate("NewThreadMessage"),
				Link:        "/login",
				LinkMessage: h.translator.Translate("NewThreadLinkMessage"),
				LinkText:    h.translator.Translate("LogIn"),
			}
			h.renderGenericMessage(res, req, data)
			return
		}
		h.renderView(res, "new-thread", TemplateData{HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: h.translator.Translate("ThreadNew")})
	case "POST":
		// Handle POST (=>
		title := req.PostFormValue("title")
		content := req.PostFormValue("content")
		// TODO (2022-01-10): unstub topicid, once we have other topics :)
		// the new thread was created: forward info to database
		threadid, err := h.db.CreateThread(title, content, userid, 1)
		if err != nil {
			data := GenericMessageData{
				Title:   h.translator.Translate("NewThreadCreateError"),
				Message: h.translator.Translate("NewThreadCreateErrorMessage"),
			}
			h.renderGenericMessage(res, req, data)
			return
		}
		// update the rss feed
		h.rssFeed = GenerateRSS(h.db, h.config)
		// when data has been stored => redirect to thread
		slug := fmt.Sprintf("thread/%d/%s/", threadid, util.SanitizeURL(title))
		http.Redirect(res, req, "/"+slug, http.StatusSeeOther)
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

func (h *RequestHandler) DeletePostRoute(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		IndexRedirect(res, req)
		return
	}
	threadURL := req.PostFormValue("thread")
	postid, ok := util.GetURLPortion(req, 3)
	loggedIn, userid := h.IsLoggedIn(req)

	// generic error message base, with specifics being swapped out depending on the error
	genericErr := GenericMessageData{
		Title:       h.translator.Translate("ErrUnaccepted"),
		LinkMessage: h.translator.Translate("GoBack"),
		Link:        threadURL,
		LinkText:    h.translator.Translate("ThreadThe"),
	}

	renderErr := func(msg string) {
		fmt.Println(msg)
		genericErr.Message = msg
		h.renderGenericMessage(res, req, genericErr)
	}

	if !loggedIn || !ok {
		renderErr("Invalid post id, or you were not allowed to delete it")
		return
	}

	post, err := h.db.GetPost(postid)
	if err != nil {
		dump(err)
		renderErr("The post you tried to delete was not found")
		return
	}

	authorized := post.AuthorID == userid
	switch req.Method {
	case "POST":
		if authorized {
			err = h.db.DeletePost(postid)
			if err != nil {
				dump(err)
				renderErr("Error happened while deleting the post")
				return
			}
		} else {
			renderErr("That's not your post to delete? Sorry buddy!")
			return
		}
		// update the rss feed, in case the deleted post was present in feed
		h.rssFeed = GenerateRSS(h.db, h.config)
	}
	http.Redirect(res, req, threadURL, http.StatusSeeOther)
}

func (h *RequestHandler) EditPostRoute(res http.ResponseWriter, req *http.Request) {
	postid, ok := util.GetURLPortion(req, 3)
	loggedIn, userid := h.IsLoggedIn(req)
	post, err := h.db.GetPost(postid)

	if !ok || errors.Is(err, sql.ErrNoRows) {
		title := h.translator.Translate("ErrEdit404")
		data := GenericMessageData{
			Title:   title,
			Message: h.translator.Translate("ErrEdit404Message"),
		}
		h.renderGenericMessage(res, req, data)
		return
	}
	if !loggedIn || userid != post.AuthorID {
		res.WriteHeader(401)
		title := h.translator.Translate("ErrGeneric401")
		data := GenericMessageData{
			Title:   title,
			Message: h.translator.Translate("ErrGeneric401Message"),
		}
		h.renderGenericMessage(res, req, data)
		return
	}
	if req.Method == "POST" {
		content := req.PostFormValue("content")
		h.db.EditPost(content, postid)
		post.Content = content
	}
	view := TemplateData{Data: post, QuickNav: loggedIn, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, LoggedInID: userid}
	h.renderView(res, "edit-post", view)
}

func Serve(allowlist []string, sessionKey string, isdev bool, dir string, conf types.Config) {
	port := ":8272"

	if isdev {
		developing = true
		port = ":8277"
	}

	forum, err := NewServer(allowlist, sessionKey, dir, conf)
	if err != nil {
		util.Check(err, "instantiate CercaForum")
	}

	l, err := net.Listen("tcp", port)
	if err != nil {
		util.Check(err, "setting up tcp listener")
	}
	fmt.Println("Serving forum on", port)

	rateLimitingInstance := NewRateLimitingWare([]string{"/rss/", "/rss.xml"})
	limitingMiddleware := rateLimitingInstance.Handler(forum)
	http.Serve(l, limitingMiddleware)
}

// CercaForum is an HTTP.ServeMux which is set up to initialize and run
// a cerca-based forum. Software developers who wish to customize the
// networks and security which they use to operate may wish to use this
// to listen with TLS, Onion, or I2P addresses without manual setup.
type CercaForum struct {
	http.ServeMux
	Directory string
}

func (u *CercaForum) directory() string {
	if u.Directory == "" {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		u.Directory = filepath.Join(dir, "CercaForum")
	}
	os.MkdirAll(u.Directory, 0755)
	return u.Directory
}

// NewServer sets up a new CercaForum object. Always use this to initialize
// new CercaForum objects. Pass the result to http.Serve() with your choice
// of net.Listener.
func NewServer(allowlist []string, sessionKey, dir string, config types.Config) (*CercaForum, error) {
	s := &CercaForum{
		ServeMux:  http.ServeMux{},
		Directory: dir,
	}

	dbpath := filepath.Join(s.directory(), "forum.db")
	db := database.InitDB(dbpath)

	config.EnsureDefaultPaths()
	// load the documents specified in the config
	// iff document doesn't exist, dump a default document where it should be and read that
	type triple struct{ key, docpath, content string }
	triples := []triple{
		{"about", config.Documents.AboutPath, defaults.DEFAULT_ABOUT},
		{"rules", config.Documents.RegisterRulesPath, defaults.DEFAULT_RULES},
		{"verification-instructions", config.Documents.VerificationExplanationPath, defaults.DEFAULT_VERIFICATION},
		{"logo", config.Documents.LogoPath, defaults.DEFAULT_LOGO},
	}

	files := make(map[string][]byte)
	for _, t := range triples {
		data, err := util.LoadFile(t.key, t.docpath, t.content)
		if err != nil {
			return s, err
		}
		files[t.key] = data
	}

	// TODO (2022-10-20): when receiving user request, inspect user-agent language and change language from server default
	// for currently translated languages, see i18n/i18n.go
	translator := i18n.Init(config.Community.Language)
	templates := template.Must(generateTemplates(config, translator))
	feed := GenerateRSS(&db, config)
	handler := RequestHandler{&db, session.New(sessionKey, developing), allowlist, files, config, translator, templates, feed}

	/* note: be careful with trailing slashes; go's default handler is a bit sensitive */
	// TODO (2022-01-10): introduce middleware to make sure there is never an issue with trailing slashes

	// moderation and admin related routes, for contents see file server/moderation.go
	s.ServeMux.HandleFunc("/reset/", handler.ResetPasswordRoute)
	s.ServeMux.HandleFunc("/admin", handler.AdminRoute)
	s.ServeMux.HandleFunc("/demote-admin", handler.AdminDemoteAdmin)
	s.ServeMux.HandleFunc("/add-user", handler.AdminManualAddUserRoute)
	s.ServeMux.HandleFunc("/moderations", handler.ModerationLogRoute)
	s.ServeMux.HandleFunc("/proposal-veto", handler.VetoProposal)
	s.ServeMux.HandleFunc("/proposal-confirm", handler.ConfirmProposal)
	// regular ol forum routes
	s.ServeMux.HandleFunc("/about", handler.AboutRoute)
	s.ServeMux.HandleFunc("/logout", handler.LogoutRoute)
	s.ServeMux.HandleFunc("/login", handler.LoginRoute)
	s.ServeMux.HandleFunc("/register", handler.RegisterRoute)
	s.ServeMux.HandleFunc("/post/delete/", handler.DeletePostRoute)
	s.ServeMux.HandleFunc("/post/edit/", handler.EditPostRoute)
	s.ServeMux.HandleFunc("/thread/new/", handler.NewThreadRoute)
	s.ServeMux.HandleFunc("/thread/", handler.ThreadRoute)
	s.ServeMux.HandleFunc("/robots.txt", handler.RobotsRoute)
	s.ServeMux.HandleFunc("/", handler.IndexRoute)
	s.ServeMux.HandleFunc("/rss/", handler.RSSRoute)
	s.ServeMux.HandleFunc("/rss.xml", handler.RSSRoute)

	fileserver := http.FileServer(http.Dir("html/assets/"))
	s.ServeMux.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	return s, nil
}
