package server

import (
	"fmt"
	"errors"
	"net/http"

	"cerca/crypto"
	"cerca/util"
)

func renderMsgAccountView (h *RequestHandler, res http.ResponseWriter, req *http.Request, caller, errInput string) {
	errMessage := fmt.Sprintf("%s: %s", caller, errInput)
	loggedIn, userid := h.IsLoggedIn(req)
	username, _ := h.db.GetUsername(userid)
	h.renderView(res, "account", TemplateData{Data: AccountData{ErrorMessage: errMessage, LoggedInUsername: username, DeleteAccountRoute: ACCOUNT_DELETE_ROUTE, ChangeUsernameRoute: ACCOUNT_CHANGE_USERNAME_ROUTE, ChangePasswordRoute: ACCOUNT_CHANGE_PASSWORD_ROUTE}, HasRSS: h.config.RSS.URL != "", LoggedIn: loggedIn, Title: "Account"})
}

func (h *RequestHandler) checkPasswordIsCorrect(userid int, password string) error  {
	ed := util.Describe("checkPasswordIsCorrect")
	username, err := h.db.GetUsername(userid)
	if err != nil {
		return errors.New("could not get the username for the logged-in user")
	}
	// * hash received password and compare to stored hash
	passwordHash, _, err := h.db.GetPasswordHash(username)
	if err = ed.Eout(err, "getting password hash and uid"); err == nil && !crypto.ValidatePasswordHash(password, passwordHash) {
		return errors.New("incorrect current password")
	}
	if err != nil {
		return errors.New("password check failed")
	}
	return nil
}

func (h *RequestHandler) AccountChangePassword (res http.ResponseWriter, req *http.Request) {
	loggedIn, userid := h.IsLoggedIn(req)
	sectionTitle := "Change password"
	renderErr := func(errMsg string) {
		renderMsgAccountView(h, res, req, sectionTitle, errMsg)
	}
	renderSuccess := renderErr
	if req.Method == "GET" || !loggedIn {
		IndexRedirect(res, req)
		return
	} else if req.Method == "POST" {
		// verify 
		currentPassword := req.PostFormValue("current-password")
		err := h.checkPasswordIsCorrect(userid, currentPassword)
		if err != nil {
			renderErr(err.Error())
			return
		}
		newPassword := req.PostFormValue("new-password")
		newPasswordCopy := req.PostFormValue("new-password-copy")
		if len(newPassword) < 9 {
			renderErr("New password is too short (needs to be at least 9 characters or longer)")
			return
		}
		if newPassword != newPasswordCopy {
			renderErr("New password was incorrectly repeated")
			return
		}
		if newPassword == newPasswordCopy {
			passwordHash, err := crypto.HashPassword(newPassword)
			if err != nil {
				renderErr("Critical failure - password hashing failed. Contact admin")
			}
			h.db.UpdateUserPasswordHash(userid, passwordHash)
			renderSuccess("Password has been updated!")
			return
		}
	}
	fmt.Println("change password!")
}
func (h *RequestHandler) AccountChangeUsername (res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	if req.Method == "GET" || !loggedIn {
		IndexRedirect(res, req)
		return
	}
	fmt.Println("change username!")
}
func (h *RequestHandler) AccountSelfServiceDelete (res http.ResponseWriter, req *http.Request) {
	loggedIn, _ := h.IsLoggedIn(req)
	if req.Method == "GET" || !loggedIn {
		IndexRedirect(res, req)
		return
	}
	fmt.Println("delete!")
}
