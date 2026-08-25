// Package config разбирает флаги и ENV-переменные сервиса.
package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
)

// Config — параметры запуска. Приоритет: флаг командной строки > ENV > дефолт.
type Config struct {
	RunAddress           string
	DatabaseURI          string
	AccrualSystemAddress string
	AuthSecret           string
	LogLevel             string
}

// Parse разбирает args. Дефолтом каждого флага служит значение из ENV (если оно
// задано), иначе встроенный дефолт; явно переданный флаг перекрывает ENV. Так
// флаг — явное намерение оператора «здесь и сейчас» — имеет высший приоритет.
func Parse(args []string, lookup func(string) (string, bool)) (*Config, error) {
	cfg := &Config{}
	fs := flag.NewFlagSet("gophermart", flag.ContinueOnError)
	fs.StringVar(&cfg.RunAddress, "a", envOr(lookup, "RUN_ADDRESS", "localhost:8080"), "address to run HTTP server")
	fs.StringVar(&cfg.DatabaseURI, "d", envOr(lookup, "DATABASE_URI", ""), "PostgreSQL DSN")
	fs.StringVar(&cfg.AccrualSystemAddress, "r", envOr(lookup, "ACCRUAL_SYSTEM_ADDRESS", ""), "accrual system address")
	fs.StringVar(&cfg.AuthSecret, "s", envOr(lookup, "AUTH_SECRET", ""), "secret key for auth signature (random ephemeral if empty)")
	fs.StringVar(&cfg.LogLevel, "l", envOr(lookup, "LOG_LEVEL", "info"), "log level")
	if err := fs.Parse(args); err != nil {
		return nil, err
	}
	return cfg, nil
}

// envOr возвращает значение ENV-переменной key, либо def, если она не задана.
func envOr(lookup func(string) (string, bool), key, def string) string {
	if v, ok := lookup(key); ok {
		return v
	}
	return def
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
