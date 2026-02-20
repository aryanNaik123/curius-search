package curius

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const baseURL = "https://curius.app/api/users"

type Client struct {
	userID     string
	httpClient *http.Client
}

func NewClient(userID string) *Client {
	return &Client{
		userID: userID,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// FetchAllLinks fetches all bookmarks for the user, paginating until no more results.
func (c *Client) FetchAllLinks() ([]Link, error) {
	var all []Link
	page := 0

	for {
		url := fmt.Sprintf("%s/%s/links?page=%d", baseURL, c.userID, page)
		log.Printf("Fetching page %d: %s", page, url)

		resp, err := c.httpClient.Get(url)
		if err != nil {
			return nil, fmt.Errorf("fetch page %d: %w", page, err)
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("fetch page %d: status %d", page, resp.StatusCode)
		}

		var apiResp apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("decode page %d: %w", page, err)
		}
		resp.Body.Close()

		if len(apiResp.UserSaved) == 0 {
			break
		}

		for _, al := range apiResp.UserSaved {
			all = append(all, convertLink(al))
		}

		log.Printf("  Got %d links (total: %d)", len(apiResp.UserSaved), len(all))
		page++
	}

	return all, nil
}

func convertLink(al apiLink) Link {
	l := Link{
		ID:    al.ID,
		Title: al.Title,
		URL:   al.Link,
	}

	if al.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, al.CreatedAt); err == nil {
			l.CreatedAt = t
		}
	}

	for _, h := range al.Highlight {
		if h.Text != "" {
			l.Highlights = append(l.Highlights, h.Text)
		}
	}

	for _, t := range al.Trails {
		l.Tags = append(l.Tags, Tag{ID: t.ID, Name: t.Name})
	}

	if al.Metadata != nil {
		l.Description = al.Metadata.Description
		l.Content = al.Metadata.Content
	}

	if l.Description == "" {
		l.Description = al.Note
	}

	return l
}
