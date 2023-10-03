package v1

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/endpoints"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"golang.org/x/oauth2"
)

const (
	// Lark
	LarkAppAccessTokenURL = "https://open.feishu.cn/open-apis/auth/v3/app_access_token/internal"
)

var (
	larkConf = oauth2.Config{
		ClientID:     "cli_a591242f4c3ad00b",
		ClientSecret: "dX7nGrIJE2flCg5zq24mKeKaqwxjqkie",
		RedirectURL:  "http://localhost:8080/api/v1/login/lark/callback",
		Scopes:       []string{},
		Endpoint:     endpoints.Lark,
	}

	githubConf = oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		RedirectURL:  "YOUR_REDIRECT_URL",
		Scopes:       []string{},
		Endpoint:     endpoints.GitHub,
	}

	qqConf = oauth2.Config{
		ClientID:     "YOUR_CLIENT_ID",
		ClientSecret: "YOUR_CLIENT_SECRET",
		RedirectURL:  "YOUR_REDIRECT_URL",
		Scopes:       []string{},
		Endpoint:     endpoints.QQ,
	}
)

func OauthLarkLogin(c *gin.Context) {
	url := larkConf.AuthCodeURL("state")

	fmt.Printf("Visit the URL for the auth dialog: %v\n", url)

	c.Redirect(http.StatusFound, url)
}

func OauthLarkCallback(c *gin.Context) {
	// code := c.Query("code")
	// accessToken, err := getLarkAppAccessToken(c.Request.Context())
	//
	// 
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, result.Failed(result.HandleError(err)))
	// 	return
	// }
	//
	// token, err := larkConf.Exchange(c.Request.Context(), code)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, result.Failed(result.HandleError(err)))
	// 	return
	// }
	// c.JSON(http.StatusOK, result.Success(token))
}

// Get Lark app_access_token
func getLarkAppAccessToken(ctx context.Context) (string, error) {
	appId := larkConf.ClientID
	appSecret := larkConf.ClientSecret
	params := url.Values{}
	params.Add("app_id", appId)
	params.Add("app_secret", appSecret)

	req, error := http.PostForm(LarkAppAccessTokenURL, params)
	if error != nil {
		fmt.Println("get app_access_token error: ", error)
		return "", error
	}

	body, error := io.ReadAll(req.Body)
	defer req.Body.Close()
	if error != nil {
		fmt.Println("get app_access_token error: ", error)
		return "", error
	}

	fmt.Println("get app_access_token response: ", string(body))

	if gjson.Get(string(body), "code").Int() != 0 {
		fmt.Println("get app_access_token error: ", string(body))
		return "", result.InternalErr
	}

	acceToken := gjson.Get(string(body), "app_access_token")
	expire := gjson.Get(string(body), "expire")

	model.Rdb.Set(ctx, "lark_app_access_token", acceToken.String(), time.Duration(expire.Int()))

	return acceToken.String(), nil
}
