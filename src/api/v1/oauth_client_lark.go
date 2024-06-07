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
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

const (
	AppAccessTokenURL = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	UserAccessTokenURL = "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
)

var (
	larkConf = oauth2.Config{
		ClientID:     config.Config.GetString("oauth.client.lark.id"),
		ClientSecret: config.Config.GetString("oauth.client.lark.secret"),
		Scopes:       []string{},
		Endpoint:     endpoints.Lark,
	}
)

// OauthLarkLogin redirect url to lark auth page.
func OauthLarkLogin(c *gin.Context) {
	larkConf.RedirectURL = c.Param("redirect_url")
	// Create oauthState cookie
	oauthState := GenerateStateOauthCookie(c.Writer)
	url := larkConf.AuthCodeURL(oauthState)

	log.Log.Warnln("ClientID: ", larkConf.ClientID)
	log.Log.Warnln("ClientSecret: ", larkConf.ClientSecret)

	log.Log.Warnf("Visit the URL for the auth dialog: %v\n", url)

	c.Redirect(http.StatusPermanentRedirect, url) 
}

// OauthLarkCallback read url from lark callback, 
// get `code`, request app_access_token,
// at last request lark url to get user_access_token.
func OauthLarkCallback(c *gin.Context) {
	oauthState, _ := c.Request.Cookie("oauthstate")
	if c.Request.FormValue("state") != oauthState.Value {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState.Value, c.Request.FormValue("state"))
		c.Redirect(http.StatusFound, "/")
		return
	}

	code := c.Query("code")
	log.Log.Debugf("\ncode ::: %s\n", code)
	accessToken, err := getLarkAppAccessToken()

	if err != nil {
		log.Log.Errorln("getLarkAppAccessToken ::: ", err)
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

	res, err := util.PostWithHeader(UserAccessTokenURL, header, data)
	if err != nil {
		log.Log.Errorln("util.PostWithHeader ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.AccessTokenErr))
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Log.Errorln("io.ReadAll ::: ", err)
		c.JSON(http.StatusInternalServerError, result.Failed(result.HandleError(err)))
		return
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Log.Errorf("gjson.Get ::: response code: %d\n", resCode)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(
			errors.New(fmt.Sprintf("OauthLarkCallback resCode: %d", resCode)),
		)))
		return
	}

	userAccessToken := gjson.Get(string(body), "data.access_token").String()
	expire := gjson.Get(string(body), "data.expire_in").Int()
	model.Rdb.Set(model.RedisCtx, "lark_user_access_token", 
				  userAccessToken, time.Duration(int64(time.Second) * expire))
	// Save needed user info to redis
	model.Rdb.Set(model.RedisCtx, 
				  fmt.Sprintf(
					"lark_%s_avatar_url", 
					gjson.Get(string(body), "data.avatar_url").Str,
				  ),
				  userAccessToken,
				  time.Duration(int64(time.Second) * expire))

	unionId := gjson.Get(string(body), "data.union_id").Str
	user, err := service.UserByLarkUnionID(unionId)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.InternalErr))
		log.Log.Errorln("service.UserByLarkUnionID ::: ", err)
		return
	} else if user == nil {
		// return with oauth lark ticket, which contains "union_id"
		oauth_token, err := util.GenerateTokenWithExp(c, model.OauthSubKey(unionId), model.OAUTH_TICKET_EXP)
		if err != nil {
			c.JSON(http.StatusOK, result.Failed(result.GenerateToken))
			log.Log.Errorln("util.GenerateTokenWithExp ::: ", err)
			return
		}
		c.JSON(http.StatusOK, result.Response{
			Success: false,
			ErrCode: result.OauthUserUnbounded.ErrCode,
			ErrMsg: result.OauthUserUnbounded.ErrMsg,
			Data: gin.H{
				"oauthTicket": oauth_token,
			},
		})
	} else {
		// User already registered and bounded lark,
		// directly return token
		uid := *user.Uid
		token, err := util.GenerateTokenWithExp(c, model.LoginJWTSubKey(uid), model.LOGIN_TOKEN_EXP)
		if err != nil {
			c.JSON(http.StatusOK, result.Failed(result.GenerateToken))
			return
		}
		// model.Rdb.Set(c, model.LoginTokenKey(uid), token, model.LOGIN_TOKEN_EXP)
		c.JSON(http.StatusOK, result.Success(gin.H{
			model.LOGIN_TOKEN_SUB: token,
		}))
	}
}

// Get Lark app_access_token
func getLarkAppAccessToken() (string, error) {
	appId := larkConf.ClientID
	appSecret := larkConf.ClientSecret

	params := url.Values{}
	params.Add("app_id", appId)
	params.Add("app_secret", appSecret)

	res, error := http.PostForm(AppAccessTokenURL, params)
	if error != nil {
		log.Log.Errorln("http.PostForm ::: ", error)
		return "", error
	}
	log.LogRes(res)

	body, error := io.ReadAll(res.Body)
	defer res.Body.Close()
	if error != nil {
		log.Log.Errorln("io.ReadAll ::: ", error)
		return "", error
	}


	if code := gjson.Get(string(body), "code").Int(); code != 0 {
		log.Log.Errorln("gjson.Get ::: code:", code)
		return "", result.InternalErr
	}

	acceToken := gjson.Get(string(body), "app_access_token").String()
	expire := gjson.Get(string(body), "expire").Int()

	model.Rdb.Set(model.RedisCtx, "lark_app_access_token", acceToken, time.Duration(expire))

	return acceToken, nil
}
