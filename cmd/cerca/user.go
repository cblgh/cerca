package main

import (
	"cerca/constants"
	"cerca/crypto"
	"cerca/database"
	"cerca/util"
	"flag"
	"fmt"
	"os"
)

type UserInfo struct {
	ID                 int
	Username, Password string
}

func createUser(username, password string, db *database.DB) UserInfo {
	ed := util.Describe("admin reset")
	// make sure username is not registered already
	var err error
	if exists, err := db.CheckUsernameExists(username); err != nil {
		complain("Database had a problem when checking username")
	} else if exists {
		complain("Username %s appears to already exist, please pick another name", username)
	}
	var hash string
	if hash, err = crypto.HashPassword(password); err != nil {
		complain("Database had a problem when hashing password")
	}
	var userID int
	if userID, err = db.CreateUser(username, hash); err != nil {
		complain("Error in db when creating user")
	}
	// log where the registration is coming from, in the case of indirect invites && for curiosity
	err = db.AddRegistration(userID, "https://example.com/admin-add-user")
	if err = ed.Eout(err, "add registration"); err != nil {
		complain("Database had a problem saving user registration location")
	}
	return UserInfo{ID: userID, Username: username, Password: password}
}

func user() {
	var username string
	var forumDomain string
	var dbPath string

	userFlags := flag.NewFlagSet("adduser", flag.ExitOnError)
	userFlags.StringVar(&forumDomain, "url", "https://forum.merveilles.town", "root url to forum, referenced in output")
	userFlags.StringVar(&username, "username", "", "username who should be created")
	userFlags.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")

	help := createHelpString("adduser", []string{
		`cerca adduser -username "<new username>"`,
	})
	userFlags.Usage = func() { usage(help, userFlags) }
	userFlags.Parse(os.Args[2:])

	// if run without flags, print the help info
	if userFlags.NFlag() == 0 {
		userFlags.Usage()
		return
	}

	if username == "" {
		complain(help)
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)

	newPassword := crypto.GeneratePassword()
	userInfo := createUser(username, newPassword, &db)

	// log cmd actions just as admin web-actions are logged
	systemUserid := db.GetSystemUserID()
	err := db.AddModerationLog(systemUserid, userInfo.ID, constants.MODLOG_ADMIN_ADD_USER)
	if err != nil {
		complain("adding mod log for adding new user failed (%w)", err)
	}

	loginRoute := fmt.Sprintf("%s/login", forumDomain)
	resetRoute := fmt.Sprintf("%s/reset", forumDomain)

	inform("[user]\n%s", username)
	inform("[password]\n%s", newPassword)
	inform("Please login at %s\n", loginRoute)
	inform("After logging in, visit %s to reset your password", resetRoute)
	inform("Admin action has been logged to /moderations")
}
