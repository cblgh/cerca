package main_test

import (
	"cerca/crypto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMain(t *testing.T) {
	a := assert.New(t)

	kp, err := crypto.GenerateKeypair()
	a.NoError(err)
	msg := []byte("hi")
	proof := crypto.CreateProof(kp, msg)
	a.NotZero(len(proof), "proof length greater than zero")
	proofVerificationCorrect := crypto.VerifyProof(kp.Public, msg, proof)
	a.True(proofVerificationCorrect)
}
