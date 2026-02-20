package curius

import "time"

type Link struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Highlights  []string  `json:"highlights"`
	Tags        []Tag     `json:"tags"`
	CreatedAt   time.Time `json:"createdAt"`
	Description string    `json:"description"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// apiResponse is the raw response from the Curius API.
type apiResponse struct {
	UserSaved []apiLink `json:"userSaved"`
}

type apiLink struct {
	ID          int            `json:"id"`
	Title       string         `json:"title"`
	Link        string         `json:"link"`
	Snippet     string         `json:"snippet"`
	CreatedDate string         `json:"createdDate"`
	Topics      []apiTopic     `json:"topics"`
	Highlights  []apiHighlight `json:"highlights"`
}

type apiTopic struct {
	ID    int    `json:"id"`
	Topic string `json:"topic"`
}

type apiHighlight struct {
	ID        int    `json:"id"`
	Highlight string `json:"highlight"`
}
