# voter

Vote bot for reddit.

## usage
You must create a configuration file and pass it in with the ``-c`` flag. See
the ``example.config.json`` file for the required fields. 

````
Usage of ./voter:
  -c string
    	config file
  -da
    	downvote everything found in scan
  -h	enables hammer mode
  -ua
    	upvote everything found in scan
  -v	enable verbose mode
````

To run voter based on the upvote/downvote configuration in your config file:

``./voter -c yourconfig.json``

Upvote every new comment and submission in the configured subreddit(s):

``./voter -c yourconfig.json -ua``

Downvote every new comment/submission in the configured subreddit(s):

``./voter -c yourconfig.json -da``

Enable "hammer" mode, where every new comment (score of 1) is upvoted by someone,
the bot upvotes it as well. If a new comment gets downvoted the bot downvotes it too.

``./voter -c yourconfig.json -h``