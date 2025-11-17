package server

import (
	"errors"

	"gomod.cblgh.org/cerca/crypto"
	"gomod.cblgh.org/cerca/util"
)

func (h *RequestHandler) checkPasswordIsCorrect(userid int, password string) error {
	ed := util.Describe("checkPasswordIsCorrect")
	// * hash received password and compare to stored hash
	passwordHash, err := h.db.GetPasswordHashByUserID(userid)
	if err = ed.Eout(err, "getting password hash and uid"); err == nil && !crypto.ValidatePasswordHash(password, passwordHash) {
		return errors.New("hashing the supplied password did not result in what was stored in the database")
	}
	if err != nil {
		return errors.New("password check failed")
	}
	return nil
}
