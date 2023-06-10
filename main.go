package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"context"

	"github.com/BurntSushi/toml"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

func main() {
	type Config struct {
		URL    string
		Domain string
	}

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
	var returnArray []map[string]interface{}
	var linkTextList []float64

	rootCollector := colly.NewCollector()

	rootCollector.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")

		// the link must be a link without domain and containx a 'x'
		if strings.Contains(link, "x") && !strings.Contains(link, "http") {
			fullPathUrl := conf.Domain + link
			links = append(links, fullPathUrl)
			fmt.Printf("Link found: %s\n", fullPathUrl)
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

		valuesArray := []map[string]interface{}{}

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

			valuesArray = append(valuesArray, map[string]interface{}{
				"linkText": linkTextNumber,
				"value":    valueFloat,
			})
		})

		if len(valuesArray) > 0 && title != "" {
			returnArray = append(returnArray, map[string]interface{}{
				"title":  title,
				"values": valuesArray,
			})
		}
	})

	linkCollector.OnError(func(r *colly.Response, err error) {
		log.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	// err = c.Visit(conf.URL)
	// visit all links
	for _, link := range links {
		err = linkCollector.Visit(link)
		if err != nil {
			panic(err)
		}
	}

	for _, item := range returnArray {
		for _, value := range item["values"].([]map[string]interface{}) {
			linkText := value["linkText"].(float64)
			linkTextList = append(linkTextList, linkText)
		}
	}

	sort.Float64s(linkTextList)

	// make sure you've exported GOOGLE_APPLICATION_CREDENTIALS="path/to/your/credentials.json"
	credentials, err := google.FindDefaultCredentials(nil, "https://www.googleapis.com/auth/spreadsheets")
	if err != nil {
		log.Fatal("Error finding Google Sheets credentials:", err)
	}

	if credentials == nil {
		log.Fatal("No Google Sheets credentials found")
	}


	ctx := context.Background()
	client, err := google.DefaultClient(ctx, sheets.SpreadsheetsScope)
	if err != nil {
		log.Fatal("Error creating Google Sheets client:", err)
	}

	sheetsService, err := sheets.New(client)
	if err != nil {
		log.Fatal("Error creating Google Sheets service:", err)
	}


	spreadsheetID := "YOUR_SPREADSHEET_ID"
	sheetName := "output"

	// Create a new sheet named output, if it doesn't exist
	sheetExists := false
	spreadsheet, err := sheetsService.Spreadsheets.Get(spreadsheetID).Do()
	if err != nil {
		log.Println("Error retrieving spreadsheet:", err)
	} else {
		for _, sheet := range spreadsheet.Sheets {
			if sheet.Properties.Title == sheetName {
				sheetExists = true
				break
			}
		}
	}

	if !sheetExists {
		addSheetRequest := &sheets.Request{
			AddSheet: &sheets.AddSheetRequest{
				Properties: &sheets.SheetProperties{
					Title: sheetName,
				},
			},
		}

		batchUpdateRequest := &sheets.BatchUpdateSpreadsheetRequest{
			Requests: []*sheets.Request{addSheetRequest},
		}

		_, err = sheetsService.Spreadsheets.BatchUpdate(spreadsheetID, batchUpdateRequest).Do()
		if err != nil {
			log.Println("Error creating new sheet:", err)
		}
	}

	// Clear the sheet
	clearValuesRequest := &sheets.ClearValuesRequest{}
	_, err = sheetsService.Spreadsheets.Values.Clear(spreadsheetID, sheetName, clearValuesRequest).Do()
	if err != nil {
		log.Println("Error clearing sheet:", err)
	}

	// Write the data to the sheet
	var rows [][]interface{}
	headers := make([]interface{}, 0)
	headers = append(headers, "Title")
	for _, linkText := range linkTextList {
		headers = append(headers, interface{}(linkText))
	}

	// Append headers to rows
	rows = append(rows, headers)

	valueRange := &sheets.ValueRange{
		Values: rows,
	}

	_, err = sheetsService.Spreadsheets.Values.Append(spreadsheetID, sheetName, valueRange).ValueInputOption("USER_ENTERED").Do()
	if err != nil {
		log.Println("Error appending values to sheet:", err)
	}

	fmt.Println(returnArray)
	fmt.Println(linkTextList)
}
