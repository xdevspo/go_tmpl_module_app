package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xdevspo/go_tmpl_module_app/internal/core/closer"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/config"
	container "github.com/xdevspo/go_tmpl_module_app/internal/core/container"
	"github.com/xdevspo/go_tmpl_module_app/internal/core/server"
)

type App struct {
	sp *container.ServiceProvider
}

func NewApp(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.Init(ctx); err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	ctx := context.Background()
	logger := a.sp.Logger()

	httpPort := a.sp.HTTPConfig().Port()

	r := server.SetupRouter(ctx, a.sp)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", httpPort),
		Handler: r,
	}

	go func() {
		logger.Info("Starting HTTP server on port %s", httpPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("Error starting server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		return err
	}

	logger.Info("Server exited gracefully")
	return nil
}

func (a *App) Init(ctx context.Context) error {
	inits := []func(ctx context.Context) error{
		a.initConfig,
		a.initServiceProvider,
	}

	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *App) initConfig(_ context.Context) error {
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}

	configFile := fmt.Sprintf(".env.%s", env)
	err := config.Load(configFile)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.sp = container.NewServiceProvider()
	return nil
}
