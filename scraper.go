package main

import (
	"log"
	"regexp"
	"time"

	"github.com/gocolly/colly"
)

func scrapeTheVerge(c1 chan<- mainStory, c2 chan<- feedStory) {
	c := colly.NewCollector()

	// Regexp - parse image url from html element in string format
	re, err := regexp.Compile(`http.*?[^"]*`)
	if err != nil {
		log.Printf("Error with regex: %s\n", err)
	}

	re2, err := regexp.Compile(`http.*?[^)]*`)
	if err != nil {
		log.Printf("Error with image regex: %s\n", err)
	}

	//Daily cover img & punch-line (or something like that shown besides date)
	c.OnHTML(".c-masthead", func(e *colly.HTMLElement) {
		coverImageHTML := e.ChildAttr(".c-masthead__main", "style")
		coverImage = re2.FindString(coverImageHTML)
		quote = e.ChildText(".c-masthead__main .l-wrapper .c-masthead__dateline .c-masthead__tagline a")
	})

	// Main/Top news
	c.OnHTML(".c-entry-box--compact--hero", func(e *colly.HTMLElement) {
		mainStory := mainStory{}
		mainStory.URL = e.ChildAttr("a", "href")

		//TODO: Parse all possible image-sizes and save in convenient way:
		//? mainStory.Image = e.ChildAttr("a .c-picture source", "srcset")
		imageHTML := e.ChildText("a .c-picture script")
		mainStory.Image = re.FindString(imageHTML)
		mainStory.Title = e.ChildText(".c-entry-box--compact__title a")
		mainStory.Author = e.ChildText(".c-byline__author-name")

		c1 <- mainStory
	})

	// Other news
	c.OnHTML(".c-compact-river__entry", func(e *colly.HTMLElement) {
		feedStory := feedStory{}

		LinkURL := e.ChildAttr(".c-entry-box--compact--article a", "href")
		if LinkURL != "" {
			feedStory.URL = e.ChildAttr(".c-entry-box--compact--article a", "href")

			imageHTML := e.ChildText(".c-entry-box--compact--article a .c-entry-box--compact__image noscript")
			feedStory.Image = re.FindString(imageHTML)

			feedStory.Title = e.ChildText(".c-entry-box--compact .c-entry-box--compact__title a")

			feedStory.Author = e.ChildText(".c-entry-box--compact .c-entry-box--compact__body .c-byline .c-byline-wrapper .c-byline__item:first-child a span")

			feedStory.Date = e.ChildAttr(".c-entry-box--compact .c-entry-box--compact__body .c-byline .c-byline-wrapper .c-byline__item:nth-child(2) time", "datetime")

			feedStory.Comments = e.ChildText(".c-entry-box--compact .c-entry-box--compact__body .c-byline .c-byline-wrapper .c-byline__item:nth-child(3) .c-entry-stat--words a .c-entry-stat__comment-data")

			c2 <- feedStory
		} else {
			return
		}
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println(err)
		close(c1)
		close(c2)
	})

	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.OnScraped(func(r *colly.Response) {
		log.Println("Finished with scraping:", r.Request.URL)
		close(c1)
		close(c2)
	})

	c.Visit("https://www.theverge.com")
}

func outputToMainNews(c <-chan mainStory) {
	for {
		mainStory := <-c
		mainStoryData = append(mainStoryData, &mainStory)
	}
}

func outputToFeedNews(c <-chan feedStory) {
	for {
		feedStory := <-c
		feedStoryData = append(feedStoryData, &feedStory)
	}
}

func startScraping() {
	for {
		log.Println("Fetching latest stories from The Verge")
		// Channels
		fromMainNews := make(chan mainStory, 10)
		fromFeedNews := make(chan feedStory, 10)
		toMainNews := make(chan mainStory, 10)
		toFeedNews := make(chan feedStory, 10)

		go scrapeTheVerge(fromMainNews, fromFeedNews)
		go outputToMainNews(toMainNews)
		go outputToFeedNews(toFeedNews)

		fromMainNewsOpen := true
		fromFeedNewsOpen := true

		for fromMainNewsOpen || fromFeedNewsOpen {
			select {
			case mainStory, open := <-fromMainNews:
				{
					if open {
						toMainNews <- mainStory
					} else {
						fromMainNewsOpen = false
					}
				}
			case feedStory, open := <-fromFeedNews:
				{
					if open {
						toFeedNews <- feedStory
					} else {
						fromFeedNewsOpen = false
					}
				}
			}
		}

		currentNews.Image = coverImage
		currentNews.Quote = quote
		currentNews.Main = mainStoryData
		currentNews.Feed = feedStoryData
		// Do not keep 'history' in slice of earlier data
		mainStoryData = nil
		feedStoryData = nil

		time.Sleep(1 * time.Hour)
	}
}
