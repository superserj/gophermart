package repository

import (
	"context"
	"database/sql"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DBStorage — PostgreSQL-хранилище на pgxpool.
type DBStorage struct {
	pool *pgxpool.Pool
}

// NewDBStorage прогоняет миграции и открывает пул соединений.
func NewDBStorage(ctx context.Context, dsn string) (*DBStorage, error) {
	if err := migrate(dsn); err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &DBStorage{pool: pool}, nil
}

func migrate(dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

// Ping проверяет доступность БД.
func (s *DBStorage) Ping(ctx context.Context) error { return s.pool.Ping(ctx) }

// Close закрывает пул.
func (s *DBStorage) Close() error {
	s.pool.Close()
	return nil
}
