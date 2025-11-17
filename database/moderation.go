package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"gomod.cblgh.org/cerca/constants"
	"gomod.cblgh.org/cerca/crypto"
	"gomod.cblgh.org/cerca/util"

	_ "github.com/mattn/go-sqlite3"
)

// there are a bunch of places that reference a user's id, so i don't want to break all of those
//
// i also want to avoid big invisible holes in a conversation's history
//
// remove user performs the following operation:
// 1. checks to see if the DELETED USER exists; otherwise create it and remember its id
//
// 2. if it exists, we swap out the userid for the DELETED_USER in tables:
// - table threads authorid
// - table posts authorid
// - table moderation_log actingid or recipientid
//
// the entry in registrations correlating to userid is removed
// if allowing deletion of post contents as well when removing account,
// userid should be used to get all posts from table posts and change the contents
// to say _deleted_
type RemoveUserOptions struct {
	KeepContent  bool
	KeepUsername bool
}

func (d DB) RemoveUser(userid int, options RemoveUserOptions) (finalErr error) {
	keepContent := options.KeepContent
	keepUsername := options.KeepUsername
	ed := util.Describe("remove user")
	// there is a single user we call the "deleted user", and we make sure this deleted user exists on startup
	// they will take the place of the old user when they remove their account.
	deletedUserID, err := d.GetUserID(DELETED_USER_NAME)
	if err != nil {
		log.Fatalln(ed.Eout(err, "get deleted user id"))
	}
	// create a transaction spanning all our removal-related ops
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{}) // proper tx options?
	rollbackOnErr := func(incomingErr error) bool {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "rolling back")
			finalErr = incomingErr
			return true
		}
		return false
	}

	if rollbackOnErr(ed.Eout(err, "start transaction")) {
		return
	}

	type Triplet struct {
		Desc      string
		Statement string
		Args      []any
	}

	// create prepared statements performing the required removal operations for tables that reference a userid as a
	// foreign key: threads, posts, moderation_log, and registrations

	rawTriples := []Triplet{}
	/* UPDATING THREADS */
	// if we remove the username we shall also have to alter the threads started by this user
	if !keepUsername {
		rawTriples = append(rawTriples, Triplet{"threads stmt", "UPDATE threads SET authorid = ? WHERE authorid = ?", []any{deletedUserID, userid}})
	}

	/* UPDATING POSTS */
	// now for interacting with authored posts, we shall have to handle all permutations of keeping/removing post contents and/or username attribution
	if !keepContent && !keepUsername {
		rawTriples = append(rawTriples, Triplet{"posts stmt", `UPDATE posts SET content = "_deleted_", authorid = ? WHERE authorid = ?`, []any{deletedUserID, userid}})
	} else if keepContent && !keepUsername {
		rawTriples = append(rawTriples, Triplet{"posts stmt", `UPDATE posts SET authorid = ? WHERE authorid = ?`, []any{deletedUserID, userid}})
	} else if !keepContent && keepUsername {
		rawTriples = append(rawTriples, Triplet{"posts stmt", `UPDATE posts SET content = "_deleted_" WHERE authorid = ?`, []any{userid}})
	}

	// TODO (2025-04-13): not sure whether altering modlog history like this is a good idea or not; accountability goes outta the window cause all you can see is "<admin> removed <deleted user>"
	/* UPDATING MODLOGS */
	if !keepUsername {
		rawTriples = append(rawTriples, Triplet{"modlog stmt#1", "UPDATE moderation_log SET recipientid = ? WHERE recipientid = ?", []any{deletedUserID, userid}})
		rawTriples = append(rawTriples, Triplet{"modlog stmt#2", "UPDATE moderation_log SET actingid= ? WHERE actingid = ?", []any{deletedUserID, userid}})
		rawTriples = append(rawTriples, Triplet{"registrations stmt", "DELETE FROM registrations where userid = ?", []any{userid}})
	}

	/* REMOVING CREDENTIALS */
	if !keepUsername {
		// remove the account entirely
		rawTriples = append(rawTriples, Triplet{"delete user stmt", "DELETE FROM users where id = ?", []any{userid}})
	} else {
		// disable using the account by generating and setting a gibberish password
		throwawayPasswordHash, err := crypto.HashPassword(crypto.GeneratePassword())
		if rollbackOnErr(ed.Eout(err, fmt.Sprintf("prepare throwaway password"))) {
			return
		}
		rawTriples = append(rawTriples, Triplet{"nullify logins by replacing user password", "UPDATE users SET passwordhash = ? where id = ?", []any{throwawayPasswordHash, userid}})
	}

	var preparedStmts []*sql.Stmt

	prepStmt := func(rawStmt string) (*sql.Stmt, error) {
		var stmt *sql.Stmt
		stmt, err = tx.Prepare(rawStmt)
		return stmt, err
	}

	for _, triple := range rawTriples {
		prep, err := prepStmt(triple.Statement)
		if rollbackOnErr(ed.Eout(err, fmt.Sprintf("prepare %s", triple.Desc))) {
			return
		}
		defer prep.Close()
		preparedStmts = append(preparedStmts, prep)
	}

	for i, stmt := range preparedStmts {
		triple := rawTriples[i]
		_, err = stmt.Exec(triple.Args...)
		if rollbackOnErr(ed.Eout(err, fmt.Sprintf("exec %s", triple.Desc))) {
			return
		}
	}

	err = tx.Commit()
	ed.Check(err, "commit transaction")
	finalErr = nil
	return
}

func (d DB) AddModerationLog(actingid, recipientid, action int) error {
	ed := util.Describe("add moderation log")
	t := time.Now()
	// we have a recipient
	var err error
	if recipientid > 0 {
		insert := `INSERT INTO moderation_log (actingid, recipientid, action, time) VALUES (?, ?, ?, ?)`
		_, err = d.Exec(insert, actingid, recipientid, action, t)
	} else {
		// we are not listing a recipient
		insert := `INSERT INTO moderation_log (actingid, action, time) VALUES (?, ?, ?)`
		_, err = d.Exec(insert, actingid, action, t)
	}
	if err = ed.Eout(err, "exec prepared statement"); err != nil {
		return err
	}
	return nil
}

type ModerationEntry struct {
	ActingUsername, RecipientUsername, QuorumUsername string
	QuorumDecision                                    bool
	Action                                            int
	Time                                              time.Time
}

func (d DB) GetModerationLogs() []ModerationEntry {
	ed := util.Describe("moderation log")
	query := `SELECT uact.name, urecp.name, uquorum.name, q.decision, m.action, m.time 
	FROM moderation_LOG m 

	LEFT JOIN users uact ON uact.id = m.actingid
	LEFT JOIN users urecp ON urecp.id = m.recipientid

	LEFT JOIN quorum_decisions q ON q.modlogid = m.id
	LEFT JOIN users uquorum ON uquorum.id = q.userid

	ORDER BY time DESC`

	stmt, err := d.db.Prepare(query)
	defer stmt.Close()
	ed.Check(err, "prep stmt")

	rows, err := stmt.Query()
	defer rows.Close()
	util.Check(err, "run query")

	var logs []ModerationEntry
	for rows.Next() {
		var entry ModerationEntry
		var actingUsername, recipientUsername, quorumUsername sql.NullString
		var quorumDecision sql.NullBool
		if err := rows.Scan(&actingUsername, &recipientUsername, &quorumUsername, &quorumDecision, &entry.Action, &entry.Time); err != nil {
			ed.Check(err, "scanning loop")
		}
		if actingUsername.Valid {
			entry.ActingUsername = actingUsername.String
		}
		if recipientUsername.Valid {
			entry.RecipientUsername = recipientUsername.String
		}
		if quorumUsername.Valid {
			entry.QuorumUsername = quorumUsername.String
		}
		if quorumDecision.Valid {
			entry.QuorumDecision = quorumDecision.Bool
		}
		logs = append(logs, entry)
	}
	return logs
}

func (d DB) ProposeModerationAction(proposerid, recipientid, action int) (finalErr error) {
	ed := util.Describe("propose mod action")

	t := time.Now()
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{})
	ed.Check(err, "open transaction")

	rollbackOnErr := func(incomingErr error) bool {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "rolling back")
			finalErr = incomingErr
			return true
		}
		return false
	}

	// start tx
	propRecipientId := -1
	// there should only be one pending proposal of each type for any given recipient
	// so let's check to make sure that's true!
	stmt, err := tx.Prepare("SELECT recipientid FROM moderation_proposals WHERE action = ?")
	defer stmt.Close()
	err = stmt.QueryRow(action).Scan(&propRecipientId)
	if err == nil && propRecipientId != -1 {
		finalErr = tx.Commit()
		return
	}
	// there was no pending proposal of the proposed action for recipient - onwards!

	// add the proposal
	stmt, err = tx.Prepare("INSERT INTO moderation_proposals (proposerid, recipientid, time, action) VALUES (?, ?, ?, ?)")
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare proposal stmt")) {
		return
	}
	_, err = stmt.Exec(proposerid, recipientid, t, action)
	if rollbackOnErr(ed.Eout(err, "insert into proposals table")) {
		return
	}

	// TODO (2023-12-18): hmm how do we do this properly now? only have one constant per action
	// {demote, make admin, remove user} but vary translations for these three depending on if there is also a decision or not?

	// add moderation log that user x proposed action y for recipient z
	stmt, err = tx.Prepare(`INSERT INTO moderation_log (actingid, recipientid, action, time) VALUES (?, ?, ?, ?)`)
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare modlog stmt")) {
		return
	}
	_, err = stmt.Exec(proposerid, recipientid, action, t)
	if rollbackOnErr(ed.Eout(err, "insert into modlog")) {
		return
	}

	err = tx.Commit()
	ed.Check(err, "commit transaction")
	return
}

type ModProposal struct {
	ActingUsername, RecipientUsername string
	ActingID, RecipientID             int
	ProposalID, Action                int
	Time                              time.Time
}

func (d DB) GetProposedActions() []ModProposal {
	ed := util.Describe("get moderation proposals")
	stmt, err := d.db.Prepare(`SELECT mp.id, proposerid, up.name, recipientid, ur.name, action, mp.time 
	FROM moderation_proposals mp
	INNER JOIN users up on mp.proposerid = up.id 
	INNER JOIN users ur on mp.recipientid = ur.id 
	ORDER BY time DESC
	;`)
	defer stmt.Close()
	ed.Check(err, "prepare stmt")
	rows, err := stmt.Query()
	ed.Check(err, "perform query")
	defer rows.Close()
	var proposals []ModProposal
	for rows.Next() {
		var prop ModProposal
		if err = rows.Scan(&prop.ProposalID, &prop.ActingID, &prop.ActingUsername, &prop.RecipientID, &prop.RecipientUsername, &prop.Action, &prop.Time); err != nil {
			ed.Check(err, "error scanning in row data")
		}
		proposals = append(proposals, prop)
	}
	return proposals
}

// finalize a proposal by either confirming or vetoing it, logging the requisite information and then finally executing
// the proposed action itself
func (d DB) FinalizeProposedAction(proposalid, adminid int, decision bool) (finalErr error) {
	ed := util.Describe("finalize proposed mod action")

	t := time.Now()
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{})
	ed.Check(err, "open transaction")

	rollbackOnErr := func(incomingErr error) bool {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "rolling back")
			finalErr = incomingErr
			return true
		}
		return false
	}

	/* start tx */
	// make sure the proposal is still there (i.e. nobody has beat us to acting on it yet)
	stmt, err := tx.Prepare("SELECT 1 FROM moderation_proposals WHERE id = ?")
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare proposal existence stmt")) {
		return
	}
	existence := -1
	err = stmt.QueryRow(proposalid).Scan(&existence)
	// proposal id did not exist (it was probably already acted on!)
	if err != nil {
		_ = tx.Commit()
		return
	}
	// retrieve the proposal & populate with our dramatis personae
	var proposerid, recipientid, proposalAction int
	var proposalDate time.Time
	stmt, err = tx.Prepare(`SELECT proposerid, recipientid, action, time from moderation_proposals WHERE id = ?`)
	defer stmt.Close()
	err = stmt.QueryRow(proposalid).Scan(&proposerid, &recipientid, &proposalAction, &proposalDate)
	if rollbackOnErr(ed.Eout(err, "retrieve proposal vals")) {
		return
	}

	isSelfConfirm := proposerid == adminid
	timeSelfConfirmOK := proposalDate.Add(constants.PROPOSAL_SELF_CONFIRMATION_WAIT)
	// TODO (2024-01-07): render err message in admin view?
	// self confirms are not allowed at this point in time, exit early without performing any changes
	if isSelfConfirm && (decision == constants.PROPOSAL_CONFIRM && !time.Now().After(timeSelfConfirmOK)) {
		err = tx.Commit()
		ed.Check(err, "commit transaction")
		finalErr = nil
		return
	}

	// convert proposed action (semantically different for the sake of logs) from the finalized action
	var action int
	switch proposalAction {
	case constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN:
		action = constants.MODLOG_ADMIN_DEMOTE
	case constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER:
		action = constants.MODLOG_REMOVE_USER
	case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
		action = constants.MODLOG_ADMIN_MAKE
	default:
		ed.Check(errors.New("unknown proposal action"), "convertin proposalAction into action")
	}

	// remove proposal from proposal table as it has been executed as desired
	stmt, err = tx.Prepare("DELETE FROM moderation_proposals WHERE id = ?")
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare proposal removal stmt")) {
		return
	}
	_, err = stmt.Exec(proposalid)
	if rollbackOnErr(ed.Eout(err, "remove proposal from table")) {
		return
	}

	// add moderation log
	stmt, err = tx.Prepare(`INSERT INTO moderation_log (actingid, recipientid, action, time) VALUES (?, ?, ?, ?)`)
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare modlog stmt")) {
		return
	}
	// the admin who proposed the action will be logged as the one performing it
	// get the modlog so we can reference it in the quorum_decisions table. this will be used to augment the moderation
	// log view with quorum info
	result, err := stmt.Exec(proposerid, recipientid, action, t)
	if rollbackOnErr(ed.Eout(err, "insert into modlog")) {
		return
	}
	modlogid, err := result.LastInsertId()
	if rollbackOnErr(ed.Eout(err, "get last insert id")) {
		return
	}

	// update the quorum decisions table so that we can use its info to augment the moderation log view
	stmt, err = tx.Prepare(`INSERT INTO quorum_decisions (userid, decision, modlogid) VALUES (?, ?, ?)`)
	defer stmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare quorum insertion stmt")) {
		return
	}
	// decision = confirm or veto => values true or false
	_, err = stmt.Exec(adminid, decision, modlogid)
	if rollbackOnErr(ed.Eout(err, "execute quorum insertion")) {
		return
	}

	err = tx.Commit()
	ed.Check(err, "commit transaction")

	// the decision was to veto the proposal: there's nothing more to do! except return outta this function ofc ofc
	if decision == constants.PROPOSAL_VETO {
		return
	}
	// perform the actual action; would be preferable to do this in the transaction somehow
	// but hell no am i copying in those bits here X)
	switch proposalAction {
	case constants.MODLOG_ADMIN_PROPOSE_DEMOTE_ADMIN:
		err = d.DemoteAdmin(recipientid)
		ed.Check(err, "remove user", recipientid)
	case constants.MODLOG_ADMIN_PROPOSE_REMOVE_USER:
		// TODO (2025-04-13): introduce/record proposal granularity for admin delete view wrt these booleans
		err = d.RemoveUser(recipientid, RemoveUserOptions{KeepContent: false, KeepUsername: false})
		ed.Check(err, "remove user", recipientid)
	case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
		d.AddAdmin(recipientid)
		ed.Check(err, "add admin", recipientid)
	}
	return
}

type User struct {
	Name               string
	ID                 int
	RegistrationOrigin string
}

func (d DB) AddAdmin(userid int) error {
	ed := util.Describe("add admin")
	// make sure the id exists
	exists, err := d.CheckUserExists(userid)
	if !exists {
		return fmt.Errorf("add admin: userid %d did not exist", userid)
	}
	if err != nil {
		return ed.Eout(err, "CheckUserExists had an error")
	}
	isAdminAlready, err := d.IsUserAdmin(userid)
	if isAdminAlready {
		return fmt.Errorf("userid %d was already an admin", userid)
	}
	if err != nil {
		// some kind of error, let's bubble it up
		return ed.Eout(err, "IsUserAdmin")
	}
	// insert into table, we gots ourselves a new sheriff in town [|:D
	stmt := `INSERT INTO admins (id) VALUES (?)`
	_, err = d.db.Exec(stmt, userid)
	if err != nil {
		return ed.Eout(err, "inserting new admin")
	}
	return nil
}

func (d DB) DemoteAdmin(userid int) error {
	ed := util.Describe("demote admin")
	// make sure the id exists
	exists, err := d.CheckUserExists(userid)
	if !exists {
		return fmt.Errorf("demote admin: userid %d did not exist", userid)
	}
	if err != nil {
		return ed.Eout(err, "CheckUserExists had an error")
	}
	isAdmin, err := d.IsUserAdmin(userid)
	if !isAdmin {
		return fmt.Errorf("demote admin: userid %d was not an admin", userid)
	}
	if err != nil {
		// some kind of error, let's bubble it up
		return ed.Eout(err, "IsUserAdmin")
	}
	// all checks are done: perform the removal
	stmt := `DELETE FROM admins WHERE id = ?`
	_, err = d.db.Exec(stmt, userid)
	if err != nil {
		return ed.Eout(err, "inserting new admin")
	}
	return nil
}

func (d DB) IsUserAdmin(userid int) (bool, error) {
	stmt := `SELECT 1 FROM admins WHERE id = ?`
	return d.existsQuery(stmt, userid)
}

func (d DB) QuorumActivated() bool {
	admins := d.GetAdmins()
	return len(admins) >= 2
}

func (d DB) GetAdmins() []User {
	ed := util.Describe("get admins")
	query := `SELECT u.name, a.id 
  FROM users u 
  INNER JOIN admins a ON u.id = a.id 
  ORDER BY u.name
  `
	stmt, err := d.db.Prepare(query)
	defer stmt.Close()
	ed.Check(err, "prep stmt")

	rows, err := stmt.Query()
	defer rows.Close()
	util.Check(err, "run query")

	var user User
	var admins []User
	for rows.Next() {
		if err := rows.Scan(&user.Name, &user.ID); err != nil {
			ed.Check(err, "scanning loop")
		}
		admins = append(admins, user)
	}
	return admins
}

type InviteBatch struct {
	BatchId          string
	ActingUsername   string
	UnclaimedInvites []string
	Label            string
	Time             time.Time
	Reusable         bool
}

func (d DB) ClaimInvite(invite string) (bool, string, error) {
	ed := util.Describe("claim invite")
	var err error
	var tx *sql.Tx
	tx, err = d.db.BeginTx(context.Background(), &sql.TxOptions{})

	rollbackOnErr := func(incomingErr error) error {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "rolling back")
			return incomingErr
		}
		return nil
	}

	type BatchQuery struct {
		stmt, desc   string
		preparedStmt *sql.Stmt
	}

	ops := []BatchQuery{
		BatchQuery{desc: "check if invite to redeem exists", stmt: "SELECT EXISTS (SELECT 1 FROM invites WHERE invite = ?)"},
		BatchQuery{desc: "get invite code's batchid and whether marked reusable", stmt: "SELECT batchid, reusable FROM invites WHERE invite = ?"},
		BatchQuery{desc: "delete invite from table", stmt: "DELETE FROM invites WHERE invite = ?"},
	}

	for i, operation := range ops {
		ops[i].preparedStmt, err = tx.Prepare(operation.stmt)
		defer ops[i].preparedStmt.Close()
		if e := rollbackOnErr(ed.Eout(err, operation.desc)); e != nil {
			return false, "", e
		}
	}

	// first: check if the invite still exists; uses QueryRow to get back results
	row := ops[0].preparedStmt.QueryRow(invite)
	var exists int
	err = row.Scan(&exists)
	if e := rollbackOnErr(ed.Eout(err, "exec "+ops[0].desc)); e != nil {
		return false, "", e
	}

	// existence check failed. end transaction by rolling back (nothing meaningful was changed)
	if exists == 0 {
		_ = tx.Rollback()
		return false, "", nil
	}

	// then: get the associated batchid, so we can associate it with the registration
	row = ops[1].preparedStmt.QueryRow(invite)
	var batchid string // uuid v4
	var reusable bool
	err = row.Scan(&batchid, &reusable)
	if e := rollbackOnErr(ed.Eout(err, "exec "+ops[1].desc)); e != nil {
		return false, "", e
	}

	if !reusable {
		// then, finally: delete the invite code being claimed
		_, err = ops[2].preparedStmt.Exec(invite)
		if e := rollbackOnErr(ed.Eout(err, "exec "+ops[2].desc)); e != nil {
			return false, "", e
		}
	}

	err = tx.Commit()
	ed.Check(err, "commit transaction")
	return true, batchid, nil
}

const maxBatchAmount = 100
const maxUnclaimedAmount = 500

func (d DB) CreateInvites(adminid int, amount int, label string, reusable bool) error {
	ed := util.Describe("create invites")
	isAdmin, err := d.IsUserAdmin(adminid)
	if err != nil {
		return ed.Eout(err, "IsUserAdmin")
	}

	if !isAdmin {
		return fmt.Errorf("userid %d was not an admin, they can't create an invite", adminid)
	}

	// check that amount is within reasonable range
	if amount > maxBatchAmount {
		return fmt.Errorf("batch amount should not exceed %d but was %d; not creating invites ", maxBatchAmount, amount)
	}

	// check that already existing unclaimed invites is within a reasonable range
	stmt := "SELECT COUNT(*) FROM invites"
	var unclaimed int
	err = d.db.QueryRow(stmt).Scan(&unclaimed)
	ed.Check(err, "querying for number of unclaimed invites")
	if unclaimed > maxUnclaimedAmount {
		msgstr := "number of unclaimed invites amount should not exceed %d but was %d; ceasing invite creation"
		return fmt.Errorf(msgstr, maxUnclaimedAmount, unclaimed)
	}

	// all cleared!
	invites := make([]string, 0, amount)
	for i := 0; i < amount; i++ {
		invites = append(invites, util.GetUUIDv4())
	}
	// adjust the amount that will be created if we are near the unclaimed amount threshold
	if (amount + unclaimed) > maxUnclaimedAmount {
		amount = maxUnclaimedAmount - unclaimed
	}

	if amount <= 0 {
		return fmt.Errorf("number of unclaimed invites amount %d has been reached; not creating invites ", maxUnclaimedAmount)
	}

	// this id identifies all invites from this batch
	batchid := util.GetUUIDv4()
	creationTime := time.Now()
	preparedStmt, err := d.db.Prepare("INSERT INTO invites (batchid, adminid, invite, label, time, reusable) VALUES (?, ?, ?, ?, ?, ?)")
	util.Check(err, "prepare invite insert stmt")
	defer preparedStmt.Close()
	for _, invite := range invites {
		// create a batch
		_, err := preparedStmt.Exec(batchid, adminid, invite, label, creationTime, reusable)
		ed.Check(err, "inserting invite into database")
	}
	return nil
}

func (d DB) DestroyInvites(invites []string) {
	ed := util.Describe("destroy invites")
	stmt := "DELETE FROM invites WHERE invite = ?"
	for _, invite := range invites {
		_, err := d.Exec(stmt, invite)
		// note: it's okay if one of the statements fails, maybe someone lucked out and redeemed it in the middle of the
		// loop - whatever
		if err != nil {
			log.Println(ed.Eout(err, "err during exec"))
		}
	}
}

func (d DB) DeleteInvitesBatch(batchid string) {
	ed := util.Describe("delete invites by batchid")

	stmt, err := d.db.Prepare("DELETE FROM invites where batchid = ?")
	ed.Check(err, "prep stmt")
	defer stmt.Close()

	_, err = stmt.Exec(batchid)
	util.Check(err, "execute delete")
}

func (d DB) GetAllInvites() []InviteBatch {
	ed := util.Describe("get all invites")

	rows, err := d.db.Query("SELECT i.batchid, u.name, i.invite, i.time, i.label, i.reusable FROM invites i INNER JOIN users u ON i.adminid = u.id")
	ed.Check(err, "create query")

	// keep track of invite batches by creating a key based on username + creation time
	batches := make(map[string]*InviteBatch)
	var keys []string
	var batchid, invite, username, label string
	var t time.Time
	var reusable bool

	for rows.Next() {
		err := rows.Scan(&batchid, &username, &invite, &t, &label, &reusable)
		ed.Check(err, "scan row")
		// starting the key with the unix epoch as a string allows us to sort the map's keys by time just by comparing strings with sort.Strings()
		unixTimestamp := strconv.FormatInt(t.Unix(), 10)
		key := unixTimestamp + username
		if batch, exists := batches[key]; exists {
			batch.UnclaimedInvites = append(batch.UnclaimedInvites, invite)
		} else {
			keys = append(keys, key)
			batches[key] = &InviteBatch{BatchId: batchid, ActingUsername: username, UnclaimedInvites: []string{invite}, Label: label, Time: t, Reusable: reusable}
		}
	}

	// convert from map to a []InviteBatch sorted by time using ts-prefixed map keys
	ret := make([]InviteBatch, 0, len(keys))
	// we want newest first
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	for _, key := range keys {
		ret = append(ret, *batches[key])
	}
	return ret
}

type RegisteredInvite struct {
	Label   string
	BatchID string
	Count   int
}

func (d DB) CountRegistrationsByInviteBatch() []RegisteredInvite {
	ed := util.Describe("database/moderation.go: count registrations by invite batch")
	stmt := `SELECT i.label, i.batchid, COUNT(*) 
	FROM registrations r INNER JOIN invites i 
	ON r.link == i.batchid 
	GROUP BY i.batchid 
	ORDER BY COUNT(*) DESC 
	;`
	rows, err := d.db.Query(stmt)
	ed.Check(err, "query stmt")
	var registrations []RegisteredInvite
	for rows.Next() {
		var info RegisteredInvite
		err = rows.Scan(&info.Label, &info.BatchID, &info.Count)
		ed.Check(err, "failed to scan returned result")
		registrations = append(registrations, info)
	}
	return registrations
}
