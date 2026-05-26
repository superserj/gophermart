package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/superserj/gophermart/internal/config"
	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/middleware"
	"github.com/superserj/gophermart/internal/repository"
)

type pinger interface {
	Ping(ctx context.Context) error
}

func pingHandler(p pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := p.Ping(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

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

	r := chi.NewRouter()
	r.Use(logger.WithLogging)
	r.Use(middleware.Gzip)
	r.Get("/ping", pingHandler(repo))

	srv := &http.Server{Addr: cfg.RunAddress, Handler: r}
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
}
