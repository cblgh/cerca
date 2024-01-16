package database

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"context"
	"database/sql"
	"strconv"
	"log"
	"errors"
	"cerca/util"
	"github.com/matthewhartstonge/argon2"
)
/* switched argon2 library to support 32 bit due to flaw in previous library. 
* change occurred in commits:
  68a689612547ff83225f9a2727cf0c14dfbf7ceb
  27c6d5684b6b464b900889c4b8a4dbae232d6b68

	migration of the password hashes from synacor's embedded salt format to
	matthewartstonge's key-val embedded format

	migration details: 

	the old format had the following default parameters:
	* time = 1
	* memory = 64MiB
	* threads = 4
	* keyLen = 32
	* saltLen = 16 bytes
	* hashLen = 32 bytes?
	* argonVersion = 13?


	the new format uses the following parameters:
	*	TimeCost:    3,
	*	MemoryCost:  64 * 1024, // 2^(16) (64MiB of RAM)
	*	Parallelism: 4,
	*	SaltLength:  16, // 16 * 8 = 128-bits
	*	HashLength:  32, // 32 * 8 = 256-bits
	*	Mode:        ModeArgon2id,
	*	Version:     Version13,

	the diff:
	* time was changed to 3 from 1
	* the version may or may not be the same (0x13)

	a regex for changing the values would be the following
	old format example value:
	$argon2id19$1,65536,4$111111111111111111111111111111111111111111111111111111111111111111
	old format was also encoding the salt and hash, not in base64 but in a slightly custom format (see var `encoding`)

	regex to grab values
	\$argon2id19\$1,65536,4\$(\S{66})
	diff regex from old to new
	$argon2id$v=19$m=65536,t=${1},p=4${passwordhash}
	new format example value: 
	$argon2id$v=19$m=65536,t=3,p=4$222222222222222222222222222222222222222222222222222222222222222222
*/

func Migration20240116_PwhashChange(filepath string) (finalErr error) {
	d := InitDB(filepath)
	ed := util.Describe("pwhash migration")

	// the encoding defined in the old hashing library for string representations
	// https://github.com/synacor/argon2id/blob/18569dfc600ba1ba89278c3c4789ad81dcab5bfb/argon2id.go#L48
	var encoding = base64.NewEncoding("./ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789").WithPadding(base64.NoPadding)

	// indices of the capture groups, index 0 is the matched string
	const (
		_ = iota
		TIME_INDEX
		SALT_INDEX
		HASH_INDEX
	)

	// regex to parse out: 
	// 1. time parameter
	// 2. salt
	// 3. hash
	const oldArgonPattern = `^\$argon2id19\$(\d),65536,4\$(\S+)\$(\S+)$`
	oldRegex, err := regexp.Compile(oldArgonPattern)
	ed.Check(err, "failed to compile old argon encoding pattern")
	// regex to confirm new records
	const newArgonPattern = `^\$argon2id\$v=19\$m=65536,t=(\d),p=4\$(\S+)\$(\S+)$`
	newRegex, err := regexp.Compile(newArgonPattern)
	ed.Check(err, "failed to compile new argon encoding pattern")

	// always perform migrations in a single transaction
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{})
	rollbackOnErr := func(incomingErr error) bool {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "\nrolling back")
			finalErr = incomingErr
			return true
		}
		return false
	}

	// check table meta's schemaversion to see that it's empty (because i didn't set it initially X)
	row := tx.QueryRow(`SELECT schemaversion FROM meta`)
	placeholder := -1
	err = row.Scan(&placeholder)
	// we *want* to have no rows
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		if rollbackOnErr(err) { return }
	}
	// in this migration, we should *not* have any schemaversion set 
	if placeholder > 0 {
		if rollbackOnErr(errors.New("schemaversion existed! there's a high likelihood that this migration has already been performed - exiting")) {
			return
		}
	}

	// alright onwards to the beesknees
	// data struct to keep passwords and ids together - dont wanna mix things up now do we
	type HashRecord struct {
		id int
		oldFormat string // the full encoded format of prev library. including salt and parameters, not just the hash
		newFormat string // the full encoded format of new library. including salt and parameters, not just the hash
		valid bool // only valid records will be updated (i.e. records whose format is confirmed by the the oldPattern regex)
	}

	var records []HashRecord
	// get all password hashes and the id of their row
	query := `SELECT id, passwordhash FROM users`
	rows, err := tx.Query(query)
	if rollbackOnErr(err) {
		return
	}

	for rows.Next() {
		var record HashRecord
		err = rows.Scan(&record.id, &record.oldFormat)
		if rollbackOnErr(err) {
			return
		}
		if record.id == 0 {
			if rollbackOnErr(errors.New("record id was not changed during scanning")) {
				return
			}
		}
		records = append(records, record)
	}

	// make the requisite pattern changes to the password hash
	config := argon2.MemoryConstrainedDefaults()
	for i := range records {
		// parse out the time, salt, and hash from the old record format
		matches := oldRegex.FindAllStringSubmatch(records[i].oldFormat, -1)
		if len(matches) > 0 {
			time, err := strconv.Atoi(matches[0][TIME_INDEX])
			rollbackOnErr(err)
			salt := matches[0][SALT_INDEX]
			hash := matches[0][HASH_INDEX]

			// decode the old format's had a custom encoding t
			// the correctly access the underlying buffers
			saltBuf, err := encoding.DecodeString(salt)
			util.Check(err, "decode salt using old format encoding")
			hashBuf, err := encoding.DecodeString(hash)
			util.Check(err, "decode hash using old format encoding")

			config.TimeCost = uint32(time) // note this change, to match old time cost (necessary!)
			raw := argon2.Raw{Config:config, Salt: saltBuf, Hash: hashBuf}
			// this is what we will store in the database instead
			newFormatEncoded := raw.Encode()
			ok := newRegex.Match(newFormatEncoded)
			if !ok {
				if rollbackOnErr(errors.New("newly formed format doesn't match regex for new pattern")) {
					return
				}
			}
			records[i].newFormat = string(newFormatEncoded)
			records[i].valid = true
		} else {
			// parsing the old format failed, let's check to see if this happens to be a new record
			// (if it is, we'll just ignore it. but if it's not we error out of here)
			ok := newRegex.MatchString(records[i].oldFormat)
			if !ok {
				// can't parse with regex matching old format or the new format 
				if rollbackOnErr(errors.New(fmt.Sprintf("unknown record format: %s", records[i].oldFormat))) {
					return
				}
			}
		}
	}

	fmt.Println(records)

	fmt.Println("parsed and re-encoded all valid records from the old to the new format. proceeding to update database records")
	for _, record := range records {
		if !record.valid {
			continue
		}
		// update each row with the password hash in the new format
		stmt, err := tx.Prepare("UPDATE users SET passwordhash = ? WHERE id = ?")
		defer stmt.Close()
		_, err = stmt.Exec(record.newFormat, record.id)
		if rollbackOnErr(err) {
			return
		}
	}
	fmt.Println("all records were updated without any error")
	// when everything is done and dudsted insert schemaversion and set its value to 1
	// _, err = tx.Exec(`INSERT INTO meta (schemaversion) VALUES (1)`)
	// if rollbackOnErr(err) {
	// 	return
	// }
	_ = tx.Commit()
	return
}
