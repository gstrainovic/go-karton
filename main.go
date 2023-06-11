package main

import (
	"os"
	"github.com/BurntSushi/toml"
	"fmt"
)

type Item struct {
	Title  string
	Values []Value
}

type Value struct {
	LinkText int
	Value    float64
}

type Config struct {
	URL    string
	Domain string
	LinksProDurchlauf int
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

	fmt.Println("Starting...")
	fmt.Println("URL:", conf.URL)
	fmt.Println("Domain:", conf.Domain)
	fmt.Println("Links pro Durchlauf:", conf.LinksProDurchlauf)

	allLinks:= getLinks(conf)
	for i := 0; i < len(allLinks); i += conf.LinksProDurchlauf {
		links := allLinks[i:i+conf.LinksProDurchlauf]
		fmt.Println("Links:", links)
		data := getData(links)
		saveData(data)
	}
}


