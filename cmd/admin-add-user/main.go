package main

import (
	"cerca/crypto"
	"cerca/database"
	"cerca/constants"
	"cerca/util"
	"flag"
	"fmt"
	"os"
)

type UserInfo struct {
	ID int
	Username, Password string
}

func createUser (username, password string, db *database.DB) UserInfo {
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

func inform(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("%s\n", fmt.Sprintf(msg, args...))
	} else {
		fmt.Printf("%s\n", msg)
	}
}

func complain(msg string, args ...interface{}) {
	if len(args) > 0 {
		inform(msg, args)
	} else {
		inform(msg)
	}
	os.Exit(0)
}

func main() {
	var username string
	var forumDomain string
	var dbPath string
	flag.StringVar(&forumDomain, "url", "https://forum.merveilles.town", "root url to forum, referenced in output")
	flag.StringVar(&username, "username", "", "username whose credentials should be reset")
	flag.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")
	flag.Parse()

	usage := `usage
	admin-add-user --url https://forum.merveilles.town --database ./data/forum.db --username <username to create account for> 
	admin-add-user --help for more information
  `

	if username == "" {
		complain(usage)
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)

	newPassword := crypto.GeneratePassword()
	userInfo := createUser(username, newPassword, &db)

	// log cmd actions just as admin web-actions are logged
	systemUserid := db.GetSystemUserid()
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
