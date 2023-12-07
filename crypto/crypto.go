package crypto

import (
	"cerca/util"
	crand "crypto/rand"
	"encoding/binary"
	// "github.com/synacor/argon2id"
	"github.com/matthewhartstonge/argon2"
	"math/big"
	rand "math/rand"
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

// TODO (2023-12-05): figure out migration of the password hashes from synacor's embedded salt format to
// matthewartstonge's key-val embedded format

// migration details: 
//
// the old format had the following default parameters:
// * time = 1
// * memory = 64MiB
// * threads = 4
// * keyLen = 32
// * saltLen = 16 bytes
// * hashLen = 32 bytes?
// * argonVersion = 13?
// 
//
// the new format uses the following parameters:
// *	HashLength:  32, // 32 * 8 = 256-bits
// *	SaltLength:  16, // 16 * 8 = 128-bits
// *	TimeCost:    3,
// *	MemoryCost:  64 * 1024, // 2^(16) (64MiB of RAM)
// *	Parallelism: 4,
// *	Mode:        ModeArgon2id,
// *	Version:     Version13,
//
// the diff:
// * time was changed to 3 from 1
// * the version may or may not be the same (0x13)
//
// a regex for changing the values would be the following
// old format example value:
// $argon2id19$1,65536,4$111111111111111111111111111111111111111111111111111111111111111111
// diff regex from old to new
// $argon2id$v=19$m=65536,t=${1},p=4${passwordhash}
// new format example value: 
// $argon2id$v=19$m=65536,t=3,p=4$222222222222222222222222222222222222222222222222222222222222222222
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
		bigN, err := crand.Int(crand.Reader, max)
		util.Check(err, "randomly generate int")
		n := bigN.Int64()
		password.WriteString(string(characterSet[n]))
	}
	return password.String()
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
