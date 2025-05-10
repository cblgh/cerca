package main

import (
	"cerca/crypto"
	"crypto/sha256"
	"flag"
)

func authkey() {
	authkeyFlags := flag.NewFlagSet("authkey", flag.ExitOnError)

	help := createHelpString("authkey", []string{
		`cerca authkey <no other args>`,
	})
	authkeyFlags.Usage = func() { usage(help, authkeyFlags) }

	hashInput := []byte(crypto.GeneratePassword())
	h := sha256.New()
	h.Write(hashInput)
	inform("Generated a random key:")
	inform("--authkey %x", h.Sum(nil))
}
