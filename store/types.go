package store

import "github.com/andygrunwald/go-jira"

type esResponse struct {
	Hits hits `json:"hits"`
}

type hits struct {
	Total int       `json:"total"`
	Hits  []hitItem `json:"hits"`
}

type hitItem struct {
	Source    *jira.Issue         `json:"_source"`
	Highlight map[string][]string `json:"highlight"`
}

type ErrResponse struct {
	ErrMsg string `json:"errMsg"`
}

type SearchResponse struct {
	Total int           `json:"total"`
	Items []*SearchItem `json:"items"`
}

type SearchItem struct {
	Link      string   `json:"link"`
	Title     string   `json:"title"`
	Highlight []string `json:"highlight"`
}

type JiraIssue struct {
	Link        string   `json:"link"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Comments    []string `json:"comments"`
}

type highlightItem struct {
	Link        []string `json:"link"`
	Title       []string `json:"title"`
	Description []string `json:"description"`
	Comments    []string `json:"comments"`
}
