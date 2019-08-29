package api

import (
	"net/http"
	"strconv"

	"github.com/andygrunwald/go-jira"
	"github.com/elastic/go-elasticsearch/v6"
	"github.com/gin-gonic/gin"
	"github.com/sdojjy/tidb-bug-search-engine/store"
)

type Server struct {
	jiraClient *jira.Client
	esClient   *elasticsearch.Client
}

func New(jiraClient *jira.Client, esClient *elasticsearch.Client) *Server {
	return &Server{
		jiraClient: jiraClient,
		esClient:   esClient,
	}
}

func (server *Server) SearchIssue(ctx *gin.Context) {
	queryString := ctx.Query("q")
	from := parseInt(ctx.DefaultQuery("from", "0"))
	size := parseInt(ctx.DefaultQuery("size", "10"))
	if len(queryString) == 0 {
		ctx.JSON(http.StatusBadRequest, store.ErrResponse{ErrMsg: "missing query string"})
	} else {
		res, err := store.SearchIssue(queryString, from, size, server.esClient)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, store.ErrResponse{ErrMsg: err.Error()})
		} else {
			ctx.JSON(http.StatusOK, res)
		}
	}
}

func (server *Server) ReSync(ctx *gin.Context) {
	go store.ReSyncAll(server.jiraClient, server.esClient)
	ctx.JSON(http.StatusOK, nil)
}

func parseInt(value string) int {
	i, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0
	}
	return (int)(i)
}
