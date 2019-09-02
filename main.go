package main

import (
	"flag"
	"log"
	"time"

	"github.com/andygrunwald/go-jira"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sdojjy/tidb-bug-search-engine/api"
	"github.com/sdojjy/tidb-bug-search-engine/store"
)

var (
	jiraUsername string
	jiraPassword string
	esURL        string
	address      string
)

func init() {
	flag.StringVar(&jiraUsername, "jira-username", "", "jira username")
	flag.StringVar(&jiraPassword, "jira-password", "", "jira password")
	flag.StringVar(&esURL, "es-url", "http://127.0.0.1:9200", "elastic search url")
	flag.StringVar(&address, "listen-address", ":8888", "web server listen address")
	flag.IntVar(&store.JiraQuerySize, "jira-query-size", 500, "jira search result size per query")
	flag.StringVar(&store.JiraJQL, "jira-jql", "project in (TIDB, ONCALL, TOOL, TIKV)", "the jira jql to search all issues that should be save to es")
}

func main() {
	flag.Parse()
	tp := jira.BasicAuthTransport{
		Username: jiraUsername,
		Password: jiraPassword,
	}
	jiraClient, _ := jira.NewClient(tp.Client(), "https://internal.pingcap.net/jira")

	cfg := elasticsearch.Config{
		Addresses: []string{esURL},
	}
	esClient, err := elasticsearch.NewClient(cfg)
	if err != nil {
		log.Fatal("init elastic search client failed", err)
	}

	//start to sync
	go store.ScheduleSyncTask(jiraClient, esClient)

	//create api router
	server := api.New(jiraClient, esClient)
	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PATCH", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
		MaxAge: 12 * time.Hour,
	}))
	search := engine.Group("/search")
	search.GET("/issues", server.SearchIssue)
	search.POST("/re-sync", server.ReSync)

	log.Fatal("gin run failed", engine.Run(address))
}
