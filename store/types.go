package store

type JiraIssue struct {
	Link        string   `json:"link"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Comments    []string `json:"comments"`
}

type esResponse struct {
	Hits hits `json:"hits"`
}

type hits struct {
	Total int       `json:"total"`
	Hits  []hitItem `json:"hits"`
}

type hitItem struct {
	Source    JiraIssue     `json:"_source"`
	Highlight highlightItem `json:"highlight"`
}

type highlightItem struct {
	Link        []string `json:"link"`
	Title       []string `json:"title"`
	Description []string `json:"description"`
	Comments    []string `json:"comments"`
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
