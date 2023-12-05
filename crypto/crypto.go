package crypto

import (
	"cerca/util"
	crand "crypto/rand"
	"encoding/binary"
	"github.com/synacor/argon2id"
	"math/big"
	rand "math/rand"
	"strings"
)

func HashPassword(s string) (string, error) {
	ed := util.Describe("hash password")
	// TODO (2023-12-05): see about argon2id replacement (or getting 32 bit bug fixed)
	hash, err := argon2id.DefaultHashPassword(s)
	if err != nil {
		return "", ed.Eout(err, "hashing with argon2id")
	}
	return hash, nil
}

func ValidatePasswordHash(password, passwordHash string) bool {
	err := argon2id.Compare(passwordHash, password)
	if err != nil {
		return false
	}
	return true
}

func GenerateVerificationCode() int {
	var src cryptoSource
	rnd := rand.New(src)
	return rnd.Intn(999999)
}

// used for generating a random reset password
const characterSet = "abcdedfghijklmnopqrstABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const pwlength = 20

func GeneratePassword() string {
	var password strings.Builder
	const maxChar = int64(len(characterSet))

	for i := 0; i < pwlength; i++ {
		max := big.NewInt(maxChar)
		bigN, err := crand.Int(crand.Reader, max)
		util.Check(err, "randomly generate int")
		n := bigN.Int64()
		password.WriteString(string(characterSet[n]))
	}
	return password.String()
}

type cryptoSource struct{}

func (s cryptoSource) Seed(seed int64) {}

func (s cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		util.Check(err, "generate random verification code")
	}
	return v
}
