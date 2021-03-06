package main

import (
	"encoding/json"
	"os"
)

type RedditAccount struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type Config struct {
	UserAgent        string           `json:"useragent"`        //user agent for client
	DataDir          string           `json:"datadir"`          //directory containing database
	Subreddits       []string         `json:"subreddits"`       //list of subreddits to scan
	Limit            int              `json:"limit"`            //limit the number of posts to retrieve
	Ignores          []string         `json:"ignores"`          //user names to ignore when voting all posts
	DownvoteUsers    []string         `json:"downvoteusers"`    //list of users to target for downvotes
	UpvoteUsers      []string         `json:"upvoteusers"`      //users to target for upvotes
	RedditAccounts   []*RedditAccount `json:"redditaccounts"`   //reddit accounts to log in and vote with
	Sleep            int              `json:"sleep"`            //number of seconds to sleep between iterations
	UpvoteKeywords   []string         `json:"upvotekeywords"`   //list of keywords to upvote
	DownvoteKeywords []string         `json:"downvotekeywords"` //list of keywords to downvote
}

//NewConfig creates a new instance of Config.
func NewConfig(file string) (*Config, error) {
	var c Config
	if file != "" {
		cf, err := os.Open(file)
		json.NewDecoder(cf).Decode(&c)
		return &c, err
	}
	return &c, nil
}
