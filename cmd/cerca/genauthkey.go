package main

import (
	"cerca/crypto"
	"crypto/sha256"
	"fmt"
	"flag"
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
