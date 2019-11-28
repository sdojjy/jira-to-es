package filters

import (
	"github.com/eucalytus/session"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Auth(manager *session.Manager) func(c *gin.Context) {
	return func(c *gin.Context) {
		request := c.Request
		if request.URL.Path == "/auth/login" || request.URL.Path == "/auth/callback" {
			c.Next()
			return
		}
		s := manager.GetSession(request)
		if s != nil {
			if _, found := s.Get("key"); found {
				c.Next()
				return
			}
		}
		c.Redirect(http.StatusTemporaryRedirect, "/auth/login")
		c.Abort()
	}
}
