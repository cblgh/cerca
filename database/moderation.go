package database
import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"cerca/util"
	"cerca/constants"

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
func (d DB) RemoveUser(userid int) (finalErr error) {
	ed := util.Describe("remove user")
	// there is a single user we call the "deleted user", and we make sure this deleted user exists on startup
	// they will take the place of the old user when they remove their account.
	deletedUserID, err := d.GetUserID(DELETED_USER_NAME)
	if err != nil {
		log.Fatalln(ed.Eout(err, "get deleted user id"))
	}
	// create a transaction spanning all our removal-related ops
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{}) // proper tx options?
	rollbackOnErr:= func(incomingErr error) bool {
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

	// create prepared statements performing the required removal operations for tables that reference a userid as a
	// foreign key: threads, posts, moderation_log, and registrations
	threadsStmt, err := tx.Prepare("UPDATE threads SET authorid = ? WHERE authorid = ?")
	defer threadsStmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare threads stmt")) {
		return
	}

	postsStmt, err := tx.Prepare(`UPDATE posts SET content = "_deleted_", authorid = ? WHERE authorid = ?`)
	defer postsStmt.Close()
	if rollbackOnErr(ed.Eout(err, "prepare posts stmt")) {
		return
	}

	modlogStmt1, err := tx.Prepare("UPDATE moderation_log SET recipientid = ? WHERE recipientid = ?")
	defer modlogStmt1.Close()
	if rollbackOnErr(ed.Eout(err, "prepare modlog stmt #1")) {
		return
	}

	modlogStmt2, err := tx.Prepare("UPDATE moderation_log SET actingid = ? WHERE actingid = ?")
	defer modlogStmt2.Close()
	if rollbackOnErr(ed.Eout(err, "prepare modlog stmt #2")) {
		return
	}

	stmtReg, err := tx.Prepare("DELETE FROM registrations where userid = ?")
	defer stmtReg.Close()
	if rollbackOnErr(ed.Eout(err, "prepare registrations stmt")) {
		return
	}

	// and finally: removing the entry from the user's table itself
	stmtUsers, err := tx.Prepare("DELETE FROM users where id = ?")
	defer stmtUsers.Close()
	if rollbackOnErr(ed.Eout(err, "prepare users stmt")) {
		return
	}

	_, err = threadsStmt.Exec(deletedUserID, userid)
	if rollbackOnErr(ed.Eout(err, "exec threads stmt")) {
		return
	}
	_, err = postsStmt.Exec(deletedUserID, userid)
	if rollbackOnErr(ed.Eout(err, "exec posts stmt")) {
		return
	}
	_, err = modlogStmt1.Exec(deletedUserID, userid)
	if rollbackOnErr(ed.Eout(err, "exec modlog #1 stmt")) {
		return
	}
	_, err = modlogStmt2.Exec(deletedUserID, userid)
	if rollbackOnErr(ed.Eout(err, "exec modlog #2 stmt")) {
		return
	}
	_, err = stmtReg.Exec(userid)
	if rollbackOnErr(ed.Eout(err, "exec registration stmt")) {
		return
	}
	_, err = stmtUsers.Exec(userid)
	if rollbackOnErr(ed.Eout(err, "exec users stmt")) {
		return
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
	QuorumDecision bool
	Action int
	Time time.Time
}

func (d DB) GetModerationLogs () []ModerationEntry {
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

	rollbackOnErr:= func(incomingErr error) bool {
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
	ActingID, RecipientID int
	ProposalID, Action int
	Time time.Time
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

	rollbackOnErr:= func(incomingErr error) bool {
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

	timeSelfConfirmOK := proposalDate.Add(constants.PROPOSAL_SELF_CONFIRMATION_WAIT)
	// TODO (2024-01-07): render err message in admin view?
	// self confirms are not allowed at this point in time, exit early without performing any changes
	if decision == constants.PROPOSAL_CONFIRM && !time.Now().After(timeSelfConfirmOK) {
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
		err = d.RemoveUser(recipientid)
		ed.Check(err, "remove user", recipientid)
	case constants.MODLOG_ADMIN_PROPOSE_MAKE_ADMIN:
		d.AddAdmin(recipientid)
		ed.Check(err, "add admin", recipientid)
	}
	return
}

type User struct {
	Name string
	ID int
}

func (d DB) AddAdmin(userid int) error {
	ed := util.Describe("add admin")
	// make sure the id exists
	exists, err := d.CheckUserExists(userid)
	if !exists {
		return errors.New(fmt.Sprintf("add admin: userid %d did not exist", userid))
	}
	if err != nil {
		return ed.Eout(err, "CheckUserExists had an error")
	}
	isAdminAlready, err := d.IsUserAdmin(userid)
	if isAdminAlready {
		return errors.New(fmt.Sprintf("userid %d was already an admin", userid))
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
		return errors.New(fmt.Sprintf("demote admin: userid %d did not exist", userid))
	}
	if err != nil {
		return ed.Eout(err, "CheckUserExists had an error")
	}
	isAdmin, err := d.IsUserAdmin(userid)
	if !isAdmin {
		return errors.New(fmt.Sprintf("demote admin: userid %d was not an admin", userid))
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

func (d DB) IsUserAdmin (userid int) (bool, error) {
	stmt := `SELECT 1 FROM admins WHERE id = ?`
	return d.existsQuery(stmt, userid)
}

func (d DB) QuorumActivated () bool {
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
