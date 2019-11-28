package main

import (
	"flag"
	"github.com/andygrunwald/go-jira"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/eucalytus/session"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
	"github.com/sdojjy/tidb-bug-search-engine/api"
	"github.com/sdojjy/tidb-bug-search-engine/filters"
	_ "github.com/sdojjy/tidb-bug-search-engine/statik"
	"github.com/sdojjy/tidb-bug-search-engine/store"
	"log"
	"time"
)

var (
	jiraUsername string
	jiraPassword string
	esURL        string
	address      string
	uiPath       string
)

func init() {
	flag.StringVar(&jiraUsername, "jira-username", "", "jira username")
	flag.StringVar(&jiraPassword, "jira-password", "", "jira password")
	flag.StringVar(&esURL, "es-url", "http://127.0.0.1:9200", "elastic search url")
	flag.StringVar(&store.IndexName, "es-index", "tidb-bug", "elastic index name")
	flag.StringVar(&address, "listen-address", ":80", "web server listen address")
	flag.IntVar(&store.JiraQuerySize, "jira-query-size", 500, "jira search result size per query")
	flag.StringVar(&store.JiraJQL, "jira-jql", "project in (TIDB, ONCALL, TOOL, TIKV)", "the jira jql to search all issues that should be save to es")
	flag.StringVar(&filters.GoogleOauthConfig.ClientID, "google-oauth-client-id", "", "google oauth client id")
	flag.StringVar(&filters.GoogleOauthConfig.ClientSecret, "google-oauth-client-secret", "", "google oauth secret")
	flag.StringVar(&filters.GoogleOauthConfig.RedirectURL, "google-oauth-callback-url", "http://jirasearch.pingcap.net/auth/callback", "google oauth callback url")
	flag.StringVar(&uiPath, "ui-path", "", "ui path directory")
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

	manager := session.NewManager(session.Options{
		MaxInactiveInterval: 1800, MaxAge: 84000, HttpOnly: false, Path: "/",
	}, session.CreateMemSession,
		//listen session event
		func(s session.Session, event int) {
			if event == session.Created {
				log.Printf("new session is created, sessionId=%s\n", s.GetMaskedSessionId())
			} else if event == session.Destroyed {
				log.Printf("session is destroyed, sessionId=%s\n", s.GetMaskedSessionId())
			} else {
				log.Printf("session is updated, sessionId=%s\n", s.GetMaskedSessionId())
			}
		},
	)
	statikFS, err := fs.New()
	if err != nil {
		log.Fatal(err)
	}

	//create api router
	server := api.New(jiraClient, esClient)
	engine := gin.Default()
	engine.Use(filters.Auth(manager))
	// Serve the contents over HTTP.
	if uiPath != "" {
		engine.Use(filters.Serve("/", statikFS))
	} else {
		engine.Use(static.ServeRoot("/", uiPath))
	}
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
	auth := engine.Group("/auth")
	auth.GET("/login", filters.OauthGoogleLogin)
	auth.GET("/callback", filters.GoogleOAuthCallback(manager))

	log.Fatal("gin run failed", engine.Run(address))
}
