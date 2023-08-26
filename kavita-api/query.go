package kavitaapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	yaml "gopkg.in/yaml.v3"
)

type FilterField int64

//go:generate go run golang.org/x/tools/cmd/stringer -type=FilterField
const (
	Summary           FilterField = 0
	SeriesName        FilterField = 1
	PublicationStatus FilterField = 2
	Languages         FilterField = 3
	AgeRating         FilterField = 4
	UserRating        FilterField = 5
	Tags              FilterField = 6
	CollectionTags    FilterField = 7
	Translators       FilterField = 8
	Characters        FilterField = 9
	Publisher         FilterField = 10
	Editor            FilterField = 11
	CoverArtist       FilterField = 12
	Letterer          FilterField = 13
	Colorist          FilterField = 14
	Inker             FilterField = 15
	Penciller         FilterField = 16
	Writers           FilterField = 17
	Genres            FilterField = 18
	Libraries         FilterField = 19
	ReadProgress      FilterField = 20
	Formats           FilterField = 21
	ReleaseYear       FilterField = 22
	ReadTime          FilterField = 23
)

var MapEnumStringToFilterField = func() map[string]FilterField {
	m := make(map[string]FilterField)
	for i := Summary; i <= ReadTime; i++ {
		m[i.String()] = i
	}
	return m
}()

func (f *FilterField) UnmarshalYAML(value *yaml.Node) error {
	field, found := MapEnumStringToFilterField[value.Value]
	if !found {
		return fmt.Errorf("couldn't convert '%s' to FilterField", value.Value)
	}
	*f = field
	return nil
}

type Comparison int64

//go:generate go run golang.org/x/tools/cmd/stringer -type=Comparison
const (
	Equal            Comparison = 0
	GreaterThan      Comparison = 1
	GreaterThanEqual Comparison = 2
	LessThan         Comparison = 3
	LessThanEqual    Comparison = 4
	Contains         Comparison = 5
	Matches          Comparison = 6
	NotContains      Comparison = 7
	NotEqual         Comparison = 9
	BeginsWith       Comparison = 10
	EndsWith         Comparison = 11
	IsBefore         Comparison = 12
	IsAfter          Comparison = 13
	IsInLast         Comparison = 14
	IsNotInLast      Comparison = 15
)

var MapEnumStringToComparison = func() map[string]Comparison {
	m := make(map[string]Comparison)
	for i := Equal; i <= IsNotInLast; i++ {
		m[i.String()] = i
	}
	return m
}()

func (c *Comparison) UnmarshalYAML(value *yaml.Node) error {
	comparison, found := MapEnumStringToComparison[value.Value]
	if !found {
		return fmt.Errorf("couldn't convert '%s' to Comparison", value.Value)
	}
	*c = comparison
	return nil
}

type Filter struct {
	Field   FilterField `yaml:"field" json:"field"`
	Compare Comparison  `yaml:"comparison" json:"comparison"`
	Value   string      `yaml:"value" json:"value"`
}

type FilterCombination int64

//go:generate go run golang.org/x/tools/cmd/stringer -type=FilterCombination
const (
	Or  FilterCombination = 0
	And FilterCombination = 1
)

var MapEnumStringToFilterCombination = map[string]FilterCombination{
	"Or":  Or,
	"And": And,
}

func (fc *FilterCombination) UnmarshalYAML(value *yaml.Node) error {
	FilterCombination, found := MapEnumStringToFilterCombination[value.Value]
	if !found {
		return fmt.Errorf("couldn't convert '%s' to FilterCombination", value.Value)
	}
	*fc = FilterCombination
	return nil
}

type Query struct {
	Name     string            `yaml:"name" json:"name"`
	JoinType FilterCombination `yaml:"join_type" json:"combination"`
	Filters  []Filter          `yaml:"filters" json:"statements"`
}

func (s *Server) QueryServer(queries []Query) ([]Series, error) {
	series := make(map[int64]Series)

	for _, query := range queries {
		// Request all series
		b, err := json.Marshal(query)
		if err != nil {
			log.Printf("Couldn't marshal query to JSON %v %v", query, err)
			continue
		}
		req, err := http.NewRequest("POST", s.baseURL+"/api/Series/v2", bytes.NewBuffer(b))
		if err != nil {
			log.Printf("Couldn't create http request %v", err)
			continue
		}
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+s.token)
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

		// Parse series
		fetchedSeries := []Series{}
		err = json.Unmarshal(body, &fetchedSeries)
		if err != nil {
			log.Printf("Failed to parse json response %s %v", string(body), err)
			continue
		}
		for _, s := range fetchedSeries {
			existing, got := series[s.ID]
			if got {
				existing.Shelves = append(existing.Shelves, query.Name)
			} else {
				s.Shelves = []string{query.Name}
				series[s.ID] = s
			}
		}
	}
	seriesList := []Series{}
	for _, s2 := range series {
		seriesList = append(seriesList, s2)
	}

	return seriesList, nil
}
