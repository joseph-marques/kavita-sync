package kavitaapi

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

type Book struct {
	Title        string   `json:"title"`
	RelativePath string   `json:"path"`
	URL          string   `json:"url"`
	Shelves      []string `json:"shelves"`
	ID           string   `json:"id"`
}

func (s *Server) makeSeriesURL(series *Series) string {
	return s.baseURL + "/api/opds/" + s.key + "/series/" + fmt.Sprint(series.ID)
}

func (s *Server) FetchBooks(seriesList []Series) ([]Book, error) {
	booksMap := make(map[string]Book)
	for _, series := range seriesList {
		req, err := http.NewRequest("GET", s.makeSeriesURL(&series), nil)
		if err != nil {
			log.Printf("Couldn't form request body %v", err)
			continue
		}
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("Couldn't query server %v", err)
			continue
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Couldn't read response body %v", err)
			continue
		}
		if resp.StatusCode != 200 {
			log.Printf("Bad return status %d body %s", resp.StatusCode, string(body))
			continue
		}
		feed := Feed{}
		err = xml.Unmarshal(body, &feed)
		if err != nil {
			log.Printf("Couldn't parse opds document %v", err)
			continue
		}

		for _, entry := range feed.Entries {
			if entry.Format != "Epub" {
				continue
			}
			book := Book{
				ID:    entry.Id,
				Title: entry.Title,
			}
			for _, link := range entry.Links {
				if link.Type == "application/epub+zip" && strings.HasPrefix(link.Rel, "http://opds-spec.org/acquisition") {
					book.URL = s.baseURL + link.Href
				}
			}
			if book.URL == "" {
				continue
			}
			gotBook, got := booksMap[book.ID]
			if got {
				gotBook.Shelves = append(gotBook.Shelves, series.Shelves...)
			} else {
				book.Shelves = series.Shelves
				booksMap[book.ID] = book
			}

		}
	}

	bookList := []Book{}
	for _, b := range booksMap {
		bookList = append(bookList, b)
	}
	return bookList, nil
}

func diffBooks(newBooks []Book, oldBooks []Book) []Book {
	mb := make(map[string]struct{}, len(oldBooks))
	for _, x := range oldBooks {
		mb[x.ID] = struct{}{}
	}
	var diff []Book
	for _, x := range newBooks {
		if _, found := mb[x.ID]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func (s *Server) DownloadBooks(books []Book, folder string) error {
	folder = strings.TrimRight(folder, "/")
	indexFile, err := os.Open(folder + "/kavita-books.json")
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	oldBooks := []Book{}
	if !os.IsNotExist(err) {
		b, err := io.ReadAll(indexFile)
		if err != nil {
			indexFile.Close()
			return fmt.Errorf("couldn't read existing index file: %v", err)
		}
		indexFile.Close()
		err = json.Unmarshal(b, &oldBooks)
		if err != nil {
			return fmt.Errorf("couldn't parse existing index file: %v", err)
		}
	}
	booksToDownload := diffBooks(books, oldBooks)
	booksToDelete := diffBooks(oldBooks, books)
	for _, book := range booksToDelete {
		err := os.Remove(folder + "/" + book.RelativePath)
		if err != nil {
			log.Printf("Failed to delete %s, %v\n", book.RelativePath, err)
		}
	}
	for i, book := range booksToDownload {
		book.RelativePath = fmt.Sprintf("%s.epub", book.ID)
		req, err := http.NewRequest("GET", book.URL, nil)
		if err != nil {
			log.Printf("Couldn't make book request %v", err)
		}
		resp, err := s.client.Do(req)
		if err != nil {
			log.Printf("Couldn't query server %v", err)
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			log.Printf("Bad return status %d for book", resp.StatusCode)
			continue
		}

		out, err := os.Create(folder + "/" + book.RelativePath)
		if err != nil {
			log.Printf("Couldn't create file: %v", err)
			continue
		}
		defer out.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			log.Printf("Couldn't write to file: %v", err)
			continue
		}
		books[i] = book
	}

	b, err := json.Marshal(books)
	if err != nil {
		return fmt.Errorf("couldn't format books list to json %v", err)
	}
	return os.WriteFile(folder+"/kavita-books.json", b, 0644)
}
