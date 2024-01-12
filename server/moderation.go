package server

import (
	"fmt"
	"html/template"
	"strconv"
	"net/http"
	"time"

	"cerca/database"
	"cerca/crypto"
	"cerca/constants"
	"cerca/i18n"
	"cerca/util"
)

type AdminData struct {
	Admins []database.User
	Users []database.User
	Proposals []PendingProposal
	IsAdmin bool
}

type ModerationData struct {
	Log []string
}

type PendingProposal struct {
	// ID is the id of the proposal
	ID, ProposerID int
	Action string
	Time time.Time // the time self-confirmations become possible for proposers
	TimePassed bool // self-confirmations valid or not
}

func (h RequestHandler) displayErr(res http.ResponseWriter, req *http.Request, err error, title string) {
	errMsg := util.Eout(err, fmt.Sprintf("%s failed", title))
	fmt.Println(errMsg)
	data := GenericMessageData{
		Title:   title,
		Message: errMsg.Error(),
	}
	h.renderGenericMessage(res, req, data)
}

func (h RequestHandler) displaySuccess(res http.ResponseWriter, req *http.Request, title, message, backRoute string) {
	data := GenericMessageData{
		Title: title,
		Message: message,
		LinkText: h.translator.Translate("GoBack"),
		Link: backRoute,
	}
	h.renderGenericMessage(res, req, data)
}

// TODO (2023-12-10): any vulns with this approach? could a user forge a session cookie with the user id of an admin?
func (h RequestHandler) IsAdmin(req *http.Request) (bool, int) {
	ed := util.Describe("IsAdmin")
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
	// make sure the user id is actually an admin
	userIsAdmin, err := h.db.IsUserAdmin(userid)
	if err != nil {
		dump(ed.Eout(err, "IsUserAdmin in db"))
		return false, -1
	} else if !userIsAdmin {
		return false, -1
	}
	return true, userid
}

// there is a 2-quorum (requires 2 admins to take effect) imposed for the following actions, which are regarded as
// consequential:
// * make admin
// * remove account
// * demote admin

// note: there is only a 2-quorum constraint imposed if there are actually 2 admins. an admin may also confirm their own
// proposal if constants.PROPOSAL_SELF_CONFIRMATION_WAIT seconds have passed (1 week)
func performQuorumCheck (ed util.ErrorDescriber, db *database.DB, adminUserId, targetUserId, proposedAction int) error {
	// checks if a quorum is necessary for the proposed action: if a quorum constarin is in effect, a proposal is created
	// otherwise (if no quorum threshold has been achieved) the action is taken directly
	quorumActivated := db.QuorumActivated()

	var err error
	var modlogErr error
	if quorumActivated {
		err = db.ProposeModerationAction(adminUserId, targetUserId, proposedAction)
	} else {
		switch proposedAction {
		case constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER:
			err = db.RemoveUser(targetUserId)
			modlogErr = db.AddModerationLog(adminUserId, -1, constants.MODLOG_REMOVE_USER)
		case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
			err = db.AddAdmin(targetUserId)
			modlogErr = db.AddModerationLog(adminUserId, targetUserId, constants.MODLOG_ADMIN_MAKE)
		case constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN:
			err = db.DemoteAdmin(targetUserId)
			modlogErr = db.AddModerationLog(adminUserId, targetUserId, constants.MODLOG_ADMIN_DEMOTE)
		}
	}
	if modlogErr != nil {
		fmt.Println(ed.Eout(err, "error adding moderation log"))
	}
	if err != nil {
		return err
	}
	return nil
}

func (h *RequestHandler) AdminRemoveUser(res http.ResponseWriter, req *http.Request, targetUserId int) {
	ed := util.Describe("Admin remove user")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserId := h.IsAdmin(req)

	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	err := performQuorumCheck(ed, h.db, adminUserId, targetUserId, constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER)

	if err != nil {
		h.displayErr(res, req, err, "User removal")
		return
	}

	// success! redirect back to /admin
	http.Redirect(res, req, "/admin", http.StatusFound)
}

func (h *RequestHandler) AdminMakeUserAdmin(res http.ResponseWriter, req *http.Request, targetUserId int) {
	ed := util.Describe("make user admin")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserId := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	title := h.translator.Translate("AdminMakeAdmin")

	err := performQuorumCheck(ed, h.db, adminUserId, targetUserId, constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN)

	if err != nil {
		h.displayErr(res, req, err, title)
		return
	}

	if !h.db.QuorumActivated() {
		username, _ := h.db.GetUsername(targetUserId)
		message := fmt.Sprintf("User %s is now a fellow admin user!", username)
		h.displaySuccess(res, req, title, message, "/admin")
	} else {
		// redirect to admin view, which should have a proposal now
		http.Redirect(res, req, "/admin", http.StatusFound)
	}
}

func (h *RequestHandler) AdminDemoteAdmin(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("demote admin route")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserId := h.IsAdmin(req)

	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	title := h.translator.Translate("AdminDemote")

	useridString := req.PostFormValue("userid")
	targetUserId, err := strconv.Atoi(useridString)
	util.Check(err, "convert user id string to a plain userid")

	err = performQuorumCheck(ed, h.db, adminUserId, targetUserId, constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN)

	if err != nil {
		h.displayErr(res, req, err, title)
		return
	}

	if !h.db.QuorumActivated() {
		username, _ := h.db.GetUsername(targetUserId)
		message := fmt.Sprintf("User %s is now a regular user", username)
		// output copy-pastable credentials page for admin to send to the user
		h.displaySuccess(res, req, title, message, "/admin")
	} else {
		http.Redirect(res, req, "/admin", http.StatusFound)
	}
}

func (h *RequestHandler) AdminManualAddUserRoute(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("admin manually add user")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserId := h.IsAdmin(req)

	if  !isAdmin {
		IndexRedirect(res, req)
		return
	}

	type AddUser struct {
		ErrorMessage string
	}

	var data AddUser
	view := TemplateData{Title: h.translator.Translate("AdminAddNewUser"), Data: &data, HasRSS: false, IsAdmin: isAdmin, LoggedIn: loggedIn}

	if req.Method == "GET" {
		h.renderView(res, "admin-add-user", view)
		return
	}

	if req.Method == "POST" && isAdmin {
		username := req.PostFormValue("username")

		// do a lil quick checky check to see if we already have that username registered, 
		// and if we do re-render the page with an error
		existed, err := h.db.CheckUsernameExists(username)
		ed.Check(err, "check username exists")

		if existed {
			data.ErrorMessage = fmt.Sprintf("Username (%s) is already registered", username)
			h.renderView(res, "admin-add-user", view)
			return
		}

		// set up basic credentials
		newPassword := crypto.GeneratePassword()
		passwordHash, err := crypto.HashPassword(newPassword)
		ed.Check(err, "hash password")
		targetUserId, err := h.db.CreateUser(username, passwordHash)
		ed.Check(err, "create new user %s", username)

		err = h.db.AddModerationLog(adminUserId, targetUserId, constants.MODLOG_ADMIN_ADD_USER)
		if err != nil {
			fmt.Println(ed.Eout(err, "error adding moderation log"))
		}

		title := h.translator.Translate("AdminAddNewUser")
		message := fmt.Sprintf(h.translator.Translate("AdminPasswordSuccessInstructions"), template.HTMLEscapeString(username), newPassword)
		h.displaySuccess(res, req, title, message, "/add-user")
	}
}

func (h *RequestHandler) AdminResetUserPassword(res http.ResponseWriter, req *http.Request, targetUserId int) {
	ed := util.Describe("admin reset password")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserId := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	title := util.Capitalize(h.translator.Translate("PasswordReset"))
	newPassword, err := h.db.ResetPassword(targetUserId)

	if err != nil {
		h.displayErr(res, req, err, title)
		return
	}

	err = h.db.AddModerationLog(adminUserId, targetUserId, constants.MODLOG_RESETPW)
	if err != nil {
		fmt.Println(ed.Eout(err, "error adding moderation log"))
	}

	username, _ := h.db.GetUsername(targetUserId)

	message := fmt.Sprintf(h.translator.Translate("AdminPasswordSuccessInstructions"), template.HTMLEscapeString(username), newPassword)
	h.displaySuccess(res, req, title, message, "/admin")
}

func (h *RequestHandler) ConfirmProposal(res http.ResponseWriter, req *http.Request) {
	h.HandleProposal(res, req, constants.PROPOSAL_CONFIRM)
}

func (h *RequestHandler) VetoProposal(res http.ResponseWriter, req *http.Request) {
	h.HandleProposal(res, req, constants.PROPOSAL_VETO)
}

func (h *RequestHandler) HandleProposal(res http.ResponseWriter, req *http.Request, decision bool) {
	ed := util.Describe("handle proposal proposal")
	isAdmin, adminUserId := h.IsAdmin(req)

	if !isAdmin {
		IndexRedirect(res, req)
		return
	}

	if req.Method == "POST" {
		proposalidString := req.PostFormValue("proposalid")
		proposalid, err := strconv.Atoi(proposalidString)
		ed.Check(err, "convert proposalid")
		err = h.db.FinalizeProposedAction(proposalid, adminUserId, decision)
		if err != nil {
			ed.Eout(err, "finalizing the proposed action returned early with an error")
		}
		http.Redirect(res, req, "/admin", http.StatusFound)
		return
	}
	IndexRedirect(res, req)
}

// Note: this route by definition contains user generated content, so we escape all usernames with
// html.EscapeString(username)
func (h *RequestHandler) ModerationLogRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)
	// logs are sorted by time descending, from latest entry to oldest
	logs := h.db.GetModerationLogs()
	viewData := ModerationData{Log: make([]string, 0)}

	type translationData struct {	
		Time, ActingUsername, RecipientUsername string
		Action template.HTML
	}

	for _, entry := range logs {
		var tdata translationData
		var translationString string
		tdata.Time = entry.Time.Format("2006-01-02 15:04:05")
		tdata.ActingUsername = template.HTMLEscapeString(entry.ActingUsername)
		tdata.RecipientUsername = template.HTMLEscapeString(entry.RecipientUsername)
		switch entry.Action {
		case constants.MODLOG_RESETPW:
			translationString = "modlogResetPassword"
			if isAdmin {
				translationString += "Admin"
			}
		case constants.MODLOG_ADMIN_MAKE:
			translationString = "modlogMakeAdmin"
		case constants.MODLOG_REMOVE_USER:
			translationString = "modlogRemoveUser"
		case constants.MODLOG_ADMIN_ADD_USER:
			translationString = "modlogAddUser"
			if isAdmin {
				translationString += "Admin"
			}
		case constants.MODLOG_ADMIN_DEMOTE:
			translationString = "modlogDemoteAdmin"
		case constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN:
			translationString = "modlogProposalDemoteAdmin"
		case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
			translationString = "modlogProposalMakeAdmin"
		case constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER:
			translationString = "modlogProposalRemoveUser"
		}

		actionString := h.translator.TranslateWithData(translationString, i18n.TranslationData{Data: tdata})

		/* rendering of decision (confirm/veto) taken on a pending proposal */
		if entry.QuorumUsername != "" {
			// use the translated actionString to embed in the translated proposal decision (confirmation/veto)
			propdata := translationData{ActingUsername: template.HTMLEscapeString(entry.QuorumUsername), Action: template.HTML(actionString)}
			// if quorumDecision is true -> proposal was confirmed
			translationString = "modlogConfirm"
			if !entry.QuorumDecision {
				translationString = "modlogVeto"
			} 
			proposalString := h.translator.TranslateWithData(translationString, i18n.TranslationData{Data: propdata})
			viewData.Log = append(viewData.Log, proposalString)
			/* rendering of "X proposed: <Y>" */
		} else if entry.Action == constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN || 
			entry.Action == constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN ||
			entry.Action == constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER {
				propXforY := translationData{Time: tdata.Time, ActingUsername: tdata.ActingUsername, Action: template.HTML(actionString)}
				proposalString := h.translator.TranslateWithData("modlogXProposedY", i18n.TranslationData{Data: propXforY})
				viewData.Log = append(viewData.Log, proposalString)
		} else {
			viewData.Log = append(viewData.Log, actionString)
		}
	}
	view := TemplateData{Title: h.translator.Translate("ModerationLog"), IsAdmin: isAdmin, LoggedIn: loggedIn, Data: viewData}
	h.renderView(res, "moderation-log", view)
}

// used for rendering /admin's pending proposals
func (h *RequestHandler) AdminRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)

	if req.Method == "POST" && loggedIn && isAdmin {
		action := req.PostFormValue("admin-action")
		useridString := req.PostFormValue("userid")
		targetUserId, err := strconv.Atoi(useridString)
		util.Check(err, "convert user id string to a plain userid")

		switch action {
		case "reset-password":
			h.AdminResetUserPassword(res, req, targetUserId)
		case "make-admin":
			h.AdminMakeUserAdmin(res, req, targetUserId)
		case "remove-account":
			h.AdminRemoveUser(res, req, targetUserId)
		}
		return
	}

	if req.Method == "GET" {
		if !loggedIn || !isAdmin {
			// non-admin users get a different view
			h.ListAdmins(res, req)
			return
		}
		admins := h.db.GetAdmins()
		normalUsers := h.db.GetUsers(false) // do not include admins
		proposedActions := h.db.GetProposedActions()
		// massage pending proposals into something we can use in the rendered view
		pendingProposals := make([]PendingProposal, len(proposedActions))
		now := time.Now()
		for i, prop := range proposedActions {
			// escape all ugc
			prop.ActingUsername = template.HTMLEscapeString(prop.ActingUsername)
			prop.RecipientUsername = template.HTMLEscapeString(prop.RecipientUsername)
			// one week from when the proposal was made
			t := prop.Time.Add(constants.PROPOSAL_SELF_CONFIRMATION_WAIT)
			var str string
			switch prop.Action {
			case constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN:
				str = "modlogProposalDemoteAdmin"
			case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
				str = "modlogProposalMakeAdmin"
			case constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER:
				str = "modlogProposalRemoveUser"
			}

			proposalString := h.translator.TranslateWithData(str, i18n.TranslationData{Data: prop})
			pendingProposals[i] = PendingProposal{ID: prop.ProposalID, ProposerID: prop.ActingID, Action: proposalString, Time: t, TimePassed: now.After(t)}
		}
		data := AdminData{Admins: admins, Users: normalUsers, Proposals: pendingProposals}
		view := TemplateData{Title: h.translator.Translate("AdminForumAdministration"), Data: &data, HasRSS: false, LoggedIn: loggedIn, LoggedInID: userid}
		h.renderView(res, "admin", view)
	}
}

// view of /admin for non-admin users (contains less information)
func (h *RequestHandler) ListAdmins(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	admins := h.db.GetAdmins()
	data := AdminData{Admins: admins}
	view := TemplateData{Title: h.translator.Translate("AdminForumAdministration"), Data: &data, HasRSS: false, LoggedIn: loggedIn}
	h.renderView(res, "admins-list", view)
	return
}
