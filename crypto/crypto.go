package crypto

import (
	"cerca/logger"
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	rand "math/rand"

	"github.com/synacor/argon2id"
)

type Keypair struct {
	Public  string `json:"public"`
	Private string `json:"private"`
}

func GenerateKeypair() (Keypair, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return Keypair{}, fmt.Errorf("generating public keypair: %w", err)
	}
	return Keypair{Public: fmt.Sprintf("%x", pub), Private: fmt.Sprintf("%x", priv)}, nil
}

func (kp *Keypair) Marshal() ([]byte, error) {
	jason, err := json.MarshalIndent(kp, "", " ")
	if err != nil {
		return []byte{}, fmt.Errorf("marshalling keypair: %w", err)
	}
	return jason, nil
}

func (kp *Keypair) Unmarshal(input []byte) error {
	err := json.Unmarshal(input, &kp)
	if err != nil {
		return fmt.Errorf("unmarshaling keypair: %w", err)
	}
	return nil
}

func HashPassword(s string) (string, error) {
	hash, err := argon2id.DefaultHashPassword(s)
	if err != nil {
		return "", fmt.Errorf("hashing with argon2id: %w", err)
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
		logger.Fatal("failed to generate random verification code: %w", err)
	}
	return v
}
