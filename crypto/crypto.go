package crypto

import (
	"cerca/util"
	"crypto/ed25519"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/synacor/argon2id"
	rand "math/rand"
	"os"
	"strings"
	"time"
)

type Keypair struct {
	Public  ed25519.PublicKey  `json:"public"`
	Private ed25519.PrivateKey `json:"private"`
}

func GenerateKeypair() (Keypair, error) {
	ed := util.Describe("generate public keypair")
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return Keypair{}, ed.Eout(err, "generating key")
	}
	return Keypair{Public: pub, Private: priv}, nil
}

func ReadKeypair(kpPath string) (Keypair, error) {
	var kp Keypair
	ed := util.Describe("read keypair from disk")
	b, err := os.ReadFile(kpPath)
	if err != nil {
		return kp, ed.Eout(err, "read file")
	}
	err = kp.Unmarshal(b)
	if err != nil {
		return kp, ed.Eout(err, "unmarshal kp")
	}
	return kp, nil
}

func CreateProof(kp Keypair, payload []byte) []byte {
	return ed25519.Sign(kp.Private, payload)
}

func VerifyProof(public ed25519.PublicKey, payload, proof []byte) bool {
	return ed25519.Verify(public, payload, proof)
}

/* kinda cludgy oh well */
func (kp *Keypair) PublicString() (string, error) {
	b, err := json.Marshal(kp.Public)
	if err != nil {
		return "", util.Eout(err, "marshal public key")
	}
	return string(b), nil
}

func (kp *Keypair) PrivateString() (string, error) {
	b, err := json.Marshal(kp.Private)
	if err != nil {
		return "", util.Eout(err, "marshal private key")
	}
	return string(b), nil
}

func (kp *Keypair) Marshal() ([]byte, error) {
	jason, err := json.MarshalIndent(kp, "", " ")
	if err != nil {
		return []byte{}, util.Eout(err, "marshal keypair")
	}
	return jason, nil
}

func (kp *Keypair) Unmarshal(input []byte) error {
	ed := util.Describe("unmarshal keypair")
	type stringKp struct {
		Public  string `json:"public"`
		Private string `json:"private"`
	}
	var m stringKp
	err := json.Unmarshal(input, &m)
	if err != nil {
		return ed.Eout(err, "unmarshal into string struct")
	}

	// handle the unfortunate case that the first generated keypairs were all in hex :')
	// meaning: convert them from hex to base64 (the format expected by crypto/ed25519)
	if len(m.Private) == 128 {
		convertedVal, err := util.Hex2Base64(m.Private)
		if err != nil {
			return ed.Eout(err, "failed to convert privkey hex to base64")
		}
		m.Private = convertedVal
		convertedVal, err = util.Hex2Base64(m.Public)
		if err != nil {
			return ed.Eout(err, "failed to convert pubkey hex to base64")
		}
		m.Public = convertedVal

		// marshal the corrected version to a slice of bytes, so that we can pretend this debacle never happened
		input, err = json.Marshal(m)
		if err != nil {
			return ed.Eout(err, "failed to marshal converted hex")
		}
	}

	err = json.Unmarshal(input, kp)
	if err != nil {
		return ed.Eout(err, "unmarshal keypair")
	}
	return nil
}

func PublicKeyFromString(s string) ed25519.PublicKey {
	ed := util.Describe("public key from string")
	var err error
	// handle legacy case of some pubkeys being stored in wrong column due to faulty query <.<
	s = strings.ReplaceAll(s, `"`, "")
	// handle legacy case of some pubkeys being stored as hex
	if len(s) == 64 {
		s, err = util.Hex2Base64(s)
		ed.Check(err, "convert hex to base64")
	}
	b, err := base64.StdEncoding.DecodeString(s)
	ed.Check(err, "decode base64 string")
	pub := (ed25519.PublicKey)(b)
	return pub
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

func GenerateNonce() string {
	const MaxUint = ^uint(0)
	const MaxInt = int(MaxUint >> 1)
	var src cryptoSource
	rnd := rand.New(src)
	return fmt.Sprintf("%d%d", time.Now().Unix(), rnd.Intn(MaxInt))
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
