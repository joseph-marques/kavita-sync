package main

import (
	"flag"
	"log"

	"github.com/joseph-marques/kavita-sync/update-bookshelves/database"
)

func main() {
	db_path := flag.String("db_path", "/mnt/onboard/.kobo/KoboReader.sqlite", "Path to KoboReader.sqlite file.")
	books_path := flag.String("books_path", "/mnt/onboard/kavita-sync", "Path where the books are synced to.")
	flag.Parse()
	db, err := database.OpenDB(*db_path)
	if err != nil {
		log.Fatalf("Couldn't open DB: %v", err)
	}
	shelves, err := db.GetShelves()
	if err != nil {
		log.Fatalf("Couldn't get shelves: %v", err)
	}
	err = db.AddBooksToShelves(*books_path, shelves)
	if err != nil {
		log.Fatalf("Couldn't add book to shelves: %v", err)
	}
}
