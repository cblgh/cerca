package database

import (
	"context"
	"database/sql"
	"cerca/crypto"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"os"
	"time"

	"cerca/util"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func CheckExists(filepath string) bool {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func InitDB(filepath string) DB {
	exists := CheckExists(filepath)
	if !exists {
		file, err := os.Create(filepath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
	}

	db, err := sql.Open("sqlite3", filepath)
	util.Check(err, "opening sqlite3 database at %s", filepath)
	if db == nil {
		log.Fatalln("db is nil")
	}
	createTables(db)
	instance := DB{db}
	instance.makeSureDefaultUsersExist()
	return instance
}

const DELETED_USER_NAME = "deleted user"
func (d DB) makeSureDefaultUsersExist() {
	ed := util.Describe("create default users")
	deletedUserExists, err := d.CheckUsernameExists(DELETED_USER_NAME)
	if err != nil {
		log.Fatalln(ed.Eout(err, "check username exists"))
	}
	if !deletedUserExists {
		passwordHash, err := crypto.HashPassword(crypto.GeneratePassword())
		_, err = d.CreateUser(DELETED_USER_NAME, passwordHash)
		if err != nil {
			log.Fatalln(ed.Eout(err, "create deleted user"))
		}
	}
}

func createTables(db *sql.DB) {
	// create the table if it doesn't exist
	queries := []string{
		/* used for versioning migrations */
		`
  CREATE TABLE IF NOT EXISTS meta (
    schemaversion INTEGER NOT NULL
  );
  `,
		`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    passwordhash TEXT NOT NULL
  );
  `,
		`
  CREATE TABLE IF NOT EXISTS admins(
    id INTEGER PRIMARY KEY
  );
  `,
		`
  CREATE TABLE IF NOT EXISTS moderation_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
		actingid INTEGER NOT NULL,
		recipientid INTEGER,
		action INTEGER,
    time DATE,

    FOREIGN KEY (actingid) REFERENCES users(id),
    FOREIGN KEY (recipientid) REFERENCES users(id)
  );
  `,
		`
  CREATE TABLE IF NOT EXISTS registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userid INTEGER,
    host STRING,
    link STRING,
    time DATE,
    FOREIGN KEY (userid) REFERENCES users(id)
  );
  `,

		/* also known as forum categories; buckets of threads */
		`
  CREATE TABLE IF NOT EXISTS topics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    description TEXT
  );
  `,
		/* thread link structure: <domain>.<tld>/thread/<id>/[<blurb>] */
		`
  CREATE TABLE IF NOT EXISTS threads (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    publishtime DATE,
    topicid INTEGER,
    authorid INTEGER,
    FOREIGN KEY(topicid) REFERENCES topics(id),
    FOREIGN KEY(authorid) REFERENCES users(id)
  );
  `,
		`
  CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    content TEXT NOT NULL,
    publishtime DATE,
    lastedit DATE,
    authorid INTEGER,
    threadid INTEGER,
    FOREIGN KEY(authorid) REFERENCES users(id),
    FOREIGN KEY(threadid) REFERENCES threads(id)
  );
  `}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			log.Fatalln(util.Eout(err, "creating database table %s", query))
		}
	}
}

/* goal for 2021-12-26
* create thread
* create post
* get thread
* + html render of begotten thread
 */

/* goal for 2021-12-28
* in browser: reply on a thread
* in browser: create a new thread
 */
func (d DB) Exec(stmt string, args ...interface{}) (sql.Result, error) {
	return d.db.Exec(stmt, args...)
}

func (d DB) CreateThread(title, content string, authorid, topicid int) (int, error) {
	ed := util.Describe("create thread")
	// create the new thread in a transaction spanning two statements
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{}) // proper tx options?
	ed.Check(err, "start transaction")
	// first, create the new thread
	publish := time.Now()
	threadStmt := `INSERT INTO threads (title, publishtime, topicid, authorid) VALUES (?, ?, ?, ?)
  RETURNING id`
	replyStmt := `INSERT INTO posts (content, publishtime, threadid, authorid) VALUES (?, ?, ?, ?)`
	var threadid int
	err = tx.QueryRow(threadStmt, title, publish, topicid, authorid).Scan(&threadid)
	if err = ed.Eout(err, "add thread %s by %d in topic %d", title, authorid, topicid); err != nil {
		_ = tx.Rollback()
		log.Println(err, "rolling back")
		return -1, err
	}
	// then add the content as the first reply to the thread
	_, err = tx.Exec(replyStmt, content, publish, threadid, authorid)
	if err = ed.Eout(err, "add initial reply for thread %d", threadid); err != nil {
		_ = tx.Rollback()
		log.Println(err, "rolling back")
		return -1, err
	}
	err = tx.Commit()
	ed.Check(err, "commit transaction")
	// finally return the id of the created thread, so we can do a friendly redirect
	return threadid, nil
}

// c.f.
// https://medium.com/aubergine-solutions/how-i-handled-null-possible-values-from-database-rows-in-golang-521fb0ee267
// type NullTime sql.NullTime
type Post struct {
	ID          int
	ThreadTitle string
	Content     template.HTML
	Author      string
	AuthorID    int
	Publish     time.Time
	LastEdit    sql.NullTime // TODO: handle json marshalling with custom type
}

func (d DB) DeleteThread() {}
func (d DB) MoveThread()   {}

// TODO(2021-12-28): return error if non-existent thread
func (d DB) GetThread(threadid int) []Post {
	// TODO: make edit work if no edit timestamp detected e.g.
	// (sql: Scan error on column index 3, name "lastedit": unsupported Scan, storing driver.Value type <nil> into type
	// *time.Time)

	// join with:
	//    users table to get user name
	//    threads table to get thread title
	query := `
  SELECT p.id, t.title, content, u.name, p.authorid, p.publishtime, p.lastedit
  FROM posts p 
  INNER JOIN users u ON u.id = p.authorid 
  INNER JOIN threads t ON t.id = p.threadid
  WHERE threadid = ? 
  ORDER BY p.publishtime
  `
	stmt, err := d.db.Prepare(query)
	util.Check(err, "get thread: prepare query")
	defer stmt.Close()

	rows, err := stmt.Query(threadid)
	util.Check(err, "get thread: query")
	defer rows.Close()

	var data Post
	var posts []Post
	for rows.Next() {
		if err := rows.Scan(&data.ID, &data.ThreadTitle, &data.Content, &data.Author, &data.AuthorID, &data.Publish, &data.LastEdit); err != nil {
			log.Fatalln(util.Eout(err, "get data for thread %d", threadid))
		}
		posts = append(posts, data)
	}
	return posts
}

func (d DB) GetPost(postid int) (Post, error) {
	stmt := `
  SELECT p.id, t.title, content, u.name, p.authorid, p.publishtime, p.lastedit
  FROM posts p 
  INNER JOIN users u ON u.id = p.authorid 
  INNER JOIN threads t ON t.id = p.threadid
  WHERE p.id = ?
  `
	var data Post
	err := d.db.QueryRow(stmt, postid).Scan(&data.ID, &data.ThreadTitle, &data.Content, &data.Author, &data.AuthorID, &data.Publish, &data.LastEdit)
	err = util.Eout(err, "get data for thread %d", postid)
	return data, err
}

type Thread struct {
	Title   string
	Author  string
	Slug    string
	ID      int
	Publish time.Time
	PostID  int
}

// get a list of threads
// NOTE: this query is setting thread.Author not by thread creator, but latest poster. if this becomes a problem, revert
// its use and employ Thread.PostID to perform another query for each thread to get the post author name (wrt server.go:GenerateRSS)
func (d DB) ListThreads(sortByPost bool) []Thread {
	query := `
  SELECT count(t.id), t.title, t.id, u.name, p.publishtime, p.id FROM threads t
  INNER JOIN users u on u.id = p.authorid
  INNER JOIN posts p ON t.id = p.threadid
  GROUP BY t.id
  %s
  `
	orderBy := `ORDER BY t.publishtime DESC`
	// get a list of threads by ordering them based on most recent post
	if sortByPost {
		orderBy = `ORDER BY max(p.id) DESC`
	}
	query = fmt.Sprintf(query, orderBy)

	stmt, err := d.db.Prepare(query)
	util.Check(err, "list threads: prepare query")
	defer stmt.Close()

	rows, err := stmt.Query()
	util.Check(err, "list threads: query")
	defer rows.Close()

	var postCount int
	var data Thread
	var threads []Thread
	for rows.Next() {
		if err := rows.Scan(&postCount, &data.Title, &data.ID, &data.Author, &data.Publish, &data.PostID); err != nil {
			log.Fatalln(util.Eout(err, "list threads: read in data via scan"))
		}
		data.Slug = util.GetThreadSlug(data.ID, data.Title, postCount)
		threads = append(threads, data)
	}
	return threads
}

func (d DB) AddPost(content string, threadid, authorid int) (postID int) {
	stmt := `INSERT INTO posts (content, publishtime, threadid, authorid) VALUES (?, ?, ?, ?) RETURNING id`
	publish := time.Now()
	err := d.db.QueryRow(stmt, content, publish, threadid, authorid).Scan(&postID)
	util.Check(err, "add post to thread %d (author %d)", threadid, authorid)
	return
}

func (d DB) EditPost(content string, postid int) {
	stmt := `UPDATE posts set content = ?, lastedit = ? WHERE id = ?`
	edit := time.Now()
	_, err := d.Exec(stmt, content, edit, postid)
	util.Check(err, "edit post %d", postid)
}

func (d DB) DeletePost(postid int) error {
	stmt := `DELETE FROM posts WHERE id = ?`
	_, err := d.Exec(stmt, postid)
	return util.Eout(err, "deleting post %d", postid)
}

func (d DB) CreateTopic(title, description string) {
	stmt := `INSERT INTO topics (name, description) VALUES (?, ?)`
	_, err := d.Exec(stmt, title, description)
	util.Check(err, "creating topic %s", title)
}

func (d DB) UpdateTopicName(topicid int, newname string) {
	stmt := `UPDATE topics SET name = ? WHERE id = ?`
	_, err := d.Exec(stmt, newname, topicid)
	util.Check(err, "changing topic %d's name to %s", topicid, newname)
}

func (d DB) UpdateTopicDescription(topicid int, newdesc string) {
	stmt := `UPDATE topics SET description = ? WHERE id = ?`
	_, err := d.Exec(stmt, newdesc, topicid)
	util.Check(err, "changing topic %d's description to %s", topicid, newdesc)
}

func (d DB) DeleteTopic(topicid int) {
	stmt := `DELETE FROM topics WHERE id = ?`
	_, err := d.Exec(stmt, topicid)
	util.Check(err, "deleting topic %d", topicid)
}

func (d DB) CreateUser(name, hash string) (int, error) {
	stmt := `INSERT INTO users (name, passwordhash) VALUES (?, ?) RETURNING id`
	var userid int
	err := d.db.QueryRow(stmt, name, hash).Scan(&userid)
	if err != nil {
		return -1, util.Eout(err, "creating user %s", name)
	}
	return userid, nil
}

func (d DB) GetUserID(name string) (int, error) {
	stmt := `SELECT id FROM users where name = ?`
	var userid int
	err := d.db.QueryRow(stmt, name).Scan(&userid)
	if err != nil {
		return -1, util.Eout(err, "get user id")
	}
	return userid, nil
}

func (d DB) GetUsername(uid int) (string, error) {
	stmt := `SELECT name FROM users where id = ?`
	var username string
	err := d.db.QueryRow(stmt, uid).Scan(&username)
	if err != nil {
		return "", util.Eout(err, "get username")
	}
	return username, nil
}

func (d DB) GetPasswordHash(username string) (string, int, error) {
	stmt := `SELECT passwordhash, id FROM users where name = ?`
	var hash string
	var userid int
	err := d.db.QueryRow(stmt, username).Scan(&hash, &userid)
	if err != nil {
		return "", -1, util.Eout(err, "get password hash")
	}
	return hash, userid, nil
}

func (d DB) existsQuery(substmt string, args ...interface{}) (bool, error) {
	stmt := fmt.Sprintf(`SELECT exists (%s)`, substmt)
	var exists bool
	err := d.db.QueryRow(stmt, args...).Scan(&exists)
	if err != nil {
		return false, util.Eout(err, "exists: %s", substmt)
	}
	return exists, nil
}

func (d DB) CheckUserExists(userid int) (bool, error) {
	stmt := `SELECT 1 FROM users WHERE id = ?`
	return d.existsQuery(stmt, userid)
}

func (d DB) CheckUsernameExists(username string) (bool, error) {
	stmt := `SELECT 1 FROM users WHERE name = ?`
	return d.existsQuery(stmt, username)
}

func (d DB) UpdateUserName(userid int, newname string) {
	stmt := `UPDATE users SET name = ? WHERE id = ?`
	_, err := d.Exec(stmt, newname, userid)
	util.Check(err, "changing user %d's name to %s", userid, newname)
}

func (d DB) UpdateUserPasswordHash(userid int, newhash string) {
	stmt := `UPDATE users SET passwordhash = ? WHERE id = ?`
	_, err := d.Exec(stmt, newhash, userid)
	util.Check(err, "changing user %d's description to %s", userid, newhash)
}

// there are a bunch of places that reference a user's id, so i don't want to break all of those
//
// i also want to avoid big invisible holes in a conversation's history

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
	rollbackOnErr:= func(incomingErr error) {
		if incomingErr != nil {
			_ = tx.Rollback()
			log.Println(incomingErr, "rolling back")
			finalErr = incomingErr
			return 
		}
	}
	rollbackOnErr(ed.Eout(err, "start transaction"))

	// create prepared statements performing the required removal operations for tables that reference a userid as a
	// foreign key: threads, posts, moderation_log, and registrations
	threadsStmt, err := tx.Prepare("UPDATE threads SET authorid = ? WHERE authorid = ?")
	rollbackOnErr(ed.Eout(err, "prepare threads stmt"))
	defer threadsStmt.Close()

	postsStmt, err := tx.Prepare(`UPDATE posts SET content = "_deleted_", authorid = ? WHERE authorid = ?`)
	rollbackOnErr(ed.Eout(err, "prepare posts stmt"))
	defer postsStmt.Close()

	modlogStmt1, err := tx.Prepare("UPDATE moderation_log SET recipientid = ? WHERE recipientid = ?")
	rollbackOnErr(ed.Eout(err, "prepare modlog stmt #1"))
	defer modlogStmt1.Close()

	modlogStmt2, err := tx.Prepare("UPDATE moderation_log SET actingid = ? WHERE actingid = ?")
	rollbackOnErr(ed.Eout(err, "prepare modlog stmt #2"))
	defer modlogStmt2.Close()

	stmtReg, err := tx.Prepare("DELETE FROM registrations where userid = ?")
	rollbackOnErr(ed.Eout(err, "prepare registrations stmt"))
	defer stmtReg.Close()

	// and finally: removing the entry from the user's table itself
	stmtUsers, err := tx.Prepare("DELETE FROM users where id = ?")
	rollbackOnErr(ed.Eout(err, "prepare users stmt"))
	defer stmtUsers.Close()

	_, err = threadsStmt.Exec(deletedUserID, userid)
	rollbackOnErr(ed.Eout(err, "exec threads stmt"))
	_, err = postsStmt.Exec(deletedUserID, userid)
	rollbackOnErr(ed.Eout(err, "exec posts stmt"))
	_, err = modlogStmt1.Exec(deletedUserID, userid)
	fmt.Println("modlog1: err?", err)
	rollbackOnErr(ed.Eout(err, "exec modlog #1 stmt"))
	_, err = modlogStmt2.Exec(deletedUserID, userid)
	fmt.Println("modlog2: err?", err)
	rollbackOnErr(ed.Eout(err, "exec modlog #2 stmt"))
	_, err = stmtReg.Exec(userid)
	rollbackOnErr(ed.Eout(err, "exec registration stmt"))
	_, err = stmtUsers.Exec(userid)
	rollbackOnErr(ed.Eout(err, "exec users stmt"))

	err = tx.Commit()
	ed.Check(err, "commit transaction")
	finalErr = nil
	return
}

func (d DB) AddRegistration(userid int, verificationLink string) error {
	ed := util.Describe("add registration")
	stmt := `INSERT INTO registrations (userid, host, link, time) VALUES (?, ?, ?, ?)`
	t := time.Now()
	u, err := url.Parse(verificationLink)
	if err = ed.Eout(err, "parse url"); err != nil {
		return err
	}
	_, err = d.Exec(stmt, userid, u.Host, verificationLink, t)
	if err = ed.Eout(err, "add registration"); err != nil {
		return err
	}
	return nil
}

func (d DB) AddModerationLog(actingid, recipientid, action int) error {
	ed := util.Describe("add moderation log")
	t := time.Now()
	// we have a recipient
	var err error
	if recipientid > 0 {
		stmt := `INSERT INTO moderation_log (actingid, recipientid, action, time) VALUES (?, ?, ?, ?)`
		_, err = d.Exec(stmt, actingid, recipientid, action, t)
		} else {
			// we are not listing a recipient
		stmt := `INSERT INTO moderation_log (actingid, action, time) VALUES (?, ?, ?)`
		_, err = d.Exec(stmt, actingid, action, t)
	}
	if err = ed.Eout(err, "exec prepared statement"); err != nil {
		return err
	}
	return nil
}

type ModerationEntry struct {
	ActingUsername, RecipientUsername string
	Action int
	Time time.Time
}
func (d DB) GetModerationLogs () []ModerationEntry {
	ed := util.Describe("moderation log")
	query := `SELECT uact.name, urecp.name, m.action, m.time 
	FROM moderation_LOG m 
	LEFT JOIN users uact ON uact.id = m.actingid
	LEFT JOIN users urecp ON urecp.id = m.recipientid
	ORDER BY time DESC`

	stmt, err := d.db.Prepare(query)
	ed.Check(err, "prep stmt")
	defer stmt.Close()

	rows, err := stmt.Query()
	util.Check(err, "run query")
	defer rows.Close()

	var entry ModerationEntry
	var logs []ModerationEntry
	for rows.Next() {
		var actingUsername, recipientUsername sql.NullString
		if err := rows.Scan(&actingUsername, &recipientUsername, &entry.Action, &entry.Time); err != nil {
			ed.Check(err, "scanning loop")
		}
		if actingUsername.Valid {
			entry.ActingUsername = actingUsername.String
		}
		if recipientUsername.Valid {
			entry.RecipientUsername = recipientUsername.String
		}
		logs = append(logs, entry)
	}
	return logs
}

func (d DB) ResetPassword(userid int) (string, error) {
	ed := util.Describe("reset password")
	exists, err := d.CheckUserExists(userid)
	if !exists {
		return "", errors.New(fmt.Sprintf("reset password: userid %d did not exist", userid))
	} else if err != nil {
		return "", fmt.Errorf("reset password encountered an error (%w)", err)
	}
	// generate new password for user and set it in the database
	newPassword := crypto.GeneratePassword()
	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return "", ed.Eout(err, "hash password")
	}
	d.UpdateUserPasswordHash(userid, passwordHash)
	return newPassword, nil
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

func (d DB) IsUserAdmin (userid int) (bool, error) {
	stmt := `SELECT 1 FROM admins WHERE id = ?`
	return d.existsQuery(stmt, userid)
}

func (d DB) GetAdmins() []User {
	ed := util.Describe("get admins")
	query := `SELECT u.name, a.id 
  FROM users u 
  INNER JOIN admins a ON u.id = a.id 
  ORDER BY u.name
  `
	stmt, err := d.db.Prepare(query)
	ed.Check(err, "prep stmt")
	defer stmt.Close()

	rows, err := stmt.Query()
	util.Check(err, "run query")
	defer rows.Close()

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

func (d DB) GetUsers(includeAdmin bool) []User {
	ed := util.Describe("get users")
	query := `SELECT u.name, u.id
  FROM users u 
	%s
  ORDER BY u.name
  `

	if includeAdmin {
		query = fmt.Sprintf(query, "") // do nothing
	} else {
		query = fmt.Sprintf(query, "WHERE u.id NOT IN (select id from admins)") // do nothing
	}

	stmt, err := d.db.Prepare(query)
	ed.Check(err, "prep stmt")
	defer stmt.Close()

	rows, err := stmt.Query()
	util.Check(err, "run query")
	defer rows.Close()

	var user User
	var users []User
	for rows.Next() {
		if err := rows.Scan(&user.Name, &user.ID); err != nil {
			ed.Check(err, "scanning loop")
		}
		users = append(users, user)
	}
	return users
}
