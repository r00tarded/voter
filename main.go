package main

import (
	"flag"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const userAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15`

var (
	oVerbose     bool
	oUpvoteAll   bool
	oDownvoteAll bool
	oConfig      string
	config       *Config
)

func init() {
	rand.Seed(time.Now().Unix())
	flag.StringVar(&oConfig, "c", "", "config file")
	flag.BoolVar(&oVerbose, "v", false, "enable verbose mode")
	flag.BoolVar(&oUpvoteAll, "ua", false, "upvote everything found in scan")
	flag.BoolVar(&oDownvoteAll, "da", false, "downvote everything found in scan")
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
		log.Println("[*] shutting down")
		db.Close()
		os.Exit(0)
	}()

	voter := NewVoter(db)

	for {
		voter.LoadComments()
		voter.LoadSubmissions()
		voter.Vote()
		log.Printf("[*] finished vote cycle, sleeping for %d seconds", config.Sleep)
		time.Sleep(time.Duration(config.Sleep) * time.Second)
	}
}
