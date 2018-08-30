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
	return d.checkForVote(redditAcct, downvoteKey(permalink))
}

//ContainsUpvote checks the db to see if an upvote exists for a post
func (d *Database) ContainsUpvote(redditAcct string, permalink string) bool {
	return d.checkForVote(redditAcct, upvoteKey(permalink))
}

//AddDownvote puts a record in the DB for the given reddit account and permalink.
func (d *Database) AddDownvote(redditAcct string, permalink string) {
	d.addVote(redditAcct, downvoteKey(permalink))
}

//AddUpvote adds a record of an upvote for a comment to the DB.
func (d *Database) AddUpvote(redditAcct string, permalink string) {
	d.addVote(redditAcct, upvoteKey(permalink))
}

//RemoveDownvote removes a record from the DB
func (d *Database) RemoveDownvote(redditAcct string, permalink string) {
	d.removeVote(redditAcct, downvoteKey(permalink))
}

//RemoveUpvote removes a record from the DB
func (d *Database) RemoveUpvote(redditAcct string, permalink string) {
	d.removeVote(redditAcct, upvoteKey(permalink))
}

//Close closes out database resources.
func (d *Database) Close() {
	d.db.Close()
}

//Reusable function for checking up/down votes
func (d *Database) checkForVote(redditAcct string, key string) bool {
	found := false
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(redditAcct))
		if err != nil {
			log.Printf("[x] error creating DB bucket: %s", err)
			return err
		}
		v := b.Get([]byte(key))
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

//Reusable function to add up/down votes to database
func (d *Database) addVote(redditAcct string, key string) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(redditAcct))
		if err != nil {
			log.Printf("[x] error creating bucket: %s", err)
			return err
		}
		err = b.Put([]byte(key), []byte("t"))
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

//Reusable function to remove votes from DB
func (d *Database) removeVote(redditAcct string, key string) {
	err := d.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(redditAcct))
		if err != nil {
			log.Printf("[x] error creating bucket: %s", err)
			return err
		}
		err = b.Delete([]byte(key))
		if err != nil {
			log.Printf("[x] error deleting value from DB: %s", err)
			return err
		}
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
}

//Generates a key usable in the database to associate a user to a downvoted post.
func downvoteKey(permalink string) string {
	return "downvote-" + permalink
}

//Generates a key usable in the database to associate a user to a downvoted post.
func upvoteKey(permalink string) string {
	return "upvote-" + permalink
}
