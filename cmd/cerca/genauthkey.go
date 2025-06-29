package main

import (
	"cerca/crypto"
	"crypto/sha256"
	"flag"
)

func genauthkey() {
	authkeyFlags := flag.NewFlagSet("genauthkey", flag.ExitOnError)

	help := createHelpString("genauthkey", []string{
		`cerca genauthkey <no other args>`,
	})
	authkeyFlags.Usage = func() { usage(help, authkeyFlags) }

	hashInput := []byte(crypto.GeneratePassword())
	h := sha256.New()
	h.Write(hashInput)
	inform("Generated a random key:")
	inform("--authkey %x", h.Sum(nil))
}
