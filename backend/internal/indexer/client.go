package indexer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// IndexEntry represents a single entry from index.golang.org.
type IndexEntry struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Timestamp time.Time `json:"Timestamp"`
}

// IndexClient defines operations for fetching from index.golang.org.
type IndexClient interface {
	FetchReleases(since time.Time, limit int) ([]IndexEntry, error)
}

// HTTPIndexClient implements IndexClient using HTTP.
type HTTPIndexClient struct {
	baseURL string
	client  *http.Client
}

// NewIndexClient creates a new IndexClient.
func NewIndexClient() IndexClient {
	return &HTTPIndexClient{
		baseURL: "https://index.golang.org", // TODO: Use environment variable or const.
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchReleases fetches releases from index.golang.org since the given timestamp.
func (c *HTTPIndexClient) FetchReleases(since time.Time, limit int) ([]IndexEntry, error) {
	url := fmt.Sprintf("%s/index?since=%s&limit=%d",
		c.baseURL,
		since.UTC().Format(time.RFC3339),
		limit,
	)

	resp, err := c.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("indexer: fetch releases: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("indexer: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var entries []IndexEntry
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var entry IndexEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("indexer: unmarshal entry: %w", err)
		}
		entries = append(entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("indexer: scan response: %w", err)
	}

	return entries, nil
}
