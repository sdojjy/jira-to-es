package store

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/pkg/errors"
	"io/ioutil"
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
      "link": {},
      "comment": {},
      "description": {},
      "title":{}
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
	return &multiMatch{
		Query:  query,
		Fields: []string{"link^1.0", "comment^1.0", "description^1.0", "title^1.0"},
		Type:   "most_fields",
	}
}

func SearchIssue(query string, from, size int, esClient *elasticsearch.Client) (*SearchResponse, error) {
	buf, err := json.Marshal(newQuery(query))
	if err != nil {
		return nil, err
	}
	res, err := esClient.Search(
		esClient.Search.WithIndex(indexName),
		esClient.Search.WithBody(strings.NewReader(fmt.Sprintf(queryFormat, string(buf)))),
		esClient.Search.WithFrom(from),
		esClient.Search.WithSize(size),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
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
		searchItem.Link = item.Source.Link
		searchItem.Title = item.Source.Title
		for _, highlight := range item.Highlight.Title {
			searchItem.Highlight = append(searchItem.Highlight, highlight)
		}
		for _, highlight := range item.Highlight.Link {
			searchItem.Highlight = append(searchItem.Highlight, highlight)
		}
		for _, highlight := range item.Highlight.Description {
			searchItem.Highlight = append(searchItem.Highlight, highlight)
		}
		for _, highlight := range item.Highlight.Comments {
			searchItem.Highlight = append(searchItem.Highlight, highlight)
		}
		searchRes.Items = append(searchRes.Items, searchItem)
	}
	return searchRes
}
