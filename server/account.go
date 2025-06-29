package server

import (
	"fmt"
	"net/http"

	"github.com/cblgh/cerca/crypto"
	"github.com/cblgh/cerca/database"
)

func renderMsgAccountView(h *RequestHandler, res http.ResponseWriter, req *http.Request, caller, errInput string) {
	errMessage := fmt.Sprintf("%s: %s", caller, errInput)
	loggedIn, userid := h.IsLoggedIn(req)
	username, _ := h.db.GetUsername(userid)
	h.renderView(res, "account", TemplateData{Data: AccountData{ErrorMessage: errMessage, LoggedInUsername: username, DeleteAccountRoute: ACCOUNT_DELETE_ROUTE, ChangeUsernameRoute: ACCOUNT_CHANGE_USERNAME_ROUTE, ChangePasswordRoute: ACCOUNT_CHANGE_PASSWORD_ROUTE}, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: "Account"})
}

func (h *RequestHandler) AccountChangePassword(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	sectionTitle := "Change password"
	renderErr := func(errMsg string) {
		renderMsgAccountView(h, res, req, sectionTitle, errMsg)
	}
	// simple alias for the same thing, to make it less confusing in the different cases :) might be changed into some
	// other success behaviour at some future point
	renderSuccess := renderErr
	if req.Method == "GET" {
		if !loggedIn {
			IndexRedirect(res, req)
			return
		}
		http.Redirect(res, req, "/account", http.StatusSeeOther)
		return
	} else if req.Method == "POST" {
		// verify existing credentials
		currentPassword := req.PostFormValue("current-password")
		err := h.checkPasswordIsCorrect(userid, currentPassword)
		if err != nil {
			renderErr("Current password did not match up with the hash stored in database")
			return
		}

		newPassword := req.PostFormValue("new-password")
		newPasswordCopy := req.PostFormValue("new-password-copy")

		// too short
		if len(newPassword) < 9 {
			renderErr("New password is too short (needs to be at least 9 characters or longer)")
			return
		}
		// repeat password did not match
		if newPassword != newPasswordCopy {
			renderErr("New password was incorrectly repeated")
			return
		}
		// happy path
		if newPassword == newPasswordCopy {
			passwordHash, err := crypto.HashPassword(newPassword)
			if err != nil {
				renderErr("Critical failure - password hashing failed. Contact admin")
			}
			h.db.UpdateUserPasswordHash(userid, passwordHash)
			renderSuccess("Password has been updated!")
		}
	}
}
func (h *RequestHandler) AccountChangeUsername(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	sectionTitle := "Change username"
	renderErr := func(errMsg string) {
		renderMsgAccountView(h, res, req, sectionTitle, errMsg)
	}
	renderSuccess := renderErr
	if req.Method == "GET" {
		if !loggedIn {
			IndexRedirect(res, req)
			return
		}
		http.Redirect(res, req, "/account", http.StatusSeeOther)
		return
	} else if req.Method == "POST" {
		// verify existing credentials
		currentPassword := req.PostFormValue("current-password")
		err := h.checkPasswordIsCorrect(userid, currentPassword)
		if err != nil {
			renderErr("Current password did not match up with the hash stored in database")
			return
		}
		newUsername := req.PostFormValue("new-username")
		var exists bool
		if exists, err = h.db.CheckUsernameExists(newUsername); err != nil {
			renderErr("Database had a problem when checking username")
			return
		} else if exists {
			renderErr(fmt.Sprintf("Username %s appears to already exist, please pick another name", newUsername))
			return
		}
		h.db.UpdateUsername(userid, newUsername)
		renderSuccess(fmt.Sprintf("You are now known as %s", newUsername))
		// TODO (2024-04-09): add modlog entry so that other forum users (or only admins?) can follow along with changing nicknames
	}
}

func (h *RequestHandler) AccountSelfServiceDelete(res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	sectionTitle := "Delete account"
	renderErr := func(errMsg string) {
		renderMsgAccountView(h, res, req, sectionTitle, errMsg)
	}
	if req.Method == "GET" {
		if !loggedIn {
			IndexRedirect(res, req)
			return
		}
		http.Redirect(res, req, "/account", http.StatusSeeOther)
		return
	} else if req.Method == "POST" {
		fmt.Printf("%s route hit with POST\n", ACCOUNT_DELETE_ROUTE)
		if !loggedIn {
			renderErr("You are, somehow, not logged in. Please refresh your browser, try to logout, and then log back in again.")
			return
		}
		// verify existing credentials
		currentPassword := req.PostFormValue("current-password")
		err := h.checkPasswordIsCorrect(userid, currentPassword)
		if err != nil {
			renderErr("Current password did not match up with the hash stored in database")
			return
		}

		/* since deletion is such a permanent action, we take some precautions with the following code and choose a verbose
		* but redundant approach to confirming the correctness of the received input */
		deleteConfirmationCheckbox := req.PostFormValue("delete-confirm")
		if deleteConfirmationCheckbox != "on" {
			renderErr("The delete account confirmation checkbox was, somehow, not ticked.")
			return
		}

		delErrMsg := "[DERR%d] The delete account functionality hit an error, please ping the forum maintainer with this message and error code!"
		var deleteOpts database.RemoveUserOptions
		// contains values from a radio button
		deleteDecision := req.PostFormValue("delete-post-decision")
		// delete-everything is a checkbox: check if it was checked
		wantDeleteEverything := req.PostFormValue("delete-everything") == "on"
		// boolean to make sure that a delete option was accurately through either the delete-everything checkbox
		// or the granular options represented by radio buttons.
		// this check ensures that `deleteOpts` was actually set and doesn't just contain default values
		deleteIsConfigured := false

		// if delete everything and a granular option is chosen, error out instead
		if (len(deleteDecision) > 0 && deleteDecision != "no-choice") && wantDeleteEverything {
			renderErr("Choose either delete everything, or one of the more granular options; not both")
			return
		}
		// no option was chosen
		if (deleteDecision == "no-choice" || deleteDecision == "") && !wantDeleteEverything {
			renderErr("You did not choose a delete option; please try again and choose one of the account delete options.")
			return
		}

		if wantDeleteEverything {
			deleteOpts = database.RemoveUserOptions{KeepContent: false, KeepUsername: false}
			deleteIsConfigured = true
		}

		switch deleteDecision {
		case "posts-intact-username-intact":
			// <b>Keep</b> post contents and <b>keep</b> username attribution
			deleteOpts = database.RemoveUserOptions{KeepContent: true, KeepUsername: true}
			deleteIsConfigured = true
		case "posts-intact-username-removed":
			// <b>Keep</b> post contents but <b>remove</b> username from posts
			deleteOpts = database.RemoveUserOptions{KeepContent: true, KeepUsername: false}
			deleteIsConfigured = true
		case "posts-removed-username-intact":
			// <b>Remove</b> post contents and <b>keep</b> my username
			deleteOpts = database.RemoveUserOptions{KeepContent: false, KeepUsername: true}
			deleteIsConfigured = true
		case "no-choice":
			break
		default:
			renderErr(fmt.Sprintf(delErrMsg, 1))
			fmt.Println("hit default for deleteDecision - not doing anything! this isn't good!!")
			return
		}

		if !deleteIsConfigured {
			renderErr(fmt.Sprintf(delErrMsg, 2))
			fmt.Println("delete was not configured! this isn't good!!")
			return
		}

		// all our checks have passed and it looks like we're in for some deleting!
		fmt.Println("deleting user with userid", userid)
		err = h.db.RemoveUser(userid, deleteOpts)
		if err != nil {
			renderErr(fmt.Sprintf(delErrMsg, 3))
			return
		}
		// log the user out
		http.Redirect(res, req, "/logout", http.StatusSeeOther)
	}
}
