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
	Subreddit      string           `json: "subreddit"`
	Users          []string         `json: "users"`
	RedditAccounts []*RedditAccount `json: "redditAccounts"`
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
