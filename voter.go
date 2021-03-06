package main

import (
	"github.com/jzelinskie/geddit"
	"log"
	"os"
	"regexp"
	"strings"
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
			config.UserAgent,
		)
		if err != nil {
			log.Printf("[x] error logging in with account (%s): %s\n", acct.User, err)
			os.Exit(1)
		}

		log.Printf("[*] logged in with reddit account: %s\n", acct.User)

		sessions = append(sessions, session)
	}

	return &Voter{db: db, sessions: sessions, uComments: uComments, uSubmissions: uSubmissions,
		dComments: dComments, dSubmissions: dSubmissions}
}

//LoadComments primes the Voter with comments.
func (v *Voter) LoadComments() {
	log.Println("[*] loading comments...")
	v.uComments = nil
	v.dComments = nil
	var err error
	options := geddit.ListingOptions{Limit: config.Limit}
	//handle vote-user mode
	if oVoteUser {
		for _, userName := range config.DownvoteUsers {
			comments, err := v.sessions[0].RedditorComments(userName, options)
			if err != nil {
				log.Printf("[x] error reading comments for user %s: %s\n", userName, err)
				continue
			}
			for _, comment := range comments {
				v.dComments = append(v.dComments, comment)
				if oVerbose {
					log.Printf("[*] added comment in /r/%s from %s to the downvote queue\n",
						comment.Subreddit, userName)
				}
			}
		}

		for _, userName := range config.UpvoteUsers {
			comments, err := v.sessions[0].RedditorComments(userName, options)
			if err != nil {
				log.Printf("[x] error reading comments for user %s: %s\n", userName, err)
				continue
			}
			for _, comment := range comments {
				v.uComments = append(v.uComments, comment)
				if oVerbose {
					log.Printf("[*] added comment in /r/%s from %s to the upvote queue\n",
						comment.Subreddit, userName)
				}
			}
		}

		return //skip the subreddit mode
	}
	//handle subreddit-specific mode
	for _, subreddit := range config.Subreddits {
		var comments []*geddit.Comment
		comments, err = v.sessions[0].SubredditComments(subreddit, options)
		if err != nil {
			log.Printf("[x] error reading comments for /r/%s: %s\n", subreddit, err)
			continue
		}

		//check for keyword mode
		if oKeyword {
			for _, comment := range comments {
				if containsKeywords(comment.Body, config.UpvoteKeywords) {
					if oVerbose {
						log.Println("[*] adding comment based on keyword to upvote queue")
					}
					v.uComments = append(v.uComments, comment)
				} else if containsKeywords(comment.Body, config.DownvoteKeywords) {
					if oVerbose {
						log.Println("[*] adding comment based on keyword to downvote queue")
					}
					v.dComments = append(v.dComments, comment)
				}
			}
			continue //continue to next subreddit
		}

		//check for up/downvote all flags
		if oUpvoteAll {
			log.Printf("[*] adding all comments from %s to upvote queue", subreddit)
			v.uComments = append(v.uComments, comments...)
			continue
		} else if oDownvoteAll {
			log.Printf("[*] adding all comments from %s to downvote queue", subreddit)
			v.dComments = append(v.dComments, comments...)
			continue
		}

		//handle hammer mode
		if oHammer {
			for _, comment := range comments {
				a := len(config.RedditAccounts)
				if comment.Score < 1 && int(comment.Score) > -a/2 {
					v.dComments = append(v.dComments, comment)
					if oVerbose {
						log.Printf("[*] added comment with score %.0f in /r/%s from %s to the downvote queue\n",
							comment.Score, subreddit, comment.Author)
					}
				} else if comment.Score > 1 && int(comment.Score) < a/2 {
					v.uComments = append(v.uComments, comment)
					if oVerbose {
						log.Printf("[*] added comment with score %.0f in /r/%s from %s to the upvote queue\n",
							comment.Score, subreddit, comment.Author)
					}
				}
			}
			continue
		}

		//default config driven behavior
		if oVerbose {
			log.Printf("[*] got a total of %d comments to check from /r/%s\n", len(comments), subreddit)
		}

		for _, comment := range comments {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if strings.EqualFold(comment.Author, userName) {
					v.dComments = append(v.dComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the downvote examine queue\n",
							subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if strings.EqualFold(comment.Author, userName) {
					v.uComments = append(v.uComments, comment)
					if oVerbose {
						log.Printf("[*] added comment in /r/%s from %s to the upvote examine queue\n",
							subreddit, userName)
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
	v.uSubmissions = nil
	v.dSubmissions = nil
	options := geddit.ListingOptions{Limit: config.Limit}
	//handle vote-user mode
	if oVoteUser {
		for _, userName := range config.DownvoteUsers {
			submissions, err := v.sessions[0].RedditorSubmissions(userName, options)
			if err != nil {
				log.Printf("[x] error reading submissions for user %s: %s\n", userName, err)
				continue
			}
			for _, submission := range submissions {
				v.dSubmissions = append(v.dSubmissions, submission)
				if oVerbose {
					log.Printf("[*] added submission in /r/%s from %s to the downvote queue\n",
						submission.Subreddit, userName)
				}
			}
		}

		for _, userName := range config.UpvoteUsers {
			submissions, err := v.sessions[0].RedditorSubmissions(userName, options)
			if err != nil {
				log.Printf("[x] error reading submissions for user %s: %s\n", userName, err)
				continue
			}
			for _, submission := range submissions {
				v.uSubmissions = append(v.uSubmissions, submission)
				if oVerbose {
					log.Printf("[*] added submission in /r/%s from %s to the upvote queue\n",
						submission.Subreddit, userName)
				}
			}
		}

		return //skip the subreddit mode
	}
	//handle subreddit-driven mode
	for _, subreddit := range config.Subreddits {
		submissions, err := v.sessions[0].SubredditSubmissions(subreddit, geddit.DefaultPopularity, options)
		if err != nil {
			log.Printf("[x] error reading submissions: %s\n", err)
		}

		//check for keyword mode
		if oKeyword {
			for _, submission := range submissions {
				if containsKeywords(submission.Selftext, config.UpvoteKeywords) {
					if oVerbose {
						log.Println("[*] adding submission based on keyword to upvote queue")
					}
					v.uSubmissions = append(v.uSubmissions, submission)
				} else if containsKeywords(submission.Selftext, config.DownvoteKeywords) {
					if oVerbose {
						log.Println("[*] adding submission based on keyword to downvote queue")
					}
					v.dSubmissions = append(v.dSubmissions, submission)
				}
			}
			continue //continue to next subreddit
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

		//handle hammer mode
		if oHammer {
			for _, submission := range submissions {
				a := len(config.RedditAccounts)
				if submission.Score == 0 && submission.Downs > 2 {
					v.dSubmissions = append(v.dSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission with score %d in /r/%s from %s to the downvote queue\n",
							submission.Score, subreddit, submission.Author)
					}
				} else if submission.Score > 2 && submission.Score < a/2 {
					v.uSubmissions = append(v.uSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission with score %d in /r/%s from %s to the upvote queue\n",
							submission.Score, subreddit, submission.Author)
					}
				}
			}
			continue
		}

		if oVerbose {
			log.Printf("[*] got a total of %d submissions from /r/%s to check\n", len(submissions), subreddit)
		}
		for _, submission := range submissions {
			for i := 0; i < len(config.DownvoteUsers); i++ {
				userName := config.DownvoteUsers[i]
				if strings.EqualFold(submission.Author, userName) {
					v.dSubmissions = append(v.dSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the downvote examine queue\n",
							subreddit, userName)
					}
					break
				}
			}
			for i := 0; i < len(config.UpvoteUsers); i++ {
				userName := config.UpvoteUsers[i]
				if strings.EqualFold(submission.Author, userName) {
					v.uSubmissions = append(v.uSubmissions, submission)
					if oVerbose {
						log.Printf("[*] added submission in /r/%s from %s to the upvote examine queue\n",
							subreddit, userName)
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
		downvoteSubmission(user, session, submission, db)
	}

	for _, submission := range uSubmissions {
		if isIgnored(submission.Author) {
			continue
		}
		upvoteSubmission(user, session, submission, db)
	}
}

func voteComments(user string, session *geddit.LoginSession, dComments []*geddit.Comment, uComments []*geddit.Comment, db *Database) {
	for _, comment := range dComments {
		if isIgnored(comment.Author) {
			continue
		}
		downvoteComment(user, session, comment, db)
	}

	for _, comment := range uComments {
		if isIgnored(comment.Author) {
			continue
		}
		upvoteComment(user, session, comment, db)
	}
}

func upvoteSubmission(acctName string, session *geddit.LoginSession, submission *geddit.Submission, db *Database) {
	if !db.ContainsUpvote(acctName, submission.Permalink) {
		session.Vote(submission, geddit.UpVote)
		log.Printf("[+] %s upvoted %s's submission in %s: %s\n",
			acctName, submission.Author, submission.Subreddit, submission.FullID)
		db.AddUpvote(acctName, submission.Permalink)
		db.RemoveDownvote(acctName, submission.Permalink)
		tUpvotes++
	}
}

func downvoteSubmission(acctName string, session *geddit.LoginSession, submission *geddit.Submission, db *Database) {
	if !db.ContainsDownvote(acctName, submission.Permalink) {
		session.Vote(submission, geddit.DownVote)
		log.Printf("[-] %s downvoted %s's submission in %s: %s\n",
			acctName, submission.Author, submission.Subreddit, submission.FullID)
		db.AddDownvote(acctName, submission.Permalink)
		db.RemoveUpvote(acctName, submission.Permalink)
		tDownvotes++
	}
}

func upvoteComment(acctName string, session *geddit.LoginSession, comment *geddit.Comment, db *Database) {
	if !db.ContainsUpvote(acctName, comment.Permalink) {
		session.Vote(comment, geddit.UpVote)
		log.Printf("[+] %s upvoted %s's comment in %s: %s\n",
			acctName, comment.Author, comment.Subreddit, comment.FullID)
		db.AddUpvote(acctName, comment.Permalink)
		db.RemoveDownvote(acctName, comment.Permalink)
		tUpvotes++
	}
}

func downvoteComment(acctName string, session *geddit.LoginSession, comment *geddit.Comment, db *Database) {
	if !db.ContainsDownvote(acctName, comment.Permalink) {
		session.Vote(comment, geddit.DownVote)
		log.Printf("[-] %s downvoted %s's comment in %s: %s\n",
			acctName, comment.Author, comment.Subreddit, comment.FullID)
		db.AddDownvote(acctName, comment.Permalink)
		db.RemoveUpvote(acctName, comment.Permalink)
		tDownvotes++
	}
}

//Check ignored users list to see if this user is in it
func isIgnored(user string) bool {
	for _, ign := range config.Ignores {
		if strings.EqualFold(user, ign) {
			return true
		}
	}
	return false
}

//Check the given text for list of keywords
func containsKeywords(text string, keywords []string) bool {
	if text == "" {
		return false
	}
	t := strings.ToUpper(text)
	for _, keyword := range keywords {
		k := strings.ToUpper(keyword)
		matched, err := regexp.MatchString(".*"+k+".*", t)
		if err != nil {
			log.Printf("[x] error searching for keyword %s: %s\n", keyword, err)
		}
		if matched {
			return true
		}
	}
	return false
}
