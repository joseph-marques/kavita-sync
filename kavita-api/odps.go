package kavitaapi

import "encoding/xml"

// The following structs are meant to parse an SUBSET of the overall opds format.

type Feed struct {
	XMLName xml.Name `xml:"feed"`
	Entries []Entry  `xml:"entry"`
}

type Entry struct {
	Format string `xml:"format"`
	Links  []Link `xml:"link"`
	Id     string `xml:"id"`
	Title  string `xml:"title"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}
