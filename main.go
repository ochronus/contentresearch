package main

import (
	"fmt"
	"net/url"

	"github.com/gocolly/colly"
)

func main() {
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"),
	)

	// questions
	c.OnHTML("div.related-question-pair", func(e *colly.HTMLElement) {
		fmt.Printf("QUESTION: %v\n", e.Text)
	})

	// actual links
	c.OnHTML("div.rc a h3", func(e *colly.HTMLElement) {
		link := ""
		for _, attr := range e.DOM.Parent().Nodes[0].Attr {
			if attr.Key == "href" {
				link = attr.Val
			}
		}
		fmt.Printf("RESULT: %s (%s)\n", e.Text, link)
	})

	// related searches
	c.OnHTML("#botstuff a", func(e *colly.HTMLElement) {
		u, err := url.Parse(e.Attr("href"))
		if err != nil {
			fmt.Printf("Error parsing %s : %v\n", e.Attr("href"), err)
		} else {
			queryVals := u.Query()
			if queryVals.Get("q") != "" {
				fmt.Printf("RELATED SEARCH: %v\n", queryVals.Get("q"))
			}
		}
	})

	c.Visit("https://www.google.com/search?hl=en&gl=en&q=tech+interview")
}
