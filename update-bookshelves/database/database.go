package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	kavitaapi "github.com/joseph-marques/kavita-sync/kavita-api"
	_ "github.com/mattn/go-sqlite3"
)

type KoboDatabase struct {
	db *sql.DB
}

func OpenDB(path string) (*KoboDatabase, error) {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("database doesn't exist: %v", err)
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &KoboDatabase{db}, nil
}

func (db *KoboDatabase) GetShelves() ([]string, error) {
	shelves := []string{}
	rows, err := db.db.Query("SELECT InternalName, Name FROM Shelf;")
	if err != nil {
		return nil, fmt.Errorf("couldn't query db: %v", err)
	}
	for rows.Next() {
		var (
			internalName string
			name         string
		)
		if err := rows.Scan(&internalName, &name); err != nil {
			return nil, fmt.Errorf("couldn't unpack shelf row: %v", err)
		}
		if internalName != name {
			log.Printf("WARNING: unexpected internalName not equal to name, skipping: %s != %s", internalName, name)
			continue
		}
		shelves = append(shelves, name)
	}

	return shelves, nil
}

func readIndexFile(folder string) ([]kavitaapi.Book, error) {
	folder = strings.TrimRight(folder, "/")
	indexFile, err := os.Open(folder + "/kavita-books.json")
	if err != nil {
		return nil, err
	}

	b, err := io.ReadAll(indexFile)
	if err != nil {
		indexFile.Close()
		return nil, fmt.Errorf("couldn't read existing index file: %v", err)
	}
	indexFile.Close()
	books := []kavitaapi.Book{}
	err = json.Unmarshal(b, &books)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse existing index file: %v", err)
	}

	return books, nil
}

func (db *KoboDatabase) AddBooksToShelves(path string, shelves []string) error {
	books, err := readIndexFile(path)
	if err != nil {
		return fmt.Errorf("couldn't read index file: %v", err)
	}
	shelfSet := make(map[string]struct{})
	for _, shelf := range shelves {
		shelfSet[shelf] = struct{}{}
	}

	for _, book := range books {
		contentId := "file://" + filepath.Join(path, book.RelativePath)
		for _, shelf := range book.Shelves {
			if _, got := shelfSet[shelf]; !got {
				log.Printf("Unrecognized shelf %s", shelf)
				continue
			}
			insertStatement := `
			INSERT OR REPLACE INTO ShelfContent (ShelfName, ContentId, DateModified, _IsDeleted, _IsSynced) 
  		VALUES ($1, $2, $3, false, false);
			`
			db.db.Exec(insertStatement, shelf, contentId, time.Now().UTC().Format(time.RFC3339))
		}
	}
	return nil
}
