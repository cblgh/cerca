package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"time"

	"cerca/logger"
	"cerca/util"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
}

func InitDB(filepath string) DB {
	if _, err := os.Stat(filepath); errors.Is(err, os.ErrNotExist) {
		file, err := os.Create(filepath)
		if err != nil {
			logger.Fatal("failed to initialize database: %v", err)
		}
		defer file.Close()
	}

	db, err := sql.Open("sqlite3", filepath)
	if err != nil {
		logger.Fatal("failed to open sqlite3 database at %s: %v", filepath, err)
	} else if db == nil {
		logger.Fatal("db is nil")
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
			logger.Fatal("failed to create database table with query %s: %v", query, err)
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
	// create the new thread in a transaction spanning two statements
	tx, err := d.db.BeginTx(context.Background(), &sql.TxOptions{}) // proper tx options?
	if err != nil {
		logger.Fatal("failed to start transaction: %v", err)
	}
	// first, create the new thread
	publish := time.Now()
	threadStmt := `INSERT INTO threads (title, publishtime, topicid, authorid) VALUES (?, ?, ?, ?)
  RETURNING id`
	replyStmt := `INSERT INTO posts (content, publishtime, threadid, authorid) VALUES (?, ?, ?, ?)`
	var threadid int
	err = tx.QueryRow(threadStmt, title, publish, topicid, authorid).Scan(&threadid)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("failed to add thread %s by %d in topic %d, rolling back: %v", title, authorid, topicid, err)
		return -1, err
	}
	// then add the content as the first reply to the thread
	_, err = tx.Exec(replyStmt, content, publish, threadid, authorid)
	if err != nil {
		_ = tx.Rollback()
		logger.Error("failed to add initial reply for thread %d, rolling back: %v", threadid, err)
		return -1, err
	}
	err = tx.Commit()
	if err != nil {
		logger.Fatal("failed to commit transaction: %v", err)
	}
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
	if err != nil {
		logger.Fatal("failed to get thread - preparing query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(threadid)
	if err != nil {
		logger.Fatal("failed to get thread - running query: %v", err)
	}
	defer rows.Close()

	var data Post
	var posts []Post
	for rows.Next() {
		if err := rows.Scan(&data.ThreadTitle, &data.Content, &data.Author, &data.Publish, &data.LastEdit); err != nil {
			logger.Fatal("failed to get data for thread %d: %v", threadid, err)
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
	if err != nil {
		logger.Fatal("failed to list threads - prepare query: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		logger.Fatal("failed to list threads - run query: %v", err)
	}
	defer rows.Close()

	var data Thread
	var threads []Thread
	for rows.Next() {
		if err := rows.Scan(&data.Title, &data.ID, &data.Author); err != nil {
			logger.Fatal("failed to list threads - scanning rows: %v", err)
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
	if err != nil {
		logger.Fatal("failed to add post to thread %d (author %d): %v", threadid, authorid, err)
	}
}

func (d DB) EditPost(content string, postid int) {
	stmt := `UPDATE posts set content = ?, lastedit = ? WHERE id = ?`
	edit := time.Now()
	_, err := d.Exec(stmt, content, edit, postid)
	if err != nil {
		logger.Fatal("failed to edit post %d: %v", postid, err)
	}
}

func (d DB) DeletePost(postid int) {
	stmt := `DELETE FROM posts WHERE id = ?`
	_, err := d.Exec(stmt, postid)
	if err != nil {
		logger.Fatal("failed to delete post %d: %v", postid, err)
	}
}

func (d DB) CreateTopic(title, description string) {
	stmt := `INSERT INTO topics (name, description) VALUES (?, ?)`
	_, err := d.Exec(stmt, title, description)
	if err != nil {
		logger.Fatal("failed to create topic %s: %v", title, err)
	}
}

func (d DB) UpdateTopicName(topicid int, newname string) {
	stmt := `UPDATE topics SET name = ? WHERE id = ?`
	_, err := d.Exec(stmt, newname, topicid)
	if err != nil {
		logger.Fatal("failed to change topic (%d) name to %s: %v", topicid, newname, err)
	}
}

func (d DB) UpdateTopicDescription(topicid int, newdesc string) {
	stmt := `UPDATE topics SET description = ? WHERE id = ?`
	_, err := d.Exec(stmt, newdesc, topicid)
	if err != nil {
		logger.Fatal("failed to change topic (%d) description to %s: %v", topicid, newdesc, err)
	}
}

func (d DB) DeleteTopic(topicid int) {
	stmt := `DELETE FROM topics WHERE id = ?`
	_, err := d.Exec(stmt, topicid)
	if err != nil {
		logger.Fatal("failed to delete topic: %d", topicid, err)
	}
}

func (d DB) CreateUser(name, hash string) (int, error) {
	stmt := `INSERT INTO users (name, passwordhash) VALUES (?, ?) RETURNING id`
	var userid int
	err := d.db.QueryRow(stmt, name, hash).Scan(&userid)
	if err != nil {
		return -1, fmt.Errorf("creating user %s: %w", name, err)
	}
	return userid, nil
}

func (d DB) GetPasswordHash(username string) (string, int, error) {
	stmt := `SELECT passwordhash, id FROM users where name = ?`
	var hash string
	var userid int
	err := d.db.QueryRow(stmt, username).Scan(&hash, &userid)
	if err != nil {
		return "", -1, fmt.Errorf("get password hash: %w", err)
	}
	return hash, userid, nil
}

func (d DB) existsQuery(substmt string, args ...interface{}) (bool, error) {
	stmt := fmt.Sprintf(`SELECT exists (%s)`, substmt)
	var exists bool
	err := d.db.QueryRow(stmt, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("exists %s: %w", substmt, err)
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
	if err != nil {
		logger.Fatal("failed to change user (%d) name to %s: %v", userid, newname, err)
	}
}

func (d DB) UpdateUserPasswordHash(userid int, newhash string) {
	stmt := `UPDATE users SET passwordhash = ? WHERE id = ?`
	_, err := d.Exec(stmt, newhash, userid)
	if err != nil {
		logger.Fatal("failed to change user (%d) description to %s: %v", userid, newhash, err)
	}
}

func (d DB) DeleteUser(userid int) {
	stmt := `DELETE FROM users WHERE id = ?`
	_, err := d.Exec(stmt, userid)
	if err != nil {
		logger.Fatal("failed to delete user %d: %v", userid, err)
	}
}

func (d DB) AddPubkey(userid int, pubkey string) error {
	stmt := `INSERT INTO pubkeys (pubkey, userid) VALUES (?, ?)`
	_, err := d.Exec(stmt, userid, pubkey)
	if err != nil {
		return fmt.Errorf("add public key - inserting record: %w", err)
	}
	return nil
}

func (d DB) AddRegistration(userid int, verificationLink string) error {
	stmt := `INSERT INTO registrations (userid, host, link, time) VALUES (?, ?, ?, ?)`
	t := time.Now()
	u, err := url.Parse(verificationLink)
	if err != nil {
		return fmt.Errorf("add registration - parse url: %w", err)
	}
	_, err = d.Exec(stmt, userid, u.Host, verificationLink, t)
	if err != nil {
		return fmt.Errorf("add registration - run query: %w", err)
	}
	return nil
}
