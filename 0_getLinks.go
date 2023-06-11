package main

import (
	"fmt"
	"strings"
	"github.com/gocolly/colly"
	_ "github.com/mattn/go-sqlite3"
)

func getLinks(conf Config) []string {
	var links []string

	collector := colly.NewCollector()
	collector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// the link must be a link without a domain and contain 'x'
		if strings.Contains(link, "x") && !strings.Contains(link, "http") {
			fullPathURL := conf.Domain + link
			links = append(links, fullPathURL)
			fmt.Printf("Link found: %s\n", fullPathURL)
		}
	})
	err := collector.Visit(conf.URL)
	if err != nil {
		panic(err)
	}

	return links
}