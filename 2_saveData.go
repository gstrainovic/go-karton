package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/tealeg/xlsx"
)

func saveData(returnArray []Item) {
	// Create a new Excel file
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Fatal("Error creating sheet:", err)
	}

	// Add the header row
	headerRow := sheet.AddRow()
	headerRow.AddCell().SetValue("H1")
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

	// create a folder data, if not exist
	if _, err := os.Stat("data"); os.IsNotExist(err) {
		os.Mkdir("data", 0755)
	}

	// Save the Excel file named yymmddmmss.xlsx in the data folder
	filename := fmt.Sprintf("%s.xlsx", time.Now().Format("20060102150405"))
	err = file.Save(fmt.Sprintf("data/%s", filename))
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

