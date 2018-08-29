package main

import (
	"flag"
	"github.com/jzelinskie/geddit"
	"log"
	"os"
		)

const userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15`

func main() {
	var oConfig string
	var config *Config
	var err error

	flag.StringVar(&oConfig, "c", "", "config file")
	flag.Parse()

	if oConfig == "" {
		log.Println("[x] need config file")
		os.Exit(1)
	}

	config, err = NewConfig(oConfig)
	if err != nil {
		log.Printf("[x] error reading config: %s", err)
		os.Exit(1)
	}

	db := DBConn(config.DataDir)
	defer db.Close()

	mAcct := config.RedditAccounts[0]
	mSession, err := geddit.NewLoginSession(
		mAcct.User,
		mAcct.Pass,
		userAgent,
	)
	if err != nil {
		log.Printf("[x] error logging in with account (%s): %s\n", mAcct.User, err)
		os.Exit(1)
	}

	aComments := make([]*geddit.Comment, 0)
	aSubmissions := make([]*geddit.Submission, 0)

	var comments []*geddit.Comment
	comments, err = mSession.SubredditComments(config.Subreddit)
	if err != nil {
		log.Printf("[x] error reading comments: %s\n", err)
	}

	log.Printf("[*] got a total of %d comments to examine\n", len(comments))

	for _, comment := range comments {
		for i := 0; i < len(config.Users); i++ {
			userName := config.Users[i]
			if comment.Author == userName {
				aComments = append(aComments, comment)
				log.Printf("[*] added comment from %s to the queue\n", userName)
				break
			}
		}
	}

	options := geddit.ListingOptions{Limit: 30}

	var submissions []*geddit.Submission
	submissions, err = mSession.SubredditSubmissions(config.Subreddit, geddit.DefaultPopularity, options)

	log.Printf("[*] got a total of %d submissions to examine\n", len(submissions))
	for _, submission := range submissions {
		for i := 0; i < len(config.Users); i++ {
			userName := config.Users[i]
			if submission.Author == userName {
				aSubmissions = append(aSubmissions, submission)
				log.Printf("[*] added submission from %s to the queue\n", userName)
				break
			}
		}
	}

	massDownvoteComments(config, aComments, mSession, db)
	massDownvoteSubmissions(config, aSubmissions, mSession, db)

	log.Println("[*] finished!")
	os.Exit(0)
}

func massDownvoteComments(config *Config, comments []*geddit.Comment, mSession *geddit.LoginSession, db *Database) {
	downvoteComments(config.RedditAccounts[0].User, mSession, comments, db)

	for j := 1; j < len(config.RedditAccounts); j++ {
		acct := config.RedditAccounts[j]

		// Login to reddit
		session, err := geddit.NewLoginSession(
			acct.User,
			acct.Pass,
			userAgent,
		)
		if err != nil {
			log.Printf("[x] error logging in with account (%s): %s\n", acct.User, err)
			continue
		}

		log.Printf("[*] logged in with reddit account: %s\n", acct.User)

		downvoteComments(acct.User, session, comments, db)
	}
}

func massDownvoteSubmissions(config *Config, submissions []*geddit.Submission, mSession *geddit.LoginSession, db *Database) {
	downvoteSubmissions(config.RedditAccounts[0].User, mSession, submissions, db)

	for j := 1; j < len(config.RedditAccounts); j++ {
		acct := config.RedditAccounts[j]

		// Login to reddit
		session, err := geddit.NewLoginSession(
			acct.User,
			acct.Pass,
			userAgent,
		)
		if err != nil {
			log.Printf("[x] error logging in with account (%s): %s\n", acct.User, err)
			continue
		}

		log.Printf("[*] logged in with reddit account: %s\n", acct.User)

		downvoteSubmissions(acct.User, session, submissions, db)
	}
}

func downvoteSubmissions(user string, session *geddit.LoginSession, submissions []*geddit.Submission, db *Database) {
	for _, submission := range submissions {
		if !db.ContainsDownvote(user, submission.Permalink) {
			session.Vote(submission, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's submission: %s\n", user, submission.Author, submission.FullPermalink())
			db.AddDownvote(user, submission.Permalink)
		}
	}
}

func downvoteComments(user string, session *geddit.LoginSession, comments []*geddit.Comment, db *Database) {
	for _, comment := range comments {
		if !db.ContainsDownvote(user, comment.Permalink) {
			session.Vote(comment, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's comment: %s\n", user, comment.Author, comment.FullPermalink())
			db.AddDownvote(user, comment.Permalink)
		}
	}
}
