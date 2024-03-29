package store

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"log"
	"strings"
)

const queryFormat = `
{
    "query": {
        "function_score": {
            "query": {
                "bool": {
                    "should": [{
                        "multi_match": %s
                    }]
                }
            }
        }
    },
  
   "highlight": {
    "fields": {
      "*": {}
    }
  }
}
`

type multiMatch struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
	Type   string   `json:"type"`
}

func newQuery(query string) *multiMatch {
	query = strings.TrimSpace(query)
	matchType := "most_fields"
	if strings.HasSuffix(query, "\"") && strings.HasPrefix(query, "\"") {
		matchType = "phrase"
	}
	return &multiMatch{
		Query:  query,
		Fields: []string{},
		//Fields: []string{"link^1.0", "comment^1.0", "description^1.0", "title^1.0"},
		Type: matchType,
	}
}

func SearchIssue(query string, from, size int, esClient *elasticsearch.Client) (*SearchResponse, error) {
	buf, err := json.Marshal(newQuery(query))
	if err != nil {
		return nil, err
	}
	res, err := esClient.Search(
		esClient.Search.WithIndex(IndexName),
		esClient.Search.WithBody(strings.NewReader(fmt.Sprintf(queryFormat, string(buf)))),
		esClient.Search.WithFrom(from),
		esClient.Search.WithSize(size),
	)
	if err != nil {
		return nil, err
	}
	defer deferClose(res.Body)
	data, _ := ioutil.ReadAll(res.Body)
	if res.StatusCode != 200 {
		dataStr := string(data)
		return nil, errors.New(dataStr)
	} else {
		esRes := &esResponse{}
		err = json.Unmarshal(data, esRes)
		if err != nil {
			return nil, err
		}
		return esResponse2SearchResponse(esRes), nil
	}
}

func esResponse2SearchResponse(esRes *esResponse) *SearchResponse {
	searchRes := &SearchResponse{}
	searchRes.Total = esRes.Hits.Total
	for _, item := range esRes.Hits.Hits {
		searchItem := &SearchItem{}
		searchItem.Link = fmt.Sprintf(issueLinkFormat, item.Source.Key)
		searchItem.Title = item.Source.Fields.Summary
		for _, highlightField := range item.Highlight {
			for _, highlight := range highlightField {
				searchItem.Highlight = append(searchItem.Highlight, highlight)
			}
		}
		searchRes.Items = append(searchRes.Items, searchItem)
	}
	return searchRes
}

func deferClose(c io.Closer) {
	if err := c.Close(); err != nil {
		log.Print("close failed", err)
	}
}
