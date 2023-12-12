package main

import (
	"cerca/database"
	"cerca/util"
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
	var keypairFlag bool
	var passwordFlag bool
	var username string
	var dbPath string
	flag.StringVar(&username, "username", "", "username whose credentials should be reset")
	flag.StringVar(&dbPath, "database", "./data/forum.db", "full path to the forum database; e.g. ./data/forum.db")
	flag.BoolVar(&keypairFlag, "keypair", false, "reset the keypair")
	flag.BoolVar(&passwordFlag, "password", false, "reset the password. if true generates a random new password")
	flag.Parse()

	usage := `usage
  admin-reset --database ./data/forum.db --username <username to reset> [--keypair, --password]
  admin-reset --help for more information

  examples:
  # only reset the keypair, leaving the password intact
  ./admin-reset --database ../../testdata/forum.db --username bambas --keypair   

  # reset password only 
  ./admin-reset --database ../../testdata/forum.db --username bambas --password

  # reset both password and keypair
  ./admin-reset --database ../../testdata/forum.db --username bambas --password --keypair   
  `

	if username == "" {
		complain(usage)
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)
	ed := util.Describe("admin reset")
	newPassword, err := db.ResetPassword(userid)
	// TODO (2023-12-12): log cmd actions just as admin web-actions are logged

	if err != nil {
		complain("reset password failed (%w)", err)
	}

	inform("successfully updated %s's password hash", username)
	inform("new temporary password %s", newPassword)
}
