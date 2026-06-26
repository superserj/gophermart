package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/superserj/gophermart/internal/accrual"
	"github.com/superserj/gophermart/internal/auth"
	"github.com/superserj/gophermart/internal/config"
	"github.com/superserj/gophermart/internal/handler"
	"github.com/superserj/gophermart/internal/logger"
	"github.com/superserj/gophermart/internal/repository"
)

const (
	shutdownTimeout   = 5 * time.Second
	readHeaderTimeout = 10 * time.Second // защита от slowloris (gosec G112)
	readTimeout       = 30 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

// run собирает зависимости и обслуживает сервис до сигнала завершения.
// Единственная точка выхода — main(): все ошибки возвращаются наружу, чтобы
// defer-функции (в т.ч. закрытие пула БД) гарантированно отработали.
func run() error {
	cfg, err := config.New()
	if err != nil {
		return err
	}
	lg, err := logger.New(cfg.LogLevel)
	if err != nil {
		return err
	}
	defer func() { _ = lg.Sync() }()

	if cfg.DatabaseURI == "" {
		return errors.New("DATABASE_URI is required")
	}
	secret, err := config.ResolveSecret(cfg.AuthSecret)
	if err != nil {
		return fmt.Errorf("resolve auth secret: %w", err)
	}
	if cfg.AuthSecret == "" {
		lg.Warn("AUTH_SECRET не задан, используется эфемерный случайный секрет")
	}

	repo, err := repository.NewDBStorage(context.Background(), cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("init repository: %w", err)
	}
	defer func() {
		if cerr := repo.Close(); cerr != nil {
			lg.Error("close repository", zap.Error(cerr))
		}
	}()

	a := auth.New(secret)
	h := handler.New(repo, a, lg)
	srv := &http.Server{
		Addr:              cfg.RunAddress,
		Handler:           handler.NewRouter(h, a),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	poller := accrual.NewPoller(
		repo,
		accrual.NewClient(cfg.AccrualSystemAddress, lg.With(zap.String("component", "accrual"))),
		lg.With(zap.String("component", "poller")),
	)

	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	g, gctx := errgroup.WithContext(sigCtx)
	g.Go(func() error {
		lg.Info("starting server", zap.String("addr", cfg.RunAddress))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		poller.Run(gctx)
		return nil
	})
	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}
		return nil
	})
	return g.Wait()
}
