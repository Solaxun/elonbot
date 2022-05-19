package main

import "github.com/dghubble/go-twitter/twitter"

func testElonDodgeTweet() *twitter.Tweet {
	return &twitter.Tweet{
		Text:          "blah blah blah blah.... Doge",
		ExtendedTweet: &twitter.ExtendedTweet{FullText: "blah blah Dogecoin to the moon!!! blah blah"}}
}
