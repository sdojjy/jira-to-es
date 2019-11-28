package filters

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/eucalytus/session"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// Scopes: OAuth 2.0 scopes provide a way to limit the amount of access that is granted to an access token.
var GoogleOauthConfig = &oauth2.Config{
	Scopes:   []string{"https://www.googleapis.com/auth/userinfo.email"},
	Endpoint: google.Endpoint,
}

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func OauthGoogleLogin(c *gin.Context) {

	// Create oauthState cookie
	oauthState := generateStateOauthCookie(c)

	/*
		AuthCodeURL receive state that is a token to protect the user from CSRF attacks. You must always provide a non-empty string and
		validate that it matches the the state query parameter on your redirect callback.
	*/
	u := GoogleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(c.Writer, c.Request, u, http.StatusTemporaryRedirect)
}

func generateStateOauthCookie(c *gin.Context) string {
	var expiration = time.Now().Add(20 * time.Minute)

	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(c.Writer, &cookie)

	return state
}

func GoogleOAuthCallback(manager *session.Manager) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Read oauthState from Cookie
		//oauthState, _ := r.Cookie("oauthstate")
		//
		//if r.FormValue("state") != oauthState.Value {
		//	log.Println("invalid oauth google state")
		//	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		//	return
		//}

		data, err := GetUserDataFromGoogle(c.Request.FormValue("code"))
		if err != nil {
			log.Println(err.Error())
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if s, err := manager.CreateSession(c.Request, c.Writer); err == nil {
			// GetOrCreate User in your db.
			// Redirect or response with a token.
			// More code .....
			fmt.Printf("UserInfo: %s\n", data)
			if err := s.Set("key", "login"); err != nil {
				log.Println(err.Error())
			}
			c.Redirect(http.StatusTemporaryRedirect, "/")
		} else {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func GetUserDataFromGoogle(code string) ([]byte, error) {
	// Use code to get token and get user info from Google.

	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}
