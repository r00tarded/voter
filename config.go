package main

import (
	"encoding/json"
	"os"
)

type RedditAccount struct {
	User string `json: "user"`
	Pass string `json: "pass"`
}

type Config struct {
	DataDir        string           `json:"datadir"`         //directory containing database
	Subreddits     []string         `json: "subreddits"`     //list of subreddits to scan
	Limit          int              `json: "limit"`          //limit the number of posts to retrieve
	DownvoteUsers  []string         `json: "downvoteusers"`          //list of users to target for downvotes
	UpvoteUsers    []string `json: "upvoteusers"` //users to target for upvotes
	RedditAccounts []*RedditAccount `json: "redditAccounts"` //reddit accounts to log in and vote with
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
