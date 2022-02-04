package main

import (
	"cerca/crypto"
	"flag"
	"fmt"
	"os"
)

func main() {
	var kpPath string
	var payload string
	flag.StringVar(&kpPath, "keypair", "", "path to your account-securing public-keypair (the thing you got during registering and was told to save:)")
	flag.StringVar(&payload, "payload", "", "the payload presented on the restore password page")
	flag.Parse()

	if kpPath == "" {
		fmt.Println(`usage:
  tool --keypair <path-to-keypair.json> --payload <payload from website>
  tool --help for more information`)
		os.Exit(0)
	}
	kp, _ := crypto.ReadKeypair(kpPath)
	proof := crypto.CreateProof(kp, []byte(payload))
	fmt.Println("your proof:")
	fmt.Println(fmt.Sprintf("%x", proof))
	fmt.Println("\nplease paste the proof in the proof box of the restore password page, thank you!")
}
