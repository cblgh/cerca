package database

import (
	"context"
	"database/sql"
	"fmt"
	"cerca/util"
	"html/template"
	"log"
	"net/url"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func InitDB(filepath string) DB {
	db, err := sql.Open("sqlite3", filepath)
	util.Check(err, "opening sqlite3 database at %s", filepath)
	if db == nil {
		log.Fatalln("db is nil")
	}
	createTables(db)
	return DB{db}
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
  CREATE TABLE IF NOT EXISTS pubkeys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    pubkey TEXT NOT NULL UNIQUE,
    userid integer NOT NULL UNIQUE,
    FOREIGN KEY (userid) REFERENCES users(id)
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
	ThreadTitle string
	Content     template.HTML
	Author      string
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
  SELECT t.title, content, u.name, p.publishtime, p.lastedit
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
		if err := rows.Scan(&data.ThreadTitle, &data.Content, &data.Author, &data.Publish, &data.LastEdit); err != nil {
			log.Fatalln(util.Eout(err, "get data for thread %d", threadid))
		}
		posts = append(posts, data)
	}
	return posts
}

type Thread struct {
	Title   string
	Author  string
	Slug    string
	ID      int
	Publish time.Time
}

// get a list of threads
func (d DB) ListThreads() []Thread {
	query := `
  SELECT t.title, t.id, u.name FROM threads t 
  INNER JOIN users u on u.id = t.authorid
  ORDER BY t.publishtime DESC
  `
	stmt, err := d.db.Prepare(query)
	util.Check(err, "list threads: prepare query")
	defer stmt.Close()

	rows, err := stmt.Query()
	util.Check(err, "list threads: query")
	defer rows.Close()

	var data Thread
	var threads []Thread
	for rows.Next() {
		if err := rows.Scan(&data.Title, &data.ID, &data.Author); err != nil {
			log.Fatalln(util.Eout(err, "list threads: read in data via scan"))
		}
		data.Slug = fmt.Sprintf("%d/%s/", data.ID, util.SanitizeURL(data.Title))
		threads = append(threads, data)
	}
	return threads
}

func (d DB) AddPost(content string, threadid, authorid int) {
	stmt := `INSERT INTO posts (content, publishtime, threadid, authorid) VALUES (?, ?, ?, ?)`
	publish := time.Now()
	_, err := d.Exec(stmt, content, publish, threadid, authorid)
	util.Check(err, "add post to thread %d (author %d)", threadid, authorid)
}

func (d DB) EditPost(content string, postid int) {
	stmt := `UPDATE posts set content = ?, lastedit = ? WHERE id = ?`
	edit := time.Now()
	_, err := d.Exec(stmt, content, edit, postid)
	util.Check(err, "edit post %d", postid)
}

func (d DB) DeletePost(postid int) {
	stmt := `DELETE FROM posts WHERE id = ?`
	_, err := d.Exec(stmt, postid)
	util.Check(err, "deleting post %d", postid)
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

func (d DB) DeleteUser(userid int) {
	stmt := `DELETE FROM users WHERE id = ?`
	_, err := d.Exec(stmt, userid)
	util.Check(err, "deleting user %d", userid)
}

func (d DB) AddPubkey(userid int, pubkey string) error {
	ed := util.Describe("add pubkey")
	stmt := `INSERT INTO pubkeys (pubkey, userid) VALUES (?, ?)`
	_, err := d.Exec(stmt, userid, pubkey)
	if err = ed.Eout(err, "inserting record"); err != nil {
		return err
	}
	return nil
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
