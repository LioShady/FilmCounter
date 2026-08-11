package pkapi

import (
	"log"
	"net/http"
	"time"
)

// Client fetches film information from the PoiskKino API.
type Client struct {
	key        string
	baseURL    string
	httpClient *http.Client
	logger     *log.Logger
}

// New creates new API client
func New(key string, logger *log.Logger) *Client {
	c := &Client{key: key, baseURL: "https://api.poiskkino.dev", httpClient: &http.Client{Timeout: 10 * time.Second}, logger: logger}
	c.logger.Println("creating new api client")
	return c
}
