package server

import (
	"fmt"
	"net/http"

	"cerca/crypto"
)

func renderMsgAccountView (h *RequestHandler, res http.ResponseWriter, req *http.Request, caller, errInput string) {
	errMessage := fmt.Sprintf("%s: %s", caller, errInput)
	loggedIn, userid := h.IsLoggedIn(req)
	username, _ := h.db.GetUsername(userid)
	h.renderView(res, "account", TemplateData{Data: AccountData{ErrorMessage: errMessage, LoggedInUsername: username, DeleteAccountRoute: ACCOUNT_DELETE_ROUTE, ChangeUsernameRoute: ACCOUNT_CHANGE_USERNAME_ROUTE, ChangePasswordRoute: ACCOUNT_CHANGE_PASSWORD_ROUTE}, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: "Account"})
}

func (h *RequestHandler) AccountChangePassword (res http.ResponseWriter, req *http.Request) {
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
func (h *RequestHandler) AccountChangeUsername (res http.ResponseWriter, req *http.Request) {
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
		fmt.Println("new username", newUsername)
		var exists bool
		if exists, err = h.db.CheckUsernameExists(newUsername); err != nil {
			renderErr("Database had a problem when checking username")
			return
		} else if exists {
			renderErr(fmt.Sprintf("Username %s appears to already exist, please pick another name", newUsername))
			return
		}
		renderSuccess(fmt.Sprintf("You are now known as %s", newUsername))
		h.db.UpdateUsername(userid, newUsername)
		// TODO (2024-04-09): add modlog entry so that other forum users (or only admins?) can follow along with changing nicknames
	}
}

func (h *RequestHandler) AccountSelfServiceDelete (res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	sectionTitle := "Delete account"
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
	}
	fmt.Println("delete!")
}
