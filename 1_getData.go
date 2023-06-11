package main

import (
	"log"
	// "net/http"
	"strconv"
	"strings"
	// "time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	_ "github.com/mattn/go-sqlite3"
)

func getData(links []string) []Item {
	var returnArray []Item

	// Create a custom HTTP client with an increased timeout duration
	// client := &http.Client{
	// 	Timeout: 30 * time.Second, // Set a timeout of 30 seconds
	// }

	collector := colly.NewCollector(
		colly.Async(true),
		colly.Debugger(&debug.LogDebugger{}),
		// colly.WithHTTPClient(client), // Use the custom HTTP client
	)
	collector.Limit(&colly.LimitRule{
		Parallelism: len(links),
	})

	collector.OnHTML("h1", func(e *colly.HTMLElement) {
		title := e.Text

		tableRows := e.DOM.ParentsUntil("~").Find("table tr")
		numRows := tableRows.Length()

		if numRows == 0 || title == "" {
			return
		}

		valuesArray := make([]Value, 0, numRows) // Allocate capacity based on the number of rows

		tableRows.Each(func(index int, element *goquery.Selection) {
			columns := element.Find("td")

			linkText := columns.Eq(0).Text()
			value := strings.Split(columns.Eq(1).Text(), " ")[0]

			if linkText == "" || value == "" {
				return
			}

			valueFloat, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", "."), 64)
			if err != nil {
				log.Println("Error parsing value:", err)
				return
			}

			linkTextNumber, err := strconv.Atoi(linkText)
			if err != nil {
				log.Println("Error parsing linkText:", err)
				return
			}

			valuesArray = append(valuesArray, Value{
				LinkText: linkTextNumber,
				Value:    valueFloat,
			})
		})

		returnArray = append(returnArray, Item{
			Title:  title,
			Values: valuesArray,
		})
	})

	collector.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	for _, link := range links {
		err := collector.Visit(link)
		if err != nil {
			panic(err)
		}
	}

	collector.Wait()

	return returnArray
}
