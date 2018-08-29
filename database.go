package main

import (
	"github.com/nanobox-io/golang-scribble"
	"fmt"
	"os"
	"time"
)

type Database struct {
	DataDir string
	driver *scribble.Driver
}

type Record struct {
	DateCreated string
}

func DBConn(datadir string) (*Database) {
	// connect to local db
	// a new scribble driver, providing the directory where it will be writing to,
	// and a qualified logger if desired
	db, err := scribble.New(datadir, nil)
	if err != nil {
		fmt.Printf("Error opening DB: %s", err)
		os.Exit(1)
	}

	return &Database{driver: db, DataDir: datadir}
}

//ContainsDownvote checks the database if there is a downvote for the given permalink on the given account.
func (d *Database) ContainsDownvote(redditAcct string, permalink string) bool {
	r := &Record{DateCreated: ""}
	d.driver.Read(redditAcct, downvoteKey(permalink), r)
	if r.DateCreated != "" {
		return true
	}
	return false
}

//AddDownvote puts a record in the DB for the given reddit account and permalink.
func (d *Database) AddDownvote(redditAcct string, permalink string) {
	d.driver.Write(redditAcct, downvoteKey(permalink), Record{DateCreated: time.Now().String()})
}

//Generates a key usable in the database to associate a user to a downvoted post.
func downvoteKey(permalink string) string {
	return "downvote:" + permalink
}