package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/andygrunwald/go-jira"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/robfig/cron"
	"io/ioutil"
	"log"
	"strings"
)

var JiraQuerySize = 500
var JiraJQL = "project in (TIDB, ONCALL, TOOL, TIKV)"

const indexName = "tidb-bug"
const issueLinkFormat = "https://internal.pingcap.net/jira/browse/%s"

func ScheduleSyncTask(jiraClient *jira.Client, esClient *elasticsearch.Client) {
	err := tryToSyncFirstTime(jiraClient, esClient)
	if err != nil {
		log.Fatal("sync failed", err)
	}
	c := cron.New()
	spec := "0 0 2 * * *"
	err = c.AddFunc(spec, func() {
		ReSyncAll(jiraClient, esClient)
	})
	if err != nil {
		log.Fatal("add cron task failed", err)
	}
	c.Start()
}

func ReSyncAll(jiraClient *jira.Client, esClient *elasticsearch.Client) {
	for {
		err := deleteIndexIfExists(esClient, indexName)
		if err != nil {
			log.Println("delete index failed", err)
			continue
		}
		err = tryToSyncFirstTime(jiraClient, esClient)
		if err != nil {
			log.Println("sync jira issue failed", err)
			continue
		}
		break
	}
}

func tryToSyncFirstTime(jiraClient *jira.Client, esClient *elasticsearch.Client) error {
	res, err := esClient.Indices.Exists([]string{indexName})
	if err != nil {
		log.Println("check index failed", err)
		return err
	}
	res.Body.Close()
	if res.StatusCode == 404 {
		log.Println("index not found,create new one")
		createRes, err := esClient.Indices.Create("tidb-bug", esClient.Indices.Create.WithHuman())
		if err != nil {
			return err
		}
		createRes.Body.Close()

		//sync jira issue to es
		log.Println("start to sync jira issue")
		err = sync(jiraClient, esClient)
		log.Println("sync jira issue done", err)
		return err
	}
	return nil
}

func sync(jiraClient *jira.Client, esClient *elasticsearch.Client) error {
	return jiraClient.Issue.SearchPages(JiraJQL,
		&jira.SearchOptions{
			Fields:     []string{ /*"comment", "description", "summary", "label", "sprint"*/ "*all"},
			MaxResults: JiraQuerySize,
		},
		func(issue jira.Issue) error {
			saveJiraIssueToES(esClient, issue)
			return nil
		})
}

func deleteIndexIfExists(esClient *elasticsearch.Client, indexName string) error {
	res, err := esClient.Indices.Exists([]string{indexName})
	if err != nil {
		log.Println("check index failed", err)
		return err
	}
	defer res.Body.Close()
	if res.StatusCode == 200 {
		log.Println("found index")
		res, err := esClient.Indices.Delete([]string{indexName})
		if err != nil {
			log.Println("delete index failed", err)
		}
		if res.StatusCode != 200 {
			data, _ := ioutil.ReadAll(res.Body)
			dataStr := string(data)
			return errors.New(fmt.Sprintf("insert issue failed, code=%d, data=%s", res.StatusCode, dataStr))
		}
	}
	return nil
}

func saveJiraIssueToES(esClient *elasticsearch.Client, issue jira.Issue) {
	data, err := issueToString(issue)
	if err != nil {
		log.Println("marshal issue failed", err)
		return
	}
	insertRs, err := esClient.Index(indexName, strings.NewReader(string(data)))
	defer insertRs.Body.Close()
	if err != nil {
		log.Println(err)
		return
	}
	if insertRs.StatusCode >= 300 {
		data, _ := ioutil.ReadAll(insertRs.Body)
		dataStr := string(data)
		log.Println(fmt.Sprintf("insert issue failed, code=%d, data=%s", insertRs.StatusCode, dataStr))
	} else {
		log.Println(fmt.Sprintf("save to es done, key=%s", issue.Key))
	}
}

func issueToString(issue jira.Issue) ([]byte, error) {
	//esIssue := &JiraIssue{
	//	Link:        fmt.Sprintf(issueLinkFormat, issue.Key),
	//	Title:       issue.Fields.Summary,
	//	Description: issue.Fields.Description,
	//}
	//var comments []string
	//if issue.Fields.Comments != nil && len(issue.Fields.Comments.Comments) > 0 {
	//	for _, comment := range issue.Fields.Comments.Comments {
	//		comments = append(comments, comment.Body)
	//	}
	//}
	//esIssue.Comments = comments
	return json.Marshal(issue)
}
