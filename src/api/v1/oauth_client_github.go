package v1

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/endpoints"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
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
		RedirectURL:  "http://localhost:3000/callback/github",
		Scopes:       []string{},
		Endpoint:     endpoints.GitHub,
	}
)

func OauthGithubLogin(c *gin.Context) {

	redirectURL := c.Query("redirect_url")
	githubConf.RedirectURL = redirectURL
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

	// githubinfo is the user info from github
	githubInfo, err := getUserInfoFromGithub(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}
	if githubInfo == "" {
		c.JSON(http.StatusOK, result.Failed(result.HandleError(result.RequestParamError)))
		return
	}

	githubID := gjson.Get(githubInfo, "id").String()
	// userInfo is the github user info in the database
	userInfo, err := service.GetUserByGithubId(githubID)
	if err != nil {
		log.Errorf("service.GetUserByGithubId ::: %s", err.Error())
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	// Store to redis
	model.Rdb.Set(model.RedisCtx, githubID,
		githubInfo, time.Duration(model.OAUTH_USER_INFO_EXP))

	// User not found, Need to register to bind the github id
	if userInfo == nil {
		log.Debugf("User not found, Need to register to bind the github id: %s", githubID)
		oauthToken, err := util.GenerateTokenWithExp(c, model.OauthSubKey(githubID, model.OAUTH_GITHUB_SUB), model.OAUTH_TICKET_EXP)

		if err != nil {
			c.JSON(http.StatusOK, result.Failed(result.GenerateToken))
			log.Log.Errorln("util.GenerateTokenWithExp ::: ", err)
			return
		}
		c.JSON(http.StatusOK, result.Response{
			Success: false,
			ErrCode: result.OauthUserUnbounded.ErrCode,
			ErrMsg:  result.OauthUserUnbounded.ErrMsg,
			Data: gin.H{
				"oauthTicket": oauthToken,
			},
		})
		return
	} else {
		// User already registered and bounded github,
		// directly return token
		uid := userInfo.UserID
		log.Debugf("User already registered and bounded github: %s", uid)
		token, err := util.GenerateTokenWithExp(c, model.LoginJWTSubKey(uid), model.LOGIN_TOKEN_EXP)
		if err != nil {
			c.JSON(http.StatusOK, result.Failed(result.GenerateToken))
			return
		}
		model.Rdb.Set(c, model.LoginTokenKey(uid), token, model.LOGIN_TOKEN_EXP)
		c.JSON(http.StatusOK, result.Success(gin.H{
			model.LOGIN_TOKEN_SUB: token,
		}))
		return
	}
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

	log.Debugf("Github user info: %s", gjson.ParseBytes(body).String())

	return gjson.ParseBytes(body).String(), nil
}
