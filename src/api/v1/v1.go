package v1

import (
	"context"

	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/log"
	"github.com/NJUPT-SAST/sast-link-backend/service"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"go.uber.org/zap"

	"github.com/labstack/echo/v4"

	linkmiddleware "github.com/NJUPT-SAST/sast-link-backend/middleware"
)

// APIV1Service acts as the central service aggregator for version 1 of the API.
// It embeds all specific services related to this API version, facilitating unified access and management.
//
// This service should be utilized within the main application to register all the API routes associated with version 1.
// It is also responsible for handling common HTTP request logic, such as extracting data from the request body,
// processing it according to the business rules defined in the embedded services, and preparing the response.
type APIV1Service struct {
	service.UserService
	service.ProfileService
	service.SysSettingService
	service.OauthService // OauthService is oauth client service

	Store       *store.Store
	Config      *config.Config
	OAuthServer *OAuthServer

	UserLog        *zap.Logger
	ProfileLog     *zap.Logger
	SysSettingLog  *zap.Logger
	OauthServerLog *zap.Logger
}

func NewAPIV1Service(store *store.Store, config *config.Config, oauthServer *OAuthServer) *APIV1Service {
	userLog := log.NewLogger(log.WithModule("user"))
	profileLog := log.NewLogger(log.WithModule("profile"))
	sysSettingLog := log.NewLogger(log.WithModule("system_setting"))
	oauthServerLog := log.NewLogger(log.WithModule("oauth_server"))
	base := service.NewBaseService(store, config)
	// WTF: Are you writing functional programming?
	return &APIV1Service{
		UserService: *service.NewUserService(base.WithOptions(
			service.WithLogger(userLog.WithOptions(log.WithLayer("service"))),
			service.WithStore(store.WithLogger(userLog.WithOptions(log.WithLayer("model")))))),
		ProfileService: *service.NewProfileService(base.WithOptions(
			service.WithLogger(profileLog.WithOptions(log.WithLayer("service"))),
			service.WithStore(store.WithLogger(profileLog.WithOptions(log.WithLayer("model")))))),
		SysSettingService: *service.NewSysSettingService(base.WithOptions(
			service.WithLogger(sysSettingLog.WithOptions(log.WithLayer("service"))),
			service.WithStore(store.WithLogger(sysSettingLog.WithOptions(log.WithLayer("model")))))),
		OauthService: *service.NewOauthService(base.WithOptions(
			service.WithLogger(oauthServerLog.WithOptions(log.WithLayer("service"))),
			service.WithStore(store.WithLogger(oauthServerLog.WithOptions(log.WithLayer("model")))))),

		Store:       store,
		Config:      config,
		OAuthServer: oauthServer,

		UserLog:        userLog.WithOptions(log.WithLayer("api")),
		ProfileLog:     profileLog.WithOptions(log.WithLayer("api")),
		SysSettingLog:  sysSettingLog.WithOptions(log.WithLayer("api")),
		OauthServerLog: oauthServerLog.WithOptions(log.WithLayer("api")),
	}
}

// RegistryRoutes register all routes for API v1.
func (s *APIV1Service) RegistryRoutes(_ context.Context, echoServer *echo.Echo) {
	v1 := echoServer.Group("/api/v1")
	// AuthInterceptor is a middleware that checks the user's authentication status.
	echoServer.Use(linkmiddleware.NewAuthInterceptor(s.Store, s.Config.Secret).AuthenticationInterceptor)

	// Set the rate limit to 3 requests per minute
	// FIXME: 3 request per minute for a user and not all users
	// v1.GET("/sendEmail", s.SendEmail, middleware_link.RequestRateLimiter(3, 1))
	v1.GET("/sendEmail", s.SendEmail)
	v1.GET("/listIDP", s.ListIDPName)
	v1.GET("/idpInfo", s.IDPInfo)
	userGroup := v1.Group("/user")
	{
		userGroup.POST("/register", s.Register)
		userGroup.POST("/login", s.Login)
		userGroup.GET("/loginWithSSO", s.LoginWithSSO)
		userGroup.POST("/logout", s.Logout)
		userGroup.GET("/info", s.UserInfo)
		userGroup.POST("/changePassword", s.ChangePassword)
		userGroup.POST("/resetPassword", s.ResetPassword)
		userGroup.POST("/bindEmailWithSSO", s.BindEmailWithSSO)
	}
	verifyGroup := v1.Group("/verify")
	{
		verifyGroup.GET("/account", s.Verify)
		verifyGroup.GET("/verifyEmail", s.VerifyEmail)
		verifyGroup.POST("/verifyCode", s.CheckVerifyCode)
	}

	oauth := v1.Group("/oauth2")
	{
		// oauth2 authorize
		oauth.GET("/authorize", s.Authorize)
		oauth.POST("/token", s.AccessToken)
		oauth.POST("/refresh", s.RefreshToken)
		oauth.POST("/createClient", s.CreateClient)
		oauth.POST("/addCallback", s.AddRedirectURI)
		oauth.GET("/userinfo", s.OauthUserInfo)
		oauth.GET("/listClient", s.ListClient)
		oauth.POST("/updateClient", s.UpdateClient)
		oauth.POST("/deleteClient", s.DelClient)
		oauth.GET("/client", s.GetClient)
	}
	profileGroup := v1.Group("/profile")
	{
		profileGroup.GET("/getProfile", s.GetProfile)
		profileGroup.GET("/bindStatus", s.BindStatus)
		profileGroup.POST("/change", s.ChangeProfile)
		profileGroup.POST("/changeAvatar", s.UploadAvatar)
		profileGroup.POST("/dealCensorRes", s.DealCensorRes)
	}
	systemSettingGroup := v1.Group("/systemSetting")
	{
		systemSettingGroup.GET("/:setting_type", s.SystemSetting)
		systemSettingGroup.POST("/:setting_type", s.UpsetSystemSetting)
	}
}
