package v1

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/NJUPT-SAST/sast-link-backend/util"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-session/session"
	"github.com/jackc/pgx/v4"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
)

var (
	srv            *server.Server
	pgxConn, _     = pgx.Connect(context.TODO(), config.Config.Sub("oauth").GetString("db_uri"))
	adapter        = pgx4adapter.NewConn(pgxConn)
	clientStore, _ = pg.NewClientStore(adapter)
)

//func newManager() (manager *manage.Manager) {
//	// use PostgreSQL token store with pgx.Connection adapter
//	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
//	defer tokenStore.Close()
//
//	mg := manage.NewDefaultManager()
//	mg.MapTokenStorage(tokenStore)
//	mg.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)
//
//	return mg
//}

func InitServer(c *gin.Context) {
	// use PostgreSQL token store with pgx.Connection adapter
	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
	defer tokenStore.Close()

	mg := manage.NewDefaultManager()
	mg.MapTokenStorage(tokenStore)
	mg.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

	// use PostgreSQL client store with pgx.Connection adapter
	mg.MapClientStorage(clientStore)

	//clientID, ok := c.GetQuery("client_id")
	//if !ok {
	//	c.JSON(http.StatusBadRequest, result.Failed(result.ClientErr))
	//	return
	//}
	//if !ok {
	//	c.JSON(http.StatusBadRequest, result.Failed(result.ClientErr))
	//	return
	//}

	//client, cErr := clientStore.GetByID(c, clientID)
	//if cErr != nil {
	//	fmt.Println(cErr)
	//	return
	//}
	//fmt.Println(client)

	//cErr := clientStore.Create(&models.Client{
	//	ID:     clientID,
	//	Secret: "test",
	//	Domain: redirectURI,
	//})
	//if cErr != nil {
	//	fmt.Println(cErr)
	//	c.JSON(http.StatusBadRequest, result.Failed(result.InternalErr))
	//	return
	//}

	srv = server.NewServer(server.NewConfig(), mg)
	//srv.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler)
	//srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	//srv.SetClientInfoHandler(server.ClientFormHandler)
	srv.SetUserAuthorizationHandler(userAuthorizeHandler)
}

// Create client
func CreateClient(c *gin.Context) {

	redirectURI := c.PostForm("redirect_uri")
	if redirectURI == "" {
		c.JSON(http.StatusBadRequest, result.Failed(result.ParamError))
		return
	}

	cErr := clientStore.Create(&models.Client{
		ID: util.GenerateUUID(),
		//Secret: ,
		Domain: redirectURI,
	})
	if cErr != nil {
		c.JSON(http.StatusBadRequest, result.Failed(result.InternalErr))
		return
	}
	// TODO
	c.JSON(http.StatusOK, result.Success(""))
}

// redirect user to login for authorization
func Authorize(c *gin.Context) {
	InitServer(c)
	r := c.Request
	w := c.Writer
	store, err := session.Start(context.Background(), w, r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
	var form url.Values
	if v, ok := store.Get("ReturnUri"); ok {
		form = v.(url.Values)
	}
	r.Form = form

	store.Delete("ReturnUri")
	_ = store.Save()

	fmt.Print(srv)
	// redirect user to login page
	err = srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
}

// User decides whether to authorize
func UserAuth(c *gin.Context) {
	w := c.Writer
	r := c.Request

	token := r.Header.Get("TOKEN")
	if token == "" {
		w.Header().Set("Location", "/api/v1/verify/account")
		w.WriteHeader(http.StatusFound)
		return
	}
}

func userAuthorizeHandler(w http.ResponseWriter, r *http.Request) (userID string, err error) {
	session, err := session.Start(context.Background(), w, r)
	//session := sessions.Default(c)
	if err != nil {
		return
	}

	// check if user is logged in
	token := r.Header.Get("TOKEN")
	if token == "" {
		w.Header().Set("Location", "/api/v1/verify/account")
		w.WriteHeader(http.StatusFound)
		return
	}
	username, err := util.GetUsername(token)
	if err != nil || username == "" {
		if r.Form == nil {
			_ = r.ParseForm()
		}

		session.Set("ReturnUri", r.Form)
		_ = session.Save()

		w.Header().Set("Location", "/api/v1/verify/account")
		w.WriteHeader(http.StatusFound)
		return
	}

	return username, nil
}
