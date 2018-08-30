package main

import (
	"flag"
	"github.com/jzelinskie/geddit"
	"log"
	"os"
	"math/rand"
	"time"
)

const userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15`

var (
	oVerbose bool
	oConfig string
)

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&oConfig, "c", "", "config file")
	flag.BoolVar(&oVerbose, "v", false, "enable verbose mode")
	flag.Parse()
}

func main() {
	var config *Config
	var err error

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

	dComments := make([]*geddit.Comment, 0)
	dSubmissions := make([]*geddit.Submission, 0)
	uComments := make([]*geddit.Comment, 0)
	uSubmissions := make([]*geddit.Submission, 0)

	log.Println("[*] looking through posts and comments...")

	for _, subreddit := range config.Subreddits {
		var comments []*geddit.Comment
		comments, err = mSession.SubredditComments(subreddit)
		if err != nil {
			log.Printf("[x] error reading comments: %s\n", err)
		}

		if oVerbose {
			log.Printf("[*] got a total of %d comments to check from /r/%s\n", len(comments), subreddit)
		}

		for _, comment := range comments {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if comment.Author == userName {
					dComments = append(dComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the downvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if comment.Author == userName {
					uComments = append(uComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the upvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
		}
	}

	options := geddit.ListingOptions{Limit: config.Limit}

	for _, subreddit := range config.Subreddits {
		var submissions []*geddit.Submission
		submissions, err = mSession.SubredditSubmissions(subreddit, geddit.DefaultPopularity, options)

		if oVerbose {
			log.Printf("[*] got a total of %d submissions from /r/%s to check\n", len(submissions), subreddit)
		}
		for _, submission := range submissions {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if submission.Author == userName {
					dSubmissions = append(dSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the downvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if submission.Author == userName {
					uSubmissions = append(uSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the upvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
		}
	}

	log.Println("[*] starting vote routines")

	massVoteComments(config, dComments, uComments, mSession, db)
	massVoteSubmissions(config, dSubmissions, uSubmissions, mSession, db)

	log.Println("[*] finished!")
	os.Exit(0)
}

func massVoteComments(config *Config, dComments []*geddit.Comment, uComments []*geddit.Comment, mSession *geddit.LoginSession, db *Database) {
	if oVerbose {
		log.Println("[*] running comment vote routine")
	}

	voteComments(config.RedditAccounts[0].User, mSession, dComments, uComments, db)

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

		if oVerbose {
			log.Printf("[*] logged in with reddit account: %s\n", acct.User)
		}

		voteComments(acct.User, session, dComments, uComments, db)
	}
}

func massVoteSubmissions(config *Config, dSubmissions []*geddit.Submission, uSubmissions []*geddit.Submission, mSession *geddit.LoginSession, db *Database) {
	if oVerbose {
		log.Println("[*] running submission vote routine")
	}

	voteSubmissions(config.RedditAccounts[0].User, mSession, dSubmissions, uSubmissions, db)

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

		if oVerbose {
			log.Printf("[*] logged in with reddit account: %s\n", acct.User)
		}

		voteSubmissions(acct.User, session, dSubmissions, uSubmissions, db)
	}
}

func voteSubmissions(user string, session *geddit.LoginSession, dSubmissions []*geddit.Submission, uSubmissions []*geddit.Submission, db *Database) {
	for _, submission := range dSubmissions {
		if !db.ContainsDownvote(user, submission.Permalink) {
			session.Vote(submission, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's submission: %s\n", user, submission.Author, submission.FullPermalink())
			db.AddDownvote(user, submission.Permalink)
		}
	}

	for _, submission := range uSubmissions {
		if !db.ContainsUpvote(user, submission.Permalink) {
			session.Vote(submission, geddit.UpVote)
			log.Printf("[+] %s upvoted %s's submission: %s\n", user, submission.Author, submission.FullPermalink())
		}
	}
}

func voteComments(user string, session *geddit.LoginSession, dComments []*geddit.Comment, uComments []*geddit.Comment, db *Database) {
	for _, comment := range dComments {
		if !db.ContainsDownvote(user, comment.Permalink) {
			session.Vote(comment, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's comment: %s\n", user, comment.Author, comment.FullPermalink())
			db.AddDownvote(user, comment.Permalink)
		}
	}

	for _, comment := range uComments {
		if !db.ContainsUpvote(user, comment.Permalink) {
			session.Vote(comment, geddit.UpVote)
			log.Printf("[+] %s upvoted %s's comment: %s\n", user, comment.Author, comment.FullPermalink())
			db.AddUpvote(user, comment.Permalink)
		}
	}
}
