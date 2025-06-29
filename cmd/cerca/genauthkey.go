package main

import (
	"github.com/cblgh/cerca/crypto"
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
	inform("%x", h.Sum(nil))
}
