package v1

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/model/result"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/jackc/pgx/v4"
	pg "github.com/vgarvardt/go-oauth2-pg/v4"
	"github.com/vgarvardt/go-pg-adapter/pgx4adapter"
)

var (
	srv        *server.Server
	pgxConn, _ = pgx.Connect(context.TODO(), config.Config.Sub("Oauth").GetString("db_uri"))
)

func InitServer(c *gin.Context) {
	manager := manage.NewDefaultManager()
	// use PostgreSQL token store with pgx.Connection adapter
	adapter := pgx4adapter.NewConn(pgxConn)
	tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
	defer tokenStore.Close()

	clientStore, _ := pg.NewClientStore(adapter)
	manager.MapTokenStorage(tokenStore)
	manager.MapClientStorage(clientStore)

	clientID, ok := c.GetQuery("client_id")
	if !ok {
		c.JSON(http.StatusBadRequest, result.Failed(result.ClientErr))
		return
	}
	clientSecret, ok := c.GetQuery("client_secret")
	if !ok {
		c.JSON(http.StatusBadRequest, result.Failed(result.ClientErr))
		return
	}
	redirectURI, ok := c.GetQuery("redirect_uri")
	if !ok {
		c.JSON(http.StatusBadRequest, result.Failed(result.ClientErr))
		return
	}

	cErr := clientStore.Create(&models.Client{
		ID:     clientID,
		Secret: clientSecret,
		Domain: redirectURI,
	})
	if cErr != nil {
		c.JSON(http.StatusBadRequest, result.Failed(result.InternalErr))
		return
	}

	srv = server.NewServer(server.NewConfig(), manager)
	//srv.SetPasswordAuthorizationHandler(PasswordAuthorizationHandler)
	//srv.SetUserAuthorizationHandler(userAuthorizeHandler)
	srv.SetClientInfoHandler(server.ClientFormHandler)
}

// redirect user to login for authorization
func Authorize(c *gin.Context) {
	InitServer(c)
	r := c.Request
	w := c.Writer
	store := sessions.Default(c)
	var form url.Values
	v := store.Get("ReturnUri")
	if v != nil {
		form = v.(url.Values)
	}
	r.Form = form

	store.Delete("ReturnUri")
	store.Save()

	// redirect user to login page
	err := srv.HandleAuthorizeRequest(w, r)
	if err != nil {
		c.JSON(http.StatusInternalServerError, result.Failed(result.InternalErr.Wrap(err)))
		return
	}
}
