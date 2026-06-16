// Package config разбирает флаги и ENV-переменные сервиса.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
)

// Config — параметры запуска. ENV приоритетнее флагов.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	AuthSecret           string
	LogLevel             string
}

// Parse разбирает args и применяет ENV-override через lookup.
func Parse(args []string, lookup func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", "localhost:8080", "address to run HTTP server")
	fs.StringVar(&cfg.DatabaseURI, "d", "", "PostgreSQL DSN")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", "", "accrual system address")
	fs.StringVar(&cfg.AuthSecret, "s", "", "secret key for auth signature (random ephemeral if empty)")
	fs.StringVar(&cfg.LogLevel, "l", "info", "log level")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	if v, ok := lookup("RUN_ADDRESS"); ok {
		cfg.RunAddress = v
	}
	if v, ok := lookup("DATABASE_URI"); ok {
		cfg.DatabaseURI = v
	}
	if v, ok := lookup("ACCRUAL_SYSTEM_ADDRESS"); ok {
		cfg.AccrualSystemAddress = v
	}
	if v, ok := lookup("AUTH_SECRET"); ok {
		cfg.AuthSecret = v
	}
	if v, ok := lookup("LOG_LEVEL"); ok {
		cfg.LogLevel = v
	}
	return cfg, nil
}

// New — продакшн-вход: реальные os.Args и os.LookupEnv.
func New() (*Config, error) { return Parse(os.Args[1:], os.LookupEnv) }

// ResolveSecret возвращает заданный секрет либо, если он пуст, генерирует
// эфемерный случайный секрет (32 байта crypto/rand в hex). Случайный секрет
// валиден лишь в пределах текущего процесса.
func ResolveSecret(s string) (string, error) {
	if s != "" {
		return s, nil
	}
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
