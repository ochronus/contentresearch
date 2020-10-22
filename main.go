package main

import (
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/gocolly/colly/queue"
	"github.com/velebak/colly-sqlite3-storage/colly/sqlite3"
)

var questions map[string]int
var links map[string]string
var stopwords []string
var collyQueue *queue.Queue

func term2SearchURL(term string) string {
	return fmt.Sprintf("https://www.google.com/search?hl=en&gl=en&q=%s", url.QueryEscape(term))
}

func handleTerm(e *colly.HTMLElement, term string) {
	fmt.Println("handleTerm", term)
	_, ok := questions[term]
	if ok {
		questions[term]++
	} else {
		for _, stopword := range stopwords {
			fmt.Printf("Testing %s for %s\n", term, stopword)
			if strings.Contains(term, stopword) {
				return
			}
		}
		collyQueue.AddURL(term2SearchURL(term))
		questions[term] = 1
	}
}

func handleLink(title string, url string) {
	_, ok := links[url]
	if !ok {
		links[url] = title
	}
}

func main() {
	links = make(map[string]string)
	questions = make(map[string]int)
	stopwords = []string{"dress", "attire", "jeans", "wear", "early", "first", "last"}

	c := colly.NewCollector()
	storage := &sqlite3.Storage{
		Filename: "./colly.db",
	}
	extensions.RandomUserAgent(c)
	extensions.Referer(c)

	defer storage.Close()

	err := c.SetStorage(storage)
	if err != nil {
		panic(err)
	}

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	collyQueue, _ := queue.New(2, storage)

	// questions
	c.OnHTML("div.related-question-pair", func(e *colly.HTMLElement) {
		handleTerm(e, e.Text)
	})

	// actual links
	c.OnHTML("div.rc a h3", func(e *colly.HTMLElement) {
		link := ""
		for _, attr := range e.DOM.Parent().Nodes[0].Attr {
			if attr.Key == "href" {
				link = attr.Val
			}
		}
		handleLink(e.Text, link)
	})

	// related searches
	c.OnHTML("#botstuff a", func(e *colly.HTMLElement) {
		u, err := url.Parse(e.Attr("href"))
		if err != nil {
			fmt.Printf("Error parsing %s : %v\n", e.Attr("href"), err)
		} else {
			queryVals := u.Query()
			if queryVals.Get("q") != "" {
				handleTerm(e, queryVals.Get("q"))
			}
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnResponse(func(r *colly.Response) {
		log.Println(r.Request.URL, "\t", r.StatusCode)
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Println(r.Request.URL, "\t", r.StatusCode, "\nError:", err)
	})

	collyQueue.AddURL(term2SearchURL("tech interview"))

	err = collyQueue.Run(c)

	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	}
	fmt.Printf("%v\n", links)
	fmt.Printf("%v\n", questions)
}
