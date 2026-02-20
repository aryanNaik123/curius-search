package curius

import "time"

type Link struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	Highlights  []string   `json:"highlights"`
	Tags        []Tag      `json:"tags"`
	CreatedAt   time.Time  `json:"createdAt"`
	Description string     `json:"description"`
	Content     string     `json:"content"`
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
	ID        int            `json:"id"`
	Title     string         `json:"title"`
	Link      string         `json:"link"`
	Note      string         `json:"note"`
	CreatedAt string         `json:"createdAt"`
	Trails    []apiTrail     `json:"trails"`
	Highlight []apiHighlight `json:"highlight"`
	Metadata  *apiMetadata   `json:"metadata"`
}

type apiTrail struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type apiHighlight struct {
	ID   int    `json:"id"`
	Text string `json:"text"`
}

type apiMetadata struct {
	Description string `json:"description"`
	Content     string `json:"content"`
}
