package v1

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/endpoints"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

const (
	// GitHub user info url
	GithubUserInfoURL = "https://api.github.com/user"
)

var (
	githubConf = oauth2.Config{
		ClientID:     config.Config.GetString("oauth.client.github.id"),
		ClientSecret: config.Config.GetString("oauth.client.github.secret"),
		RedirectURL:  config.Config.GetString("oauth.client.github.redirect_url"),
		Scopes:       []string{},
		Endpoint:     endpoints.GitHub,
	}
)

func OauthGithubLogin(c *gin.Context) {

	// Create oauthState cookie
	oauthState := GenerateStateOauthCookie(c.Writer)
	url := githubConf.AuthCodeURL(oauthState)

	// log.Log.Warnf("Visit the URL for the auth dialog: %v\n", url)
	log.Debug("Visit the URL for the auth dialog: ", url)
	log.Debug("RedirectURL: ", githubConf.RedirectURL)
	log.Debug("ClientID: ", githubConf.ClientID)

	c.SetCookie("oauthstate", oauthState, 3600, "", "", false, true)
	c.Redirect(http.StatusFound, url)
}

func OauthGithubCallback(c *gin.Context) {
	oauthState, _ := c.Request.Cookie("oauthstate")
	log.Debugf("oauthState: %s", oauthState.Value)

	if c.Request.FormValue("state") != oauthState.Value {
		log.Log.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState.Value, c.Request.FormValue("state"))
		c.Redirect(http.StatusFound, "/")
		return
	}

	code := c.Query("code")

	githubId, err := getUserInfoFromGithub(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}
	if githubId == "" {
		c.JSON(http.StatusOK, result.Failed(result.HandleError(result.RequestParamError)))
		return
	}
	log.Debug("GithubId: ", githubId)

	user, err := service.GetUserByGithubId(githubId)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	// User not found, Need to register to bind the github id
	if user == nil {
		return
	}

	c.JSON(http.StatusOK, result.Success(githubId))
}

func getUserInfoFromGithub(ctx context.Context, code string) (string, error) {

	token, err := githubConf.Exchange(ctx, code)
	if err != nil {
		// log.Log.Errorf("exchange github code error: %s", err.Error())
		log.Errorf("Exchange github code error: %s", err.Error())
		return "", fmt.Errorf("Exchange github code error: %s", err.Error())
	}
	client := &http.Client{}
	req, err := http.NewRequest("GET", GithubUserInfoURL, nil)
	if err != nil {
		// log.Log.Errorf("New request error: %s", err.Error())
		log.Errorf("New request error: %s", err.Error())
		return "", fmt.Errorf("New request error: %s", err.Error())
	}
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	res, err := client.Do(req)
	if err != nil {
		log.Log.Errorf("Failt to getting user info: %s", err.Error())
		return "", fmt.Errorf("Failt to getting user info: %s", err.Error())
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", result.InternalErr
	}

	// TODO:Now just get the github id
	githubId := gjson.Get(string(body), "id").String()
	return githubId, nil
}
