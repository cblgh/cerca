package crypto

import (
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"cerca/util"
	"github.com/synacor/argon2id"
	rand "math/rand"
)

type Keypair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

func GenerateKeypair() (Keypair, error) {
	ed := util.Describe("generate public keypair")
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return Keypair{}, ed.Eout(err, "generating key")
	}
	return Keypair{Public: fmt.Sprintf("%x", pub), Private: fmt.Sprintf("%x", priv)}, nil
}

func (kp *Keypair) Marshal() ([]byte, error) {
	jason, err := json.MarshalIndent(kp, "", " ")
	if err != nil {
		return []byte{}, util.Eout(err, "marshal keypair")
	}
	return jason, nil
}

func (kp *Keypair) Unmarshal(input []byte) error {
	err := json.Unmarshal(input, &kp)
	if err != nil {
		return util.Eout(err, "unmarshal keypair")
	}
	return nil
}

func HashPassword(s string) (string, error) {
	ed := util.Describe("hash password")
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
