package v1

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	reqLog "github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/model"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/jackc/pgx/v4"
	"github.com/sirupsen/logrus"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
)

var (
	srv          *server.Server
	pgxConn, _   = pgx.Connect(context.Background(), config.Config.Sub("oauth.server").GetString("db_uri"))
	tokenAdapter = pgx4adapter.NewConn(pgxConn)
	// FIXME: tokenStore, clientStore maybe have some problem
	tokenStore, _  = pg.NewTokenStore(tokenAdapter, pg.WithTokenStoreGCInterval(time.Minute))
	clientAdapter  = pgx4adapter.NewConn(pgxConn)
	clientStore, _ = pg.NewClientStore(clientAdapter)
)

// ClientStoreItem data item
type ClientStoreItem struct {
	ID     string `db:"id"`
	Secret string `db:"secret"`
	Domain string `db:"domain"`
	Data   []byte `db:"data"`
}

func init() {
	InitServer()
}

func InitServer() {
	mg := manage.NewDefaultManager()
	mg.MapTokenStorage(tokenStore)
	mg.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
	// use PostgreSQL client store with pgx.Connection adapter
	mg.MapClientStorage(clientStore)

	srv = server.NewServer(server.NewConfig(), mg)
	srv.SetClientInfoHandler(clientInfoHandler)
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	// TODO: error handler
	srv.SetInternalErrorHandler(InternalErrorHandler)
	srv.SetResponseErrorHandler(ResponseErrorHandler)

	srv.SetResponseTokenHandler(ResponseTokenHandler)
}

func InternalErrorHandler(err error) (re *errors.Response) {
	log.Log.Errorf("Oauth2 ::: InternalErrorHandler:[%s]", err.Error())
	error := errors.NewResponse(err, http.StatusInternalServerError)
	error.ErrorCode = 500
	error.StatusCode = http.StatusInternalServerError
	error.Description = err.Error()
	return error
}

func ResponseErrorHandler(re *errors.Response) {
	log.Log.Errorf("Oauth2 ::: ResponseErrorHandler:[%s]", re.Error.Error())
}

func ResponseTokenHandler(w http.ResponseWriter, data map[string]interface{}, header http.Header, statusCode ...int) error {
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Pragma", "no-cache")

	for key := range header {
		w.Header().Set(key, header.Get(key))
	}

	status := http.StatusOK
	if len(statusCode) > 0 && statusCode[0] > 0 {
		status = statusCode[0]
	}

	log.Log.Errorf("Oauth2 ::: ResponseTokenHandler:data:[%s]", data)
	log.Log.Errorf("Oauth2 ::: ResponseTokenHandler:status:[%d]", status)
	log.Log.Errorf("Oauth2 ::: ResponseTokenHandler:header:[%s]", header)

	w.WriteHeader(status)
	if data["error"] != nil {
		var errMsg string
		var errCode int
		if data["error_description"] != nil {
			errMsg = data["error_description"].(string)
		}
		if data["error_code"] != nil {
			errCode = data["error_code"].(int)
		}
		error := result.LocalError{errCode, errMsg, nil}
		log.Log.Errorf("Oauth2 ::: ResponseTokenHandler:error:[%s]", error)
		return json.NewEncoder(w).Encode(result.Failed(error))
	} else {
		return json.NewEncoder(w).Encode(result.Success(data))
	}
}

// Create client
func CreateClient(c *gin.Context) {
	redirectURI := c.PostForm("redirect_uri")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, result.Failed(result.RequestParamError))
		return
	}

	token := c.GetHeader("TOKEN")
	uid, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	if err != nil || uid == "" {
		c.JSON(http.StatusOK, result.Failed(result.TokenError))
		return
	}

	clientID := util.GenerateUUID()
	secret, err := util.GenerateRandomString(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr))
		return
	}

	cErr := clientStore.Create(&models.Client{
		ID:     clientID,
		Secret: secret,
		Domain: redirectURI,
		UserID: uid,
	})

	if cErr != nil {
		c.JSON(http.StatusBadRequest, result.Failed(result.InternalErr))
		return
	}

	c.JSON(http.StatusOK, result.Success(gin.H{
		"client_id":     clientID,
		"client_secret": secret,
	}))
}

func OauthUserInfo(c *gin.Context) {
	// Bearer
	bearerToken := c.GetHeader("Authorization")
	if bearerToken == "" ||
		!strings.HasPrefix(bearerToken, "Bearer ") {
		c.JSON(http.StatusOK, result.Failed(result.AccessTokenErr))
		return
	}
	accessToken := strings.Split(bearerToken, " ")[1]
	mg := srv.Manager
	ti, err := mg.LoadAccessToken(c, accessToken)
	if err != nil {
		c.JSON(http.StatusOK, result.Failed(result.AccessTokenErr))
		return
	}
	// TODO: scope check
	ti.GetScope()

	user, err := service.OauthUserInfo(ti.GetUserID())
	if err != nil {
		controllerLogger.WithFields(
			logrus.Fields{
				"username": user.Uid,
			}).Error(err)
		c.JSON(http.StatusOK, result.Failed(result.GetUserinfoFail))
		return
	}

	profileInfo, serErr := service.GetProfileInfo(*user.Uid)
	if serErr != nil {
		controllerLogger.Errorln("GetProfile service wrong", serErr)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(serErr)))
		return
	}

	if dep, org, getOrgErr := service.GetProfileOrg(profileInfo.OrgId); getOrgErr != nil {
		controllerLogger.Errorln("GetProfileOrg Err", getOrgErr)
		c.JSON(http.StatusOK, result.Failed(result.HandleError(getOrgErr)))
		return
	} else {
		c.JSON(http.StatusOK, result.Success(gin.H{
			"nickname": profileInfo.Nickname,
			"userId":   user.Uid,
			"dep":      dep,
			"org":      org,
			"email":    profileInfo.Email,
			"avatar":   profileInfo.Avatar,
			"bio":      profileInfo.Bio,
			"link":     profileInfo.Link,
			"badge":    profileInfo.Badge,
			"hide":     profileInfo.Hide,
		}))
		return
	}
}

func Authorize(c *gin.Context) {
	r := c.Request
	w := c.Writer
	// store, err := session.Start(c, w, r)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
	// 	return
	// }
	_ = r.ParseForm()
	// var form url.Values
	// if v, ok := store.Get("ReturnUri"); ok {
	// 	form = v.(url.Values)
	// }
	// r.Form = form
	// store.Delete("ReturnUri")
	// _ = store.Save()

	// Redirect user to login page if user not login or
	// Get code directly if user has logged in
	reqLog.LogReq(r)
	err := srv.HandleAuthorizeRequest(w, r)
	clients, _ := srv.Manager.GetClient(r.Context(), r.Form.Get("client_id"))
	log.Log.Println(clients)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
}

// User decides whether to authorize
// func UserAuth(c *gin.Context) {
// 	w := c.Writer
// 	r := c.Request
//
// 	//token := r.Header.Get("TOKEN")
// 	_ = r.ParseMultipartForm(0)
// 	token := c.PostForm("token")
// 	if token == "" {
// 		w.Header().Set("Content-Type", "application/json")
// 		response := result.Failed(result.AuthError)
// 		json, _ := json.Marshal(response)
// 		w.Write(json)
// 		return
// 	}
// }

// Get AccessToken
func AccessToken(c *gin.Context) {
	w := c.Writer
	r := c.Request
	err := srv.HandleTokenRequest(w, r)
	id, _, _ := clientInfoHandler(r)
	var item ClientStoreItem
	// TODO: DEBUG
	if err := clientAdapter.SelectOne(c, &item, fmt.Sprintf("SELECT * FROM %s WHERE id = $1", "oauth2_clients"), id); err != nil {
		log.Log.Errorf("----DEBUG----: %s", err.Error())
		log.Log.Printf("\nitem: %v\n", item)
		return
	}

	// FIXME: err is always nil
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
}

// Refresh AccessToken
func RefreshToken(c *gin.Context) {
	w := c.Writer
	r := c.Request
	err := srv.HandleTokenRequest(w, r)
	if err == nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
}

func clientInfoHandler(r *http.Request) (clientID, clientSecret string, err error) {
	_ = r.ParseMultipartForm(0)
	_ = r.ParseForm()
	if r.Form.Get("grant_type") == "refresh_token" {
		ti, err := srv.Manager.LoadRefreshToken(r.Context(), r.Form.Get("refresh_token"))
		if err != nil {
			return "", "", result.RefreshTokenErr
		}
		clientID = ti.GetClientID()
		if clientID == "" {
			return "", "", result.ClientErr
		}
		cli, err := srv.Manager.GetClient(r.Context(), clientID)
		if err != nil {
			return "", "", result.ClientErr
		}
		clientSecret = cli.GetSecret()
		if clientSecret == "" {
			return "", "", result.ClientErr
		}
		return clientID, clientSecret, nil
	}
	clientID = r.Form.Get("client_id")
	if clientID == "" {
		return "", "", result.ClientErr
	}
	clientSecret = r.Form.Get("client_secret")
	if clientSecret == "" {
		return "", "", result.ClientErr
	}
	return clientID, clientSecret, nil

}

func getTokenByUUID(c context.Context, uuid string) (token string, err error) {
	token, err = model.Rdb.Get(c, uuid).Result()
	if err != nil {
		log.Log.Errorln("invalid uuid, can't find token")
		return "", err
	}
	return token, nil
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	// session, err := session.Start(r.Context(), w, r)
	// if err != nil {
	// 	return
	// }

	token := r.Form.Get("part")
	if token == "" {
		if r.Form == nil {
			// check if user is logged in
			_ = r.ParseMultipartForm(0)
			_ = r.ParseForm()
		}

		// session.Set("ReturnUri", r.Form)
		// _ = session.Save()

		w.Header().Set("Content-Type", "application/json")
		response := result.Failed(result.TokenError)
		log.Log.Errorln("Oauth2 ::: token is empty")
		json, _ := json.Marshal(response)
		w.Write(json)
		return
	}

	username, err := util.GetUsername(token, model.LOGIN_TOKEN_SUB)
	log.Log.Println("Oauth2 ::: username: ", username)
	if err != nil || username == "" {
		if r.Form == nil {
			_ = r.ParseForm()
		}

		// session.Set("ReturnUri", r.Form)
		// _ = session.Save()

		w.Header().Set("Content-Type", "application/json")
		response := result.Failed(result.TokenError)
		log.Log.Errorln("Oauth2 ::: token is invalid")
		json, _ := json.Marshal(response)
		w.Write(json)
		return
	}

	rToken, err := model.Rdb.Get(r.Context(), model.LoginTokenKey(username)).Result()
	if err != nil {
		if r.Form == nil {
			_ = r.ParseForm()
		}

		// session.Set("ReturnUri", r.Form)
		// _ = session.Save()

		w.Header().Set("Content-Type", "application/json")
		response := result.Failed(result.TokenError)
		log.Log.Errorln("Oauth2 ::: token is invalid")
		json, _ := json.Marshal(response)
		w.Write(json)
		return
	}
	if rToken != token {
		if r.Form == nil {
			_ = r.ParseForm()
		}

		// session.Set("ReturnUri", r.Form)
		// _ = session.Save()

		w.Header().Set("Content-Type", "application/json")
		response := result.Failed(result.TokenError)
		log.Log.Errorln("Oauth2 ::: token is invalid")
		json, _ := json.Marshal(response)
		w.Write(json)
		return
	}
	return username, nil
}
