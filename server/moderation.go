package server

import (
	"fmt"
	"html/template"
	"strconv"
	"net/http"
	"time"

	"cerca/crypto"
	"cerca/constants"
	"cerca/i18n"
	"cerca/util"
)

type AdminsData struct {
	Admins []database.User
	Users []database.User
	Proposals []PendingProposal
	IsAdmin bool
}

type ModerationData struct {
	Log []string
}

type PendingProposal struct {
	ID, ProposerID int
	Action string
	Time time.Time // the time self-confirmations become possible for proposers
	TimePassed bool // self-confirmations valid or not
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

func (h *RequestHandler) AdminRemoveUser(res http.ResponseWriter, req *http.Request, targetUserid int) {
	ed := util.Describe("Admin remove user")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserid := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	quorumActivated := h.db.QuorumActivated()
	var err error
	if quorumActivated {
		err = h.db.ProposeModerationAction(adminUserid, targetUserid, constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER)
	} else {
		err = h.db.RemoveUser(targetUserid)
	}

	if err != nil {
		// TODO (2023-12-09): bubble up error to visible page as feedback for admin
		errMsg := ed.Eout(err, "remove user failed")
		fmt.Println(errMsg)
		data := GenericMessageData{
			Title:   "User removal",
			Message: errMsg.Error(),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	if !quorumActivated {
		err = h.db.AddModerationLog(adminUserid, -1, constants.MODLOG_REMOVE_USER)
		if err != nil {
			fmt.Println(ed.Eout(err, "error adding moderation log"))
		}
	}
	// success! redirect back to /admin
	http.Redirect(res, req, "/admin", http.StatusFound)
}

func (h *RequestHandler) AdminMakeUserAdmin(res http.ResponseWriter, req *http.Request, targetUserid int) {
	ed := util.Describe("make user admin")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserid := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	quorumActivated := h.db.QuorumActivated()
	var err error
	if quorumActivated {
		err = h.db.ProposeModerationAction(adminUserid, targetUserid, constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN)
	} else {
		err = h.db.AddAdmin(targetUserid)
	}

	if err != nil {
		// TODO (2023-12-09): bubble up error to visible page as feedback for admin
		errMsg := ed.Eout(err, "make admin failed")
		fmt.Println(errMsg)
		data := GenericMessageData{
			Title:   "Make admin",
			Message: errMsg.Error(),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	if !quorumActivated {
		username, _ := h.db.GetUsername(targetUserid)
		err = h.db.AddModerationLog(adminUserid, targetUserid, constants.MODLOG_ADMIN_MAKE)
		if err != nil {
			fmt.Println(ed.Eout(err, "error adding moderation log"))
		}

		// output copy-pastable credentials page for admin to send to the user
		data := GenericMessageData{
			Title: "Make admin success",
			Message: fmt.Sprintf("User %s is now a fellow admin user!", username),
			LinkMessage: "Go back to the",
			LinkText: "admin view",
			Link: "/admin",
		}
		h.renderGenericMessage(res, req, data)
	} else {
		// redirect to admin view, which should have a proposal now
		http.Redirect(res, req, "/admin", http.StatusFound)
	}
}

func (h *RequestHandler) AdminDemoteAdmin(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("demote admin route")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserid := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}
	useridString := req.PostFormValue("userid")
	targetUserid, err := strconv.Atoi(useridString)
	util.Check(err, "convert user id string to a plain userid")

	quorumActivated := h.db.QuorumActivated()
	if quorumActivated {
		err = h.db.ProposeModerationAction(adminUserid, targetUserid, constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN)
	} else {
		err = h.db.DemoteAdmin(targetUserid)
	}

	if err != nil {
		errMsg := ed.Eout(err, "demote admin failed")
		fmt.Println(errMsg)
		data := GenericMessageData{
			Title:   "Demote admin",
			Message: errMsg.Error(),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	if !quorumActivated {
		username, _ := h.db.GetUsername(targetUserid)
		err = h.db.AddModerationLog(adminUserid, targetUserid, constants.MODLOG_ADMIN_DEMOTE)
		if err != nil {
			fmt.Println(ed.Eout(err, "error adding moderation log"))
		}

		// output copy-pastable credentials page for admin to send to the user
		data := GenericMessageData{
			Title: "Demote admin success",
			Message: fmt.Sprintf("User %s is now a regular user", username),
			LinkMessage: "Go back to the",
			LinkText: "admin view",
			Link: "/admin",
		}
		h.renderGenericMessage(res, req, data)
	} else {
		http.Redirect(res, req, "/admin", http.StatusFound)
	}
}

func (h *RequestHandler) AdminManualAddUserRoute(res http.ResponseWriter, req *http.Request) {
	ed := util.Describe("admin manually add user")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserid := h.IsAdmin(req)

	if  !isAdmin {
		IndexRedirect(res, req)
		return
	}

	type AddUser struct {
		ErrorMessage string
	}

	var data AddUser
	view := TemplateData{Title: "Add a new user", Data: &data, HasRSS: false, IsAdmin: isAdmin, LoggedIn: loggedIn}

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
		targetUserid, err := h.db.CreateUser(username, passwordHash)
		ed.Check(err, "create new user %s", username)

		// if err != nil {
		// 	// TODO (2023-12-09): bubble up error to visible page as feedback for admin
		// 	errMsg := ed.Eout(err, "reset password failed")
		// 	fmt.Println(errMsg)
		// 	data := GenericMessageData{
		// 		Title:   "Admin reset password",
		// 		Message: errMsg.Error(),
		// 	}
		// 	h.renderGenericMessage(res, req, data)
		// 	return
		// }

		err = h.db.AddModerationLog(adminUserid, targetUserid, constants.MODLOG_ADMIN_ADD_USER)
		if err != nil {
			fmt.Println(ed.Eout(err, "error adding moderation log"))
		}

		// output copy-pastable credentials page for admin to send to the user
		data := GenericMessageData{
			Title: "User successfully added",
			Message: fmt.Sprintf("Instructions: %s's password was set to: %s. After logging in, please change your password by going to /reset", username, newPassword),
			LinkMessage: "Go back to the",
			LinkText: "add user view",
			Link: "/add-user",
		}
		h.renderGenericMessage(res, req, data)
	}
}

func (h *RequestHandler) AdminResetUserPassword(res http.ResponseWriter, req *http.Request, targetUserid int) {
	ed := util.Describe("admin reset password")
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, adminUserid := h.IsAdmin(req)
	if req.Method == "GET" || !loggedIn || !isAdmin {
		IndexRedirect(res, req)
		return
	}

	newPassword, err := h.db.ResetPassword(targetUserid)

	if err != nil {
		// TODO (2023-12-09): bubble up error to visible page as feedback for admin
		errMsg := ed.Eout(err, "reset password failed")
		fmt.Println(errMsg)
		data := GenericMessageData{
			Title:   "Admin reset password",
			Message: errMsg.Error(),
		}
		h.renderGenericMessage(res, req, data)
		return
	}

	err = h.db.AddModerationLog(adminUserid, targetUserid, constants.MODLOG_RESETPW)
	if err != nil {
		fmt.Println(ed.Eout(err, "error adding moderation log"))
	}

	username, _ := h.db.GetUsername(targetUserid)

	// output copy-pastable credentials page for admin to send to the user
	data := GenericMessageData{
		Title: "Password reset successful!",
		Message: fmt.Sprintf("Instructions: %s's password was reset to: %s. After logging in, please change your password by going to /reset", username, newPassword),
		LinkMessage: "Go back to the",
		LinkText: "admin view",
		Link: "/admin",
	}
	h.renderGenericMessage(res, req, data)
}

func (h *RequestHandler) HandleProposal(res http.ResponseWriter, req *http.Request, decision bool) {
	ed := util.Describe("handle proposal proposal")
	isAdmin, adminUserid := h.IsAdmin(req)

	if !isAdmin {
		IndexRedirect(res, req)
		return
	}

	if req.Method == "POST" && isAdmin {
		proposalidString := req.PostFormValue("proposalid")
		proposalid, err := strconv.Atoi(proposalidString)
		ed.Check(err, "convert proposalid")
		err = h.db.FinalizeProposedAction(proposalid, adminUserid, decision)
		ed.Check(err, "finalize proposal error")
		http.Redirect(res, req, "/admin", http.StatusFound)
		return
	}
	IndexRedirect(res, req)
}

func (h *RequestHandler) ConfirmProposal(res http.ResponseWriter, req *http.Request) {
	h.HandleProposal(res, req, constants.PROPOSAL_CONFIRM)
}

func (h *RequestHandler) VetoProposal(res http.ResponseWriter, req *http.Request) {
	h.HandleProposal(res, req, constants.PROPOSAL_VETO)
}

// Note: this will by definition contain ugc, so we need to escape all usernames with html.EscapeString(username) before
// populating ModerationLogEntry
/* sorted by time descending, from latest entry to oldest */

func (h *RequestHandler) ModerationLogRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)
	logs := h.db.GetModerationLogs()
	viewData := ModerationData{Log: make([]string, 0)}

	type translationData struct {	
		Time, ActingUsername, RecipientUsername string
		Action template.HTML
	}
	type proposalData struct {	
		QuorumUsername string
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
			propdata := proposalData{QuorumUsername: template.HTMLEscapeString(entry.QuorumUsername), Action: template.HTML(actionString)}
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
	view := TemplateData{Title: "Moderation log", IsAdmin: isAdmin, LoggedIn: loggedIn, Data: viewData}
	h.renderView(res, "moderation-log", view)
}
// used for rendering /admin's pending proposals
// TODO (2023-12-10): there is a 2-quorum (requires 2 admins to take effect) imposed for the following actions, which
// are regarded as consequential:
// * make admin
// * remove account
// * demote admin

// note: there is only a 2-quorum constraint if there are actually 2 admins. an admin may also confirm their own
// proposal if constants.PROPOSAL_SELF_CONFIRMATION_WAIT seconds have passed (1 week)

func (h *RequestHandler) AdminRoute(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	isAdmin, _ := h.IsAdmin(req)

	if req.Method == "POST" && loggedIn && isAdmin {
		action := req.PostFormValue("admin-action")
		useridString := req.PostFormValue("userid")
		targetUserid, err := strconv.Atoi(useridString)
		util.Check(err, "convert user id string to a plain userid")

		switch action {
		case "reset-password":
			h.AdminResetUserPassword(res, req, targetUserid)
		case "make-admin":
			h.AdminMakeUserAdmin(res, req, targetUserid)
		case "remove-account":
			h.AdminRemoveUser(res, req, targetUserid)
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
		data := AdminsData{Admins: admins, Users: normalUsers, Proposals: pendingProposals}
		view := TemplateData{Title: "Forum Administration", Data: &data, HasRSS: false, LoggedIn: loggedIn, LoggedInID: userid}
		h.renderView(res, "admin", view)
	}
}

func (h *RequestHandler) ListAdmins(res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	admins := h.db.GetAdmins()
	data := AdminsData{Admins: admins}
	view := TemplateData{Title: "Forum Administrators", Data: &data, HasRSS: false, LoggedIn: loggedIn}
	h.renderView(res, "admins-list", view)
	return
}
