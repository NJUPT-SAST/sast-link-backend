package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/NJUPT-SAST/sast-link-backend/api/v1"
	"github.com/NJUPT-SAST/sast-link-backend/config"
	"github.com/NJUPT-SAST/sast-link-backend/store"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Server struct {
	Profile *config.Config
	Store   *store.Store

	echoServer *echo.Echo
}

// NewServer create a server instance with configuration and store.
func NewServer(ctx context.Context, profile *config.Config, store *store.Store) (*Server, error) {
	s := &Server{
		Profile: profile,
		Store:   store,
	}

	echoServer := echo.New()
	echoServer.Debug = true
	echoServer.HideBanner = true
	echoServer.HidePort = true
	echoServer.Use(middleware.Recover())
	s.echoServer = echoServer

	echoServer.GET("/healthcheck", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	oauthServer, err := v1.NewOAuthServer(ctx, profile, *store)
	if err != nil {
		fmt.Printf("failed to create oauth server: %v\n", err)
		return nil, err
	}

	apiV1Service := v1.NewAPIV1Service(store, profile, oauthServer)
	// Adding routes must beforer echo server start.
	apiV1Service.RegistryRoutes(ctx, echoServer)

	return s, nil
}

func (s *Server) Start() error {
	go func() {
		if err := s.echoServer.Start(fmt.Sprintf("%s:%d", "0.0.0.0", s.Profile.Port)); err != nil {
			if err != http.ErrServerClosed {
				fmt.Printf("failed to start echo server: %v\n", err)
			}
		}
	}()

	return nil
}

// Shutdown shutdown the echo server and close the database connection.
func (s *Server) Shutdown(ctx context.Context) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	// Shutdown echo server.
	if err := s.echoServer.Shutdown(ctx); err != nil {
		fmt.Printf("failed to shutdown server, error: %v\n", err)
	}

	// Close database connection.
	if err := s.Store.Close(); err != nil {
		fmt.Printf("failed to close database, error: %v\n", err)
	}

	fmt.Printf("SAST Link is shutdown, bye\n")
}
