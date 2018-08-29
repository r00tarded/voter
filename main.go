package main

import (
	"flag"
	"github.com/jzelinskie/geddit"
	"log"
	"os"
)

//todo optimize loops

const userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15`

// Please don't handle errors this way.
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

	for i := 0; i < len(config.Users); i++ {
		userName := config.Users[i]
		log.Printf("[-] looking for: %s\n", userName)
		for j := 0; j < len(config.RedditAccounts); j++ {
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

			log.Printf("[-] logged in with reddit account: %s\n", acct.User)

			var comments []*geddit.Comment
			comments, err = session.SubredditComments(config.Subreddit)
			if err != nil {
				log.Printf("[x] error reading comments: %s\n", err)
			}

			log.Printf("[-] got a total of %d comments\n", len(comments))

			for _, comment := range comments {
				permalink := comment.FullPermalink()
				//log.Printf("[-] comment: %s", permalink)
				if comment.Author == userName {
					session.Vote(comment, geddit.DownVote)
					log.Printf("[*] downvoted %s comment: %s\n", userName, permalink)
				}
			}

			options := geddit.ListingOptions{Limit: 50}

			var submissions []*geddit.Submission
			submissions, err = session.SubredditSubmissions(config.Subreddit, geddit.DefaultPopularity, options)

			log.Printf("[-] got a total of %d submissions", len(submissions))
			for _, submission := range submissions {
				if submission.Author == userName {
					session.Vote(submission, geddit.DownVote)
				}
			}
		}
	}
}
