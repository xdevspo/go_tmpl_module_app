package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	container "github.com/xdevspo/go_tmpl_module_app/internal/core/container"
)

type HTTPServer struct {
	sp     *container.ServiceProvider
	router *gin.Engine
	server *http.Server
}

func NewHTTPServer(sp *container.ServiceProvider, router Router) *HTTPServer {
	if sp.AppConfig().Env() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	router.SetupRouter(engine)

	return &HTTPServer{
		sp:     sp,
		router: engine,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%s", sp.HTTPConfig().Port()),
			Handler: engine,
		},
	}
}

func (s *HTTPServer) Run() error {
	logrus.Infof("Starting HTTP server on port %s", s.sp.HTTPConfig().Port())
	return s.server.ListenAndServe()
}

func (s *HTTPServer) Shutdown(ctx context.Context) error {
	logrus.Info("Shutting down HTTP server...")
	return s.server.Shutdown(ctx)
}
