package server

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strings"
	"time"

	"cerca/crypto"
	"cerca/database"
	cercaHTML "cerca/html"
	"cerca/server/session"
	"cerca/util"

	"github.com/carlmjohnson/requests"
)

/* TODO (2022-01-03): include csrf token via gorilla, or w/e, when rendering */

type TemplateData struct {
	Data       interface{}
	QuickNav   bool
	LoggedIn   bool // TODO (2022-01-09): put this in a middleware || template function or sth?
	LoggedInID int
	Title      string
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
	VerificationCode string
	ErrorMessage     string
}

type RegisterSuccessData struct {
	Keypair string
}

type LoginData struct {
	FailedAttempt bool
}

type ThreadData struct {
	Title     string
	Posts     []database.Post
	ThreadURL string
}

type RequestHandler struct {
	db        *database.DB
	session   *session.Session
	allowlist []string // allowlist of domains valid for forum registration
}

var developing bool

func dump(err error) {
	if developing {
		fmt.Println(err)
	}
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

var (
	templateFuncs = template.FuncMap{
		"formatDateTime": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDateTimeRFC3339": func(t time.Time) string {
			return t.Format(time.RFC3339Nano)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"formatDateRelative": func(t time.Time) string {
			diff := time.Since(t)
			if diff < time.Hour*24 {
				return "today"
			} else if diff >= time.Hour*24 && diff < time.Hour*48 {
				return "yesterday"
			}
			return t.Format("2006-01-02")
		},
	}

	templates = template.Must(generateTemplates())
)

func generateTemplates() (*template.Template, error) {
	views := []string{
		"about",
		"footer",
		"generic-message",
		"head",
		"index",
		"login",
		"login-component",
		"new-thread",
		"register",
		"register-success",
		"thread",
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

	view := fmt.Sprintf("%s.html", viewName)
	if err := templates.ExecuteTemplate(res, view, data); err != nil {
		util.Check(err, "rendering %q view", view)
	}
}

func (h RequestHandler) ThreadRoute(res http.ResponseWriter, req *http.Request) {
	threadid, ok := util.GetURLPortion(req, 2)
	loggedIn, userid := h.IsLoggedIn(req)

	if !ok {
		title := "Thread not found"
		data := GenericMessageData{
			Title:   title,
			Message: "The thread does not exist (anymore?)",
		}
		h.renderView(res, "generic-message", TemplateData{Data: data, LoggedIn: loggedIn})
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
		http.Redirect(res, req, newSlug, http.StatusFound)
		return
	}
	// TODO (2022-01-07):
	// * handle error
	thread := h.db.GetThread(threadid)
	// markdownize content (but not title)
	for i, post := range thread {
		thread[i].Content = util.Markup(post.Content)
	}
	data := ThreadData{Posts: thread, ThreadURL: req.URL.Path}
	view := TemplateData{Data: &data, QuickNav: loggedIn, LoggedIn: loggedIn, LoggedInID: userid}
	if len(thread) > 0 {
		data.Title = thread[0].ThreadTitle
		view.Title = data.Title
	}
	h.renderView(res, "thread", view)
}

func (h RequestHandler) ErrorRoute(res http.ResponseWriter, req *http.Request, status int) {
	title := "Page not found"
	data := GenericMessageData{
		Title:   title,
		Message: fmt.Sprintf("The visited page does not exist (anymore?). Error code %d.", status),
	}
	h.renderView(res, "generic-message", TemplateData{Data: data, Title: title})
}

func (h RequestHandler) IndexRoute(res http.ResponseWriter, req *http.Request) {
	// handle 404
	if req.URL.Path != "/" {
		h.ErrorRoute(res, req, http.StatusNotFound)
		return
	}
	loggedIn, _ := h.IsLoggedIn(req)
	var mostRecentPost bool

	params := req.URL.Query()
	if q, exists := params["sort"]; exists {
		sortby := q[0]
		mostRecentPost = sortby == "posts"
	}
	// show index listing
	threads := h.db.ListThreads(mostRecentPost)
	view := TemplateData{Data: IndexData{threads}, LoggedIn: loggedIn, Title: "threads"}
	h.renderView(res, "index", view)
}

func IndexRedirect(res http.ResponseWriter, req *http.Request) {
	http.Redirect(res, req, "/", http.StatusSeeOther)
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
		h.renderView(res, "login", TemplateData{Data: LoginData{}, LoggedIn: loggedIn, Title: ""})
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
			h.renderView(res, "login", TemplateData{Data: LoginData{FailedAttempt: true}, LoggedIn: loggedIn, Title: ""})
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

func (h RequestHandler) RegisterRoute(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("register route")
	loggedIn, _ := h.IsLoggedIn(req)
	if loggedIn {
		data := GenericMessageData{
			Title:       "Register",
			Message:     "You already have an account (you are logged in with it).",
			Link:        "/",
			LinkMessage: "Visit the",
			LinkText:    "index",
		}
		h.renderView(res, "generic-message", TemplateData{Data: data, LoggedIn: loggedIn, Title: "register"})
		return
	}

	var verificationCode string
	renderErr := func(errFmt string, args ...interface{}) {
		errMessage := fmt.Sprintf(errFmt, args...)
		fmt.Println(errMessage)
		h.renderView(res, "register", TemplateData{Data: RegisterData{verificationCode, errMessage}})
	}

	var err error
	switch req.Method {
	case "GET":
		// try to get the verification code from the session (useful in case someone refreshed the page)
		verificationCode, err = h.session.GetVerificationCode(req)
		// we had an error getting the verification code, generate a code and set it on the session
		if err != nil {
			verificationCode = fmt.Sprintf("MRV%06d\n", crypto.GenerateVerificationCode())
			err = h.session.SaveVerificationCode(req, res, verificationCode)
			if err != nil {
				renderErr("Had troubles setting the verification code on session")
				return
			}
		}
		h.renderView(res, "register", TemplateData{Data: RegisterData{verificationCode, ""}})
	case "POST":
		verificationCode, err = h.session.GetVerificationCode(req)
		if err != nil {
			renderErr("There was no verification record for this browser session; missing data to compare against verification link content")
			return
		}
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		// read verification code from form
		verificationLink := req.PostFormValue("verificationlink")
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
		// generate and pass public keypair
		keypair, err := crypto.GenerateKeypair()
		// record generated pubkey in database for eventual later use
		err = h.db.AddPubkey(userID, keypair.Public)
		if err = ed.Eout(err, "insert pubkey in db"); err != nil {
			dump(err)
		}
		ed.Check(err, "generate keypair")
		kpJson, err := keypair.Marshal()
		ed.Check(err, "marshal keypair")
		h.renderView(res, "register-success", TemplateData{Data: RegisterSuccessData{string(kpJson)}, LoggedIn: loggedIn, Title: "registered successfully"})
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

func (h RequestHandler) GenericRoute(res http.ResponseWriter, req *http.Request) {
	data := GenericMessageData{
		Title:       "GenericTitle",
		Message:     "Generic message",
		Link:        "/",
		LinkMessage: "Generic link messsage",
		LinkText:    "with link",
	}
	h.renderView(res, "generic-message", TemplateData{Data: data})
}

func (h RequestHandler) AboutRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	h.renderView(res, "about", TemplateData{LoggedIn: loggedIn})
}

func (h RequestHandler) RobotsRoute(res http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(res, "User-agent: *\nDisallow: /")
}

func (h RequestHandler) NewThreadRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	switch req.Method {
	// Handle GET (=> want to start a new thread)
	case "GET":
		if !loggedIn {
			title := "Not logged in"
			data := GenericMessageData{
				Title:       title,
				Message:     "Only members of this forum may create new threads",
				Link:        "/login",
				LinkMessage: "If you are a member,",
				LinkText:    "log in",
			}
			h.renderView(res, "generic-message", TemplateData{Data: data, Title: title})
			return
		}
		h.renderView(res, "new-thread", TemplateData{LoggedIn: loggedIn, Title: "new thread"})
	case "POST":
		// Handle POST (=>
		title := req.PostFormValue("title")
		content := req.PostFormValue("content")
		// TODO (2022-01-10): unstub topicid, once we have other topics :)
		// the new thread was created: forward info to database
		threadid, err := h.db.CreateThread(title, content, userid, 1)
		if err != nil {
			data := GenericMessageData{
				Title:   "Error creating thread",
				Message: "There was a database error when creating the thread, apologies.",
			}
			h.renderView(res, "generic-message", TemplateData{Data: data, Title: "new thread"})
			return
		}
		// when data has been stored => redirect to thread
		slug := fmt.Sprintf("thread/%d/%s/", threadid, util.SanitizeURL(title))
		http.Redirect(res, req, "/"+slug, http.StatusSeeOther)
	default:
		fmt.Println("non get/post method, redirecting to index")
		IndexRedirect(res, req)
	}
}

func (h RequestHandler) DeletePostRoute(res http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		IndexRedirect(res, req)
		return
	}
	threadURL := req.PostFormValue("thread")
	postid, ok := util.GetURLPortion(req, 3)
	loggedIn, userid := h.IsLoggedIn(req)

	// generic error message base, with specifics being swapped out depending on the error
	genericErr := GenericMessageData{
		Title:       "Unaccepted request",
		LinkMessage: "Go back to",
		Link:        threadURL,
		LinkText:    "the thread",
	}

	renderErr := func(msg string) {
		fmt.Println(msg)
		genericErr.Message = msg
		h.renderView(res, "generic-message", TemplateData{Data: genericErr, LoggedIn: loggedIn})
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
	}
	http.Redirect(res, req, threadURL, http.StatusSeeOther)
}

func Serve(allowlist []string, sessionKey string, isdev bool) {
	port := ":8272"
	dbpath := "./data/forum.db"
	if isdev {
		developing = true
		dbpath = "./data/forum.test.db"
		port = ":8277"
	}

	db := database.InitDB(dbpath)
	handler := RequestHandler{&db, session.New(sessionKey, developing), allowlist}
	/* note: be careful with trailing slashes; go's default handler is a bit sensitive */
	// TODO (2022-01-10): introduce middleware to make sure there is never an issue with trailing slashes
	http.HandleFunc("/about", handler.AboutRoute)
	http.HandleFunc("/logout", handler.LogoutRoute)
	http.HandleFunc("/login", handler.LoginRoute)
	http.HandleFunc("/register", handler.RegisterRoute)
	http.HandleFunc("/post/delete/", handler.DeletePostRoute)
	http.HandleFunc("/thread/new/", handler.NewThreadRoute)
	http.HandleFunc("/thread/", handler.ThreadRoute)
	http.HandleFunc("/robots.txt", handler.RobotsRoute)
	http.HandleFunc("/", handler.IndexRoute)

	fileserver := http.FileServer(http.Dir("html/assets/"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fileserver))

	fmt.Println("Serving forum on", port)
	http.ListenAndServe(port, nil)
}
