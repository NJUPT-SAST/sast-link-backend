package v1

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/endpoints"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

const (
	appAccessTokenURL = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	userAccessTokenURL = "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
)

var (
	larkConf = oauth2.Config{
		ClientID:     config.Config.GetString("oauth.client.lark.id"),
		ClientSecret: config.Config.GetString("oauth.client.lark.secret"),
		RedirectURL:  "http://localhost:8080/api/v1/login/lark/callback",
		Scopes:       []string{},
		Endpoint:     endpoints.Lark,
	}
)

// OauthLarkLogin redirect url to lark auth page.
func OauthLarkLogin(c *gin.Context) {
	oauthState := GenerateStateOauthCookie(c.Writer)
	url := larkConf.AuthCodeURL(oauthState)

	log.Logger.Warnf("Visit the URL for the auth dialog: %v\n", url)

	c.Redirect(http.StatusFound, url) 
}

// OauthLarkCallback read url from lark callback, 
// get `code`, request app_access_token,
// at last request lark url to get user_access_token.
func OauthLarkCallback(c *gin.Context) {
	oauthState, _ := c.Request.Cookie("oauthstate")

	if c.Request.FormValue("state") != oauthState.Value {
		log.Logger.Errorln("invalid oauth state, expected '%s', got '%s'\n", oauthState.Value, c.Request.FormValue("state"))
		c.Redirect(http.StatusFound, "/")
		return
	}

	code := c.Query("code")
	log.Logger.Debugf("\ncode ::: %s\n", code)
	accessToken, err := getLarkAppAccessToken()

	if err != nil {
		log.Logger.Errorln("getLarkAppAccessToken ::: ", err)
		c.JSON(http.StatusInternalServerError, result.Failed(result.HandleError(err)))
		return
	}

	data := map[string]string {
		"grant_type": "authorization_code",
		"code": code,
	}

	header := map[string]string {
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		"Content-Type": "application/json; charset=utf-8",
	}

	res, err := util.PostWithHeader(userAccessTokenURL, header, data)
	if err != nil {
		log.Logger.Errorln("util.PostWithHeader ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(result.AccessTokenErr)))
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Logger.Errorln("io.ReadAll ::: ", err)
		c.JSON(http.StatusInternalServerError, result.Failed(result.HandleError(err)))
		return
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Logger.Errorf("gjson.Get ::: response code: %d\n", resCode)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(errors.New(fmt.Sprintf("OauthLarkCallback resCode: %d", resCode)))))
		return
	}

	userAccessToken := gjson.Get(string(body), "data.access_token").String()
	expire := gjson.Get(string(body), "data.expire_in").Int()

	model.Rdb.Set(model.RedisCtx, "lark_user_access_token", userAccessToken, time.Duration(expire))

	c.JSON(http.StatusOK, result.Success(map[string]string{"user_access_token": userAccessToken}))

}

// Get Lark app_access_token
func getLarkAppAccessToken() (string, error) {
	appId := larkConf.ClientID
	appSecret := larkConf.ClientSecret

	params := url.Values{}
	params.Add("app_id", appId)
	params.Add("app_secret", appSecret)

	res, error := http.PostForm(appAccessTokenURL, params)
	if error != nil {
		log.Logger.Errorln("http.PostForm ::: ", error)
		return "", error
	}
	log.LogRes(res)

	body, error := io.ReadAll(res.Body)
	defer res.Body.Close()
	if error != nil {
		log.Logger.Errorln("io.ReadAll ::: ", error)
		return "", error
	}


	if code := gjson.Get(string(body), "code").Int(); code != 0 {
		log.Logger.Errorln("gjson.Get ::: code:", code)
		return "", result.InternalErr
	}

	acceToken := gjson.Get(string(body), "app_access_token").String()
	expire := gjson.Get(string(body), "expire").Int()

	model.Rdb.Set(model.RedisCtx, "lark_app_access_token", acceToken, time.Duration(expire))

	return acceToken, nil
}


// func getUserInfo(user_access_token string) 
