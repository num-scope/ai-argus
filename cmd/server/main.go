package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/xtj/ai-argus/api"
	"github.com/xtj/ai-argus/config"
	"github.com/xtj/ai-argus/database"
	"github.com/xtj/ai-argus/internal/service"
	"github.com/xtj/ai-argus/migrations"
	"github.com/xtj/ai-argus/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		log.Printf("AI Argus stopped: %v", err)
		os.Exit(1)
	}
}

func run() (runErr error) {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := logger.Init(cfg.LogLevel, cfg.LogFormat); err != nil {
		return fmt.Errorf("logger init: %w", err)
	}
	defer logger.Sync()
	if err := database.Init(cfg.DatabasePath, cfg.GormLogLevel); err != nil {
		return fmt.Errorf("database init: %w", err)
	}
	defer func() {
		if err := database.Close(); err != nil && runErr == nil {
			runErr = fmt.Errorf("database close: %w", err)
		}
	}()
	if err := migrations.Run(); err != nil {
		return err
	}
	if err := service.ReconcileInterruptedRuns(context.Background()); err != nil {
		return fmt.Errorf("reconcile interrupted runs: %w", err)
	}

	appContext, stopApp := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopApp()
	service.ConfigureRuns(appContext, cfg.MaxConcurrency)
	router, err := api.NewRouter()
	if err != nil {
		return fmt.Errorf("router init: %w", err)
	}
	server := &http.Server{
		Addr:              cfg.Address,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		logger.L().Info("server started", zap.String("address", cfg.Address), zap.String("database", cfg.DatabasePath))
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
		close(serverErrors)
	}()

	select {
	case <-appContext.Done():
	case err := <-serverErrors:
		if err != nil {
			stopApp()
			runErr = fmt.Errorf("serve HTTP: %w", err)
		}
	}

	httpContext, cancelHTTP := context.WithTimeout(context.Background(), 15*time.Second)
	if err := server.Shutdown(httpContext); err != nil && runErr == nil {
		runErr = fmt.Errorf("shutdown HTTP: %w", err)
	}
	cancelHTTP()
	runContext, cancelRuns := context.WithTimeout(context.Background(), 15*time.Second)
	if err := service.ShutdownRuns(runContext); err != nil && runErr == nil {
		runErr = fmt.Errorf("shutdown runs: %w", err)
	}
	cancelRuns()
	return runErr
}
