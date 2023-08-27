package main

import (
	"flag"
	"log"
	"os"

	kavitaapi "github.com/joseph-marques/kavita-sync/kavita-api"
	yaml "gopkg.in/yaml.v3"
)

type conf struct {
	BaseURL      string            `yaml:"base_url"`
	APIKey       string            `yaml:"api_key"`
	OutputFolder string            `yaml:"output_folder"`
	Queries      []kavitaapi.Query `yaml:"queries"`
}

func main() {
	var config_path = flag.String("config", "", "The path to the config file for your sync")
	flag.Parse()
	yamlFile, err := os.ReadFile(*config_path)
	if err != nil {
		log.Fatalf("yamlFile.Get err   #%v ", err)
	}
	var c conf
	err = yaml.Unmarshal(yamlFile, &c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}
	if c.BaseURL == "" {
		log.Fatalf("Specify a base_url in the yaml config!")
	}
	if c.APIKey == "" {
		log.Fatalf("Specify an api_key in the yaml config!")
	}
	server, err := kavitaapi.CreateServer(c.BaseURL, c.APIKey)
	if err != nil {
		log.Fatalf("%v", err)
	}
	series, err := server.QueryServer(c.Queries)
	if err != nil {
		log.Fatalf("%v", err)
	}
	books, err := server.FetchBooks(series)
	if err != nil {
		log.Fatalf("%v", err)
	}
	//log.Println(books)

	err = server.DownloadBooks(books, c.OutputFolder)
}
