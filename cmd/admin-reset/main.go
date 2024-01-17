package main

import (
	"cerca/database"
	"cerca/constants"
	"flag"
	"fmt"
	"os"
)

func inform(msg string, args ...interface{}) {
	if len(args) > 0 {
		fmt.Printf("admin-reset: %s\n", fmt.Sprintf(msg, args...))
	} else {
		fmt.Printf("admin-reset: %s\n", msg)
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
	var dbPath string
	flag.StringVar(&username, "username", "", "username whose credentials should be reset")
	flag.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")
	flag.Parse()

	usage := `usage
  admin-reset --database ./data/forum.db --username <username to reset>
  admin-reset --help for more information

  # example
  ./admin-reset --database ../../testdata/forum.db --username bambas 
  `

	if username == "" {
		complain(usage)
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)

	userid, err := db.GetUserID(username)
	if err != nil {
		complain("reset password failed (%w)", err)
	}
	newPassword, err := db.ResetPassword(userid)

	if err != nil {
		complain("reset password failed (%w)", err)
	}

	// log cmd actions just as admin web-actions are logged
	systemUserid := db.GetSystemUserid()
	err = db.AddModerationLog(systemUserid, userid, constants.MODLOG_RESETPW)
	if err != nil {
		complain("adding mod log for password reset failed (%w)", err)
	}

	inform("Successfully updated %s's password hash", username)
	inform("New temporary password: %s", newPassword)
	inform("Admin action has been logged to /moderations")
}
