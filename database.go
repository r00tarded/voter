package main

import (
	"github.com/boltdb/bolt"
	"log"
	"os"
)

type Database struct {
	DataDir string
	db      *bolt.DB
}

type Record struct {
	DateCreated string
}

func DBConn(datadir string) *Database {
	db, err := bolt.Open(datadir+"/my.db", 0600, nil)
	if err != nil {
		log.Printf("[x] error opening DB: %s", err)
		os.Exit(1)
	}

	return &Database{db: db, DataDir: datadir}
}

//ContainsDownvote checks the database if there is a downvote for the given permalink on the given account.
func (d *Database) ContainsDownvote(redditAcct string, permalink string) bool {
	found := false
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(redditAcct))
		if err != nil {
			log.Printf("[x] error creating DB bucket: %s", err)
			return err
		}
		v := b.Get([]byte(downvoteKey(permalink)))
		if v != nil {
			found = true
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
	return found
}

//AddDownvote puts a record in the DB for the given reddit account and permalink.
func (d *Database) AddDownvote(redditAcct string, permalink string) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(redditAcct))
		if err != nil {
			log.Printf("[x] error creating bucket: %s", err)
			return err
		}
		err = b.Put([]byte(downvoteKey(permalink)), []byte("t"))
		if err != nil {
			log.Printf("[x] error adding value to DB: %s", err)
			return err
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
}

//Close closes out database resources.
func (d *Database) Close() {
	d.db.Close()
}

//Generates a key usable in the database to associate a user to a downvoted post.
func downvoteKey(permalink string) string {
	return "downvote-" + permalink
}
