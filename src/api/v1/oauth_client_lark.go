package v1

import (
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
	AppAccessTokenURL  = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
	UserAccessTokenURL = "https://open.feishu.cn/open-apis/authen/v1/oidc/access_token"
	UserInfoBasicURL   = "https://open.feishu.cn/open-apis/authen/v1/user_info"
	UserInfoDetailURL  = "https://open.feishu.cn/open-apis/contact/v3/users/"
)

var (
	larkConf = oauth2.Config{
		ClientID:     config.Config.GetString("oauth.client.lark.id"),
		ClientSecret: config.Config.GetString("oauth.client.lark.secret"),
		Scopes:       []string{"contact:contact.base:readonly", "contact:user.base:readonly", "contact:user.department:readonly", "contact:user.department_path:readonly"},
		Endpoint:     endpoints.Lark,
	}
)

// OauthLarkLogin redirect url to lark auth page.
func OauthLarkLogin(c *gin.Context) {
	redirectURL := c.Query("redirect_url")
	log.Debugf("redirectURL ::: %s\n", redirectURL)
	larkConf.RedirectURL = redirectURL
	// Create oauthState cookie
	oauthState := GenerateStateOauthCookie(c.Writer)
	log.Debugf("oauthState ::: %s\n", redirectURL)

	url := larkConf.AuthCodeURL(oauthState)

	log.Debug("ClientID: ", larkConf.ClientID)
	log.Debug("ClientSecret: ", larkConf.ClientSecret)

	log.Log.Warnf("Visit the URL for the auth dialog: %v\n", url)

	c.Redirect(http.StatusFound, url)
}

// OauthLarkCallback read url from lark callback,
// get `code`, request app_access_token,
// then request lark url to get user_access_token.
// at last request user info
func OauthLarkCallback(c *gin.Context) {
	oauthState, _ := c.Request.Cookie("oauthstate")
	log.Debugf("oauthState ::: %v\n", oauthState)
	if c.Request.FormValue("state") != oauthState.Value {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", oauthState.Value, c.Request.FormValue("state"))
		c.Redirect(http.StatusFound, "/")
		return
	}

	code := c.Query("code")
	log.Debugf("\ncode ::: %s\n", code)

	accessToken, err := larkAppAccessToken()
	if err != nil {
		log.Error("larkAppAccessToken ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	userAccessTokenBody, err := larkUserAccessToken(code, accessToken)
	if err != nil {
		log.Error("larkUserAccessToken ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	userAccessToken := gjson.Get(userAccessTokenBody, "data.access_token").String()

	userInfoBasicBody, err := larkUserInfoBasic(userAccessToken)
	if err != nil {
		log.Error("larkUserInfoBasic ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(err)))
		return
	}

	openId := gjson.Get(userInfoBasicBody, "data.open_id").Str
	unionId := gjson.Get(userInfoBasicBody, "data.union_id").Str

	userInfoDetailBody, err := larkUserInfoDetail(openId, userAccessToken)

	// save user info in redis (then retrive in login)
	userInfo := gjson.Get(userInfoDetailBody, "data").String()
	if err := model.Rdb.Set(c, unionId,
		userInfo, time.Duration(model.OAUTH_USER_INFO_EXP)).Err(); err != nil {
		log.Error("model.Rdb.Set ::: ", err)
		c.JSON(http.StatusOK, result.Failed(result.InternalErr))
		return
	}

	userLarkInfo, err := service.OauthInfoByLarkID(unionId)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.InternalErr))
		log.Error("service.UserByLarkUnionID ::: ", err)
		return
	} else if userLarkInfo == nil {
		// return with oauth lark ticket, which contains "union_id"
		oauthToken, err := util.GenerateTokenWithExp(c, model.OauthSubKey(unionId, model.OAUTH_LARK_SUB), model.OAUTH_TICKET_EXP)

		if err != nil {
			c.JSON(http.StatusOK, result.Failed(result.GenerateToken))
			log.Error("util.GenerateTokenWithExp ::: ", err)
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
		// User already registered and bounded lark,
		// directly return token
		uid := userLarkInfo.UserID
		log.Debugf("User already registered and bounded lark: %s", uid)
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

// Get Lark app_access_token
func larkAppAccessToken() (string, error) {
	appId := larkConf.ClientID
	appSecret := larkConf.ClientSecret

	params := url.Values{}
	params.Add("app_id", appId)
	params.Add("app_secret", appSecret)

	res, error := http.PostForm(AppAccessTokenURL, params)
	if error != nil {
		log.Error("http.PostForm ::: ", error)
		return "", error
	}
	// log.LogRes(res)

	body, error := io.ReadAll(res.Body)
	defer res.Body.Close()
	if error != nil {
		log.Error("io.ReadAll ::: ", error)
		return "", error
	}

	if code := gjson.Get(string(body), "code").Int(); code != 0 {
		log.Error("gjson.Get ::: code:", code)
		return "", result.InternalErr
	}

	acceToken := gjson.Get(string(body), "app_access_token").String()

	return acceToken, nil
}

func larkUserAccessToken(code string, accessToken string) (string, error) {
	data := map[string]string{
		"grant_type": "authorization_code",
		"code":       code,
	}

	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
		"Content-Type":  "application/json; charset=utf-8",
	}

	res, err := util.PostWithHeader(UserAccessTokenURL, header, data)
	if err != nil {
		log.Log.Errorln("util.PostWithHeader ::: ", err)
		return "", result.AccessTokenErr
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error("io.ReadAll ::: ", err)
		return "", result.InternalErr
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Errorf("larkUserAccessToken ::: gjson.Get ::: response code: %d\n", resCode)
		return "", fmt.Errorf("OauthLarkCallback resCode: %d", resCode)
	}
	return string(body), nil
}

// Get userinfo using user_access_token
func larkUserInfoBasic(userAccessToken string) (string, error) {
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", userAccessToken),
	}
	res, err := util.GetWithHeader(UserInfoBasicURL, header)
	if err != nil {
		log.Error("util.GetWithHeader ::: ", err)
		return "", result.AccessTokenErr
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error("io.ReadAll ::: ", err)
		return "", result.InternalErr
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Errorf("larkUserInfoBasic ::: gjson.Get ::: response code: %d\n", resCode)
		return "", fmt.Errorf("OauthLarkCallback resCode: %d", resCode)
	}
	return string(body), nil
}

// get user detail info
func larkUserInfoDetail(userId string, userAccessToken string) (string, error) {
	header := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", userAccessToken),
	}
	res, err := util.GetWithHeader(UserInfoDetailURL+userId, header)
	if err != nil {
		log.Error("util.GetWithHeader ::: ", err)
		return "", result.AccessTokenErr
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Error("io.ReadAll ::: ", err)
		return "", result.InternalErr
	}
	if resCode := gjson.Get(string(body), "code").Int(); resCode != 0 {
		log.Errorf("larkUserInfoDetail ::: gjson.Get ::: response code: %d\n", resCode)
		return "", fmt.Errorf("OauthLarkCallback resCode: %d", resCode)
	}
	return string(body), nil
}
