package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/tealeg/xlsx"
	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	URL    string
	Domain string
}

type Item struct {
	Title  string
	Values []Value
}

type Value struct {
	LinkText int
	Value    float64
}

func main() {
	b, err := os.ReadFile("./config.toml")
	if err != nil {
		panic(err)
	}

	var conf Config
	_, err = toml.Decode(string(b), &conf)
	if err != nil {
		panic(err)
	}

	fmt.Println(conf.URL)

	var links []string
	var returnArray []Item

	rootCollector := colly.NewCollector()

	rootCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// the link must be a link without a domain and contain 'x'
		if strings.Contains(link, "x") && !strings.Contains(link, "http") {
			fullPathURL := conf.Domain + link
			links = append(links, fullPathURL)
			fmt.Printf("Link found: %s\n", fullPathURL)
		}
	})

	err = rootCollector.Visit("https://www.karton.eu/Unsere-Kartonagen/")
	if err != nil {
		panic(err)
	}

	linkCollector := colly.NewCollector()
	linkCollector.OnHTML("h1", func(e *colly.HTMLElement) {
		title := e.Text
		fmt.Println("Title:", title)

		valuesArray := []Value{}

		tableRows := e.DOM.ParentsUntil("~").Find("table tr")

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

		if len(valuesArray) > 0 && title != "" {
			returnArray = append(returnArray, Item{
				Title:  title,
				Values: valuesArray,
			})
		}
	})

	linkCollector.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// For debugging only the first 10 links
	links = links[:10]

	for _, link := range links {
		err = linkCollector.Visit(link)
		if err != nil {
			panic(err)
		}
	}

	// Create a new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Fatal("Error creating sheet:", err)
	}

	// Add the header row
	headerRow := sheet.AddRow()
	headerRow.AddCell().SetValue("Title")
	linkTexts := getSortedLinkTexts(returnArray)
	for _, linkText := range linkTexts {
		headerRow.AddCell().SetValue(strconv.Itoa(linkText))
	}

	// Add data rows
	for _, item := range returnArray {
		dataRow := sheet.AddRow()
		dataRow.AddCell().SetValue(item.Title)

		// Initialize the values map
		valuesMap := make(map[int]float64)
		for _, value := range item.Values {
			valuesMap[value.LinkText] = value.Value
		}

		for _, linkText := range linkTexts {
			dataRow.AddCell().SetValue(strconv.FormatFloat(valuesMap[linkText], 'f', 2, 64))
		}
	}

	// Save the Excel file
	err = file.Save("data.xlsx")
	if err != nil {
		log.Fatal("Error saving Excel file:", err)
	}

	fmt.Println("Data successfully saved to data.xlsx.")
}

// Helper function to get sorted link texts from the returnArray
func getSortedLinkTexts(items []Item) []int {
	linkTexts := make(map[int]bool)
	for _, item := range items {
		for _, value := range item.Values {
			linkTexts[value.LinkText] = true
		}
	}

	var sortedLinkTexts []int
	for linkText := range linkTexts {
		sortedLinkTexts = append(sortedLinkTexts, linkText)
	}

	sort.Ints(sortedLinkTexts)
	return sortedLinkTexts
}
