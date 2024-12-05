package crypto

import (
	"cerca/util"
	"crypto/rand"
	"github.com/matthewhartstonge/argon2"
	"math/big"
	"strings"
)

func HashPassword(s string) (string, error) {
	ed := util.Describe("hash password")
	config := argon2.MemoryConstrainedDefaults()
	hash, err := config.HashEncoded([]byte(s))
	if err != nil {
		return "", ed.Eout(err, "hashing with argon2id")
	}
	return string(hash), nil
}

func ValidatePasswordHash(password, passwordHash string) bool {
	ed := util.Describe("validate password hash")
	hashStruct, err := argon2.Decode([]byte(passwordHash))
	ed.Check(err, "argon2.decode")
	correct, err := hashStruct.Verify([]byte(password))
	if err != nil {
		return false
	}
	return correct
}

// used for generating a random reset password
const characterSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const pwlength = 20

func GeneratePassword() string {
	var password strings.Builder
	const maxChar = int64(len(characterSet))

	for i := 0; i < pwlength; i++ {
		max := big.NewInt(maxChar)
		bigN, err := rand.Int(rand.Reader, max)
		util.Check(err, "randomly generate int")
		n := bigN.Int64()
		password.WriteString(string(characterSet[n]))
	}
	return password.String()
}
