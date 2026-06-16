package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/accrual"
	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/config"
	"github.com/superserj/gophermart/internal/handler"
	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/repository"
)

func main() {
	cfg, err := config.New()
	if err != nil {
		log.Fatal(err)
	}
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal(err)
	}
	if cfg.DatabaseURI == "" {
		logger.Log.Fatal("DATABASE_URI is required")
	}

	repo, err := repository.NewDBStorage(context.Background(), cfg.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("init repository", zap.Error(err))
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			logger.Log.Error("close repository", zap.Error(cerr))
		}
	}()

	secret, err := config.ResolveSecret(cfg.AuthSecret)
	if err != nil {
		logger.Log.Fatal("resolve auth secret", zap.Error(err))
	}
	if cfg.AuthSecret == "" {
		logger.Log.Warn("AUTH_SECRET не задан, используется эфемерный случайный секрет")
	}
	a := auth.New(secret)
	h := handler.New(repo, a)

	pollCtx, pollCancel := context.WithCancel(context.Background())
	poller := accrual.NewPoller(repo, accrual.NewClient(cfg.AccrualSystemAddress))
	pollDone := make(chan struct{})
	go func() {
		poller.Run(pollCtx)
		close(pollDone)
	}()

	srv := &http.Server{Addr: cfg.RunAddress, Handler: handler.NewRouter(h, a)}
	go func() {
		logger.Log.Info("starting server", zap.String("addr", cfg.RunAddress))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("listen and serve", zap.Error(err))
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-sigCtx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("server shutdown", zap.Error(err))
	}

	pollCancel()
	<-pollDone
}
