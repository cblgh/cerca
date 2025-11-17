package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"gomod.cblgh.org/cerca/crypto"
)

func runAuthKeyGenFunction() string {
	hashInput := []byte(crypto.GeneratePassword())
	h := sha256.New()
	h.Write(hashInput)
	return fmt.Sprintf("%x", h.Sum(nil))
}

func genauthkey() {
	authkeyFlags := flag.NewFlagSet("genauthkey", flag.ExitOnError)

	help := createHelpString("genauthkey", []string{
		`cerca genauthkey <no other args>`,
	})
	authkeyFlags.Usage = func() { usage(help, authkeyFlags) }
	keyHash := runAuthKeyGenFunction()
	inform(keyHash)
}
