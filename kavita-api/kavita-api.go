package kavitaapi

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Server struct {
	baseURL string
	token   string
	client  *http.Client
	key     string
}

type Series struct {
	ID      int64    `json:"id"`
	Shelves []string `json:"-"`
}

func CreateServer(server string, apiKey string) (*Server, error) {
	server = strings.TrimRight(server, "/")
	m, b := map[string]string{"apiKey": apiKey}, new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(m)
	if err != nil {
		return nil, err
	}
	resp, err := http.Post(server+"/api/Account/login", "application/json", b)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var respBody map[string]interface{}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bodyBytes, &respBody)
	if err != nil {
		return nil, err
	}
	return &Server{
		baseURL: server,
		token:   respBody["token"].(string),
		client:  &http.Client{},
		key:     apiKey,
	}, nil
}
