package main

import "log"

type mainStory struct {
	Title  string `json:"Title"`
	Author string `json:"Author"`
	URL    string `json:"URL"`
	Image  string `json:"Image"`
}

type feedStory struct {
	Title    string `json:"Title"`
	Author   string `json:"Author"`
	URL      string `json:"URL"`
	Date     string `json:"Date"`
	Image    string `json:"Image"`
	Comments string `json:"Comments"`
}

type data struct {
	Image string       `json:"BackgroundImg"`
	Quote string       `json:"Quote"`
	Main  []*mainStory `json:"MainNews"`
	Feed  []*feedStory `json:"FeedNews"`
}

var coverImage string
var quote string
var mainStoryData []*mainStory
var feedStoryData []*feedStory
var currentNews data

func main() {
	log.Println("Starting the server")
	go startScraping()
	runServer()
	log.Println("Server has stopped running")
}
