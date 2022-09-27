package main

import (
	"cerca/crypto"
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
	if !keypairFlag && !passwordFlag {
		complain("nothing to reset, exiting")
	}

	// check if database exists! we dont wanna create a new db in this case ':)
	if !database.CheckExists(dbPath) {
		complain("couldn't find database at %s", dbPath)
	}

	db := database.InitDB(dbPath)
	ed := util.Describe("admin reset")

	userid, err := db.GetUserID(username)
	if err != nil {
		complain("username %s not in database", username)
	}

	// generate new password for user and set it in the database
	if passwordFlag {
		newPassword := crypto.GeneratePassword()
		passwordHash, err := crypto.HashPassword(newPassword)
		ed.Check(err, "hash new password")
		db.UpdateUserPasswordHash(userid, passwordHash)

		inform("successfully updated %s's password hash", username)
		inform("new temporary password %s", newPassword)
	}

	// generate a new keypair for user and update user's pubkey record with new pubkey
	if keypairFlag {
		kp, err := crypto.GenerateKeypair()
		ed.Check(err, "generate keypair")
		kpBytes, err := kp.Marshal()
		ed.Check(err, "marshal keypair")
		pubkey, err := kp.PublicString()
		ed.Check(err, "get pubkey string")
		err = db.SetPubkey(userid, pubkey)
		ed.Check(err, "set new pubkey in database")

		inform("successfully changed %s's stored public key", username)
		inform("new keypair\n%s", string(kpBytes))
	}
}
