package main

import (
	"github.com/jzelinskie/geddit"
	"log"
	"os"
)

//Voter contains the logic for Reddit voting.
type Voter struct {
	db           *Database
	sessions     []*geddit.LoginSession
	uComments    []*geddit.Comment
	dComments    []*geddit.Comment
	uSubmissions []*geddit.Submission
	dSubmissions []*geddit.Submission
}

//NewVoter returns a configured instance of Voter.
func NewVoter(db *Database) *Voter {
	sessions := make([]*geddit.LoginSession, 0)
	dComments := make([]*geddit.Comment, 0)
	dSubmissions := make([]*geddit.Submission, 0)
	uComments := make([]*geddit.Comment, 0)
	uSubmissions := make([]*geddit.Submission, 0)

	//login to reddit with all accounts and store the sessions
	for i := 0; i < len(config.RedditAccounts); i++ {
		acct := config.RedditAccounts[i]

		session, err := geddit.NewLoginSession(
			acct.User,
			acct.Pass,
			userAgent,
		)
		if err != nil {
			log.Printf("[x] error logging in with account (%s): %s\n", acct.User, err)
			os.Exit(1)
		}

		log.Printf("[*] logged in with reddit account: %s\n", acct.User)

		sessions = append(sessions, session)
	}

	return &Voter{db: db, sessions: sessions, uComments: uComments, uSubmissions: uSubmissions, dComments: dComments, dSubmissions: dSubmissions}
}

//LoadComments primes the Voter with comments.
func (v *Voter) LoadComments() {
	log.Println("[*] loading comments...")
	var err error
	for _, subreddit := range config.Subreddits {
		var comments []*geddit.Comment
		comments, err = v.sessions[0].SubredditComments(subreddit)
		if err != nil {
			log.Printf("[x] error reading comments: %s\n", err)
		}

		if oUpvoteAll {
			log.Printf("[*] adding all comments from %s to upvote queue", subreddit)
			v.uComments = append(v.uComments, comments...)
			continue
		} else if oDownvoteAll {
			log.Printf("[*] adding all comments from %s to downvote queue", subreddit)
			v.dComments = append(v.dComments, comments...)
			continue
		}

		if oVerbose {
			log.Printf("[*] got a total of %d comments to check from /r/%s\n", len(comments), subreddit)
		}

		for _, comment := range comments {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if comment.Author == userName {
					v.dComments = append(v.dComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the downvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if comment.Author == userName {
					v.uComments = append(v.uComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the upvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
		}
	}
}

//LoadSubmissions primes the voter with submissions.
func (v *Voter) LoadSubmissions() {
	log.Println("[*] loading submissions...")
	options := geddit.ListingOptions{Limit: config.Limit}
	for _, subreddit := range config.Subreddits {
		submissions, err := v.sessions[0].SubredditSubmissions(subreddit, geddit.DefaultPopularity, options)
		if err != nil {
			log.Printf("[x] error reading submissions: %s\n", err)
		}

		if oUpvoteAll {
			log.Printf("[*] adding all submissions from %s to upvote queue\n", subreddit)
			v.uSubmissions = append(v.uSubmissions, submissions...)
			continue
		} else if oDownvoteAll {
			log.Printf("[*] adding all submissions from %s to downvote queue\n", subreddit)
			v.dSubmissions = append(v.dSubmissions, submissions...)
			continue
		}

		if oVerbose {
			log.Printf("[*] got a total of %d submissions from /r/%s to check\n", len(submissions), subreddit)
		}
		for _, submission := range submissions {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if submission.Author == userName {
					v.dSubmissions = append(v.dSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the downvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if submission.Author == userName {
					v.uSubmissions = append(v.uSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the upvote examine queue\n", subreddit, userName)
					}
					break
				}
			}
		}
	}
}

//Vote executes voting routines after the Voter has been primed.
func (v *Voter) Vote() {
	log.Println("[*] running voting routines...")
	massVoteComments(v.dComments, v.uComments, v.sessions, v.db)
	massVoteSubmissions(v.dSubmissions, v.uSubmissions, v.sessions, v.db)
}

func massVoteComments(dComments []*geddit.Comment, uComments []*geddit.Comment, sessions []*geddit.LoginSession, db *Database) {
	if oVerbose {
		log.Println("[*] running comment vote routine")
	}

	for i := 0; i < len(sessions); i++ {
		voteComments(config.RedditAccounts[i].User, sessions[i], dComments, uComments, db)
	}
}

func massVoteSubmissions(dSubmissions []*geddit.Submission, uSubmissions []*geddit.Submission, sessions []*geddit.LoginSession, db *Database) {
	if oVerbose {
		log.Println("[*] running submission vote routine")
	}

	for i := 0; i < len(sessions); i++ {
		voteSubmissions(config.RedditAccounts[i].User, sessions[i], dSubmissions, uSubmissions, db)
	}
}

func voteSubmissions(user string, session *geddit.LoginSession, dSubmissions []*geddit.Submission, uSubmissions []*geddit.Submission, db *Database) {
	for _, submission := range dSubmissions {
		if isIgnored(submission.Author) {
			continue
		}
		if !db.ContainsDownvote(user, submission.Permalink) {
			session.Vote(submission, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's submission: %s\n", user, submission.Author, submission.FullPermalink())
			db.AddDownvote(user, submission.Permalink)
			db.RemoveUpvote(user, submission.Permalink)
		}
	}

	for _, submission := range uSubmissions {
		if isIgnored(submission.Author) {
			continue
		}
		if !db.ContainsUpvote(user, submission.Permalink) {
			session.Vote(submission, geddit.UpVote)
			log.Printf("[+] %s upvoted %s's submission: %s\n", user, submission.Author, submission.FullPermalink())
			db.AddUpvote(user, submission.Permalink)
			db.RemoveDownvote(user, submission.Permalink)
		}
	}
}

func voteComments(user string, session *geddit.LoginSession, dComments []*geddit.Comment, uComments []*geddit.Comment, db *Database) {
	for _, comment := range dComments {
		if isIgnored(comment.Author) {
			continue
		}
		if !db.ContainsDownvote(user, comment.Permalink) {
			session.Vote(comment, geddit.DownVote)
			log.Printf("[-] %s downvoted %s's comment: %s\n", user, comment.Author, comment.FullPermalink())
			db.AddDownvote(user, comment.Permalink)
			db.RemoveUpvote(user, comment.Permalink)
		}
	}

	for _, comment := range uComments {
		if isIgnored(comment.Author) {
			continue
		}
		if !db.ContainsUpvote(user, comment.Permalink) {
			session.Vote(comment, geddit.UpVote)
			log.Printf("[+] %s upvoted %s's comment: %s\n", user, comment.Author, comment.FullPermalink())
			db.AddUpvote(user, comment.Permalink)
			db.RemoveDownvote(user, comment.Permalink)
		}
	}
}

//Check ignored users list to see if this user is in it
func isIgnored(user string) bool {
	for _, ign := range config.Ignores {
		if user == ign {
			return true
		}
	}
	return false
}
