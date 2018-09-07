package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	oVerbose     bool
	oUpvoteAll   bool
	oDownvoteAll bool
	oHammer      bool
	oConfig      string
	config       *Config
	startTime    = time.Now()
	tUpvotes     = 0 //total upvotes given out this session
	tDownvotes   = 0 //total downvotes given out this session
)

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&oConfig, "c", "", "config file")
	flag.BoolVar(&oVerbose, "v", false, "enable verbose mode")
	flag.BoolVar(&oUpvoteAll, "ua", false, "upvote everything found in scan")
	flag.BoolVar(&oDownvoteAll, "da", false, "downvote everything found in scan")
	flag.BoolVar(&oHammer, "h", false, "enables hammer mode")
	flag.Parse()

	if oConfig == "" {
		log.Println("[x] need config file")
		os.Exit(1)
	}

	var err error
	config, err = NewConfig(oConfig)
	if err != nil {
		log.Printf("[x] error reading config: %s", err)
		os.Exit(1)
	}
}

func main() {
	db := DBConn(config.DataDir)

	//gracefully handle shutdown
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGINT)
	go func() {
		<-c
		fmt.Println("")
		log.Println("[*] shutting down")
		log.Printf("[*] total upvotes given:   %d", tUpvotes)
		log.Printf("[*] total downvotes given: %d", tDownvotes)
		log.Printf("[*] total run time: %s", time.Since(startTime).String())
		db.Close()
		os.Exit(0)
	}()

	if oVerbose {
		log.Printf("[*] starting voter with configuration: %s", oConfig)
		log.Printf("[*] user agent: %s", config.UserAgent)
		log.Printf("[*] post limit: %d", config.Limit)
		log.Printf("[*] sleep timer: %d seconds", config.Sleep)
		log.Printf("[*] logging in with %d different accounts", len(config.RedditAccounts))
	}

	voter := NewVoter(db)

	for {
		voter.LoadComments()
		voter.LoadSubmissions()
		voter.Vote()
		log.Printf("[*] finished vote cycle, sleeping for %d seconds", config.Sleep)
		time.Sleep(time.Duration(config.Sleep) * time.Second)
	}
}
