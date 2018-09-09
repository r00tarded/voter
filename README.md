# voter

Vote bot for reddit.

## build

````
$ go get github.com/jpclark/voter...
$ cd $GOPATH/src/github.com/jpclark/voter
$ go build
````

The first time you run the bot it will create a database file in the directory
named by the "datadir" configuration variable. You will need to create this directory
in the voter working directory (or wherever you wish to store it), e.g. 

````
$ mkdir $GOPATH/src/github.com/jpclark/voter/data
```` 

*note: this database tracks which posts have been voted on to prevent wasting vote requests* 

## usage
You must create a configuration file and pass it in with the ``-c`` flag. See
the ``example.config.json`` file and documentation below for the required fields.

````
$ ./voter -?
Usage of ./voter:
  -c string
    	config file
  -da
    	downvote everything found in configured subreddits
  -h	enables hammer mode
  -ua
    	upvote everything found in configured subreddits
  -v	enable verbose mode
  -vu
    	upvote/downvote everything by users in config
````

###default mode

This runs the voter based on the subreddit and user configuration in your config file.
The voter will upvote/downvote users in the user lists, but only if they are found in
the configured subreddits.

``$ ./voter -c yourconfig.json``

###upvote everything in given subreddit

Upvote every new comment and submission in the configured subreddit(s). The **limit**
configuration property is important to configure if using this mode.

``$ ./voter -c yourconfig.json -ua``

###downvote everything in given subreddit

Downvote every new comment/submission in the configured subreddit(s).

``$ ./voter -c yourconfig.json -da``

###vote on everything posted by given user(s)

Vote on all comments and submissions from user names in config (regardless of subreddit).

``$ ./voter -c yourconfig.json -vu``

###smash votes with a hammer

Enable "hammer" mode based on subreddit, where every new comment (score of 1) is upvoted by someone,
the bot upvotes it as well. If a new comment gets downvoted the bot downvotes it too:

``$ ./voter -c yourconfig.json -h``

## configuration

### config fields

**useragent** - http user agent to send to reddit from client

**datadir** - path to folder where you want to store database file(s)

**limit** - limit the number of posts to retrieve each iteration. tune this to popularity of the target subreddit(s)

**sleep** - number of seconds to rest between iterations of voting loop

**ignores** - array containing user names to leave out of voting

**subreddits** - array containing subreddit names to target

**downvoteusers** - array containing user names to downvote

**upvoteusers** -- array containing user names to upvote

**redditaccounts** - array containing logins for the reddit accounts you wish to vote with

### example config file

````
{  
   "useragent":"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_13_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/11.1.2 Safari/605.1.15",
   "datadir":"data",
   "limit":100,
   "sleep":30,
   "ignores":[  
      "user_to_ignore",
      "my_own_user_name",
      "another_user"
   ],
   "subreddits":[  
      "list",
      "of",
      "subreddit_names"
   ],
   "downvoteusers":[  
      "users_to_downvote", 
      "another_user"
   ],
   "upvoteusers":[  
      "users_to_upvote",
      "another_user"
   ],
   "redditaccounts":[  
      {  
         "user":"your_user",
         "pass":"your_pass"
      },
      {  
         "user":"another_user",
         "pass":"another_pass"
      },
      {  
         "user":"yet_another",
         "pass":"yet_another_pass"
      }
   ]
}
````