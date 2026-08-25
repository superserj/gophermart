package config

import "testing"

func lookup(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) { v, ok := m[k]; return v, ok }
}

func TestDefaults(t *testing.T) {
	cfg, err := Parse(nil, lookup(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RunAddress != "localhost:8080" {
		t.Fatalf("default RunAddress = %q", cfg.RunAddress)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("default LogLevel = %q", cfg.LogLevel)
	}
	if cfg.AuthSecret != "" {
		t.Fatalf("default AuthSecret = %q, want empty", cfg.AuthSecret)
	}
}

func TestResolveSecretKeepsExplicit(t *testing.T) {
	got, err := ResolveSecret("my-secret")
	if err != nil {
		t.Fatal(err)
	}
	if got != "my-secret" {
		t.Fatalf("ResolveSecret(%q) = %q, want unchanged", "my-secret", got)
	}
}

func TestResolveSecretGeneratesRandom(t *testing.T) {
	first, err := ResolveSecret("")
	if err != nil {
		t.Fatal(err)
	}
	if first == "" {
		t.Fatal("ResolveSecret(\"\") returned empty secret")
	}
	second, err := ResolveSecret("")
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatalf("ResolveSecret(\"\") produced identical secrets %q", first)
	}
}

func TestFlags(t *testing.T) {
	cfg, err := Parse([]string{"-a", "127.0.0.1:9000", "-d", "postgres://flag", "-r", "http://accrual:8888", "-s", "sekret", "-l", "debug"}, lookup(nil))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RunAddress != "127.0.0.1:9000" || cfg.DatabaseURI != "postgres://flag" ||
		cfg.AccrualSystemAddress != "http://accrual:8888" || cfg.AuthSecret != "sekret" || cfg.LogLevel != "debug" {
		t.Fatalf("flags not applied: %+v", cfg)
	}
}

func TestFlagsOverrideEnv(t *testing.T) {
	env := map[string]string{
		"RUN_ADDRESS": "0.0.0.0:7000", "DATABASE_URI": "postgres://env",
		"ACCRUAL_SYSTEM_ADDRESS": "http://env:1", "AUTH_SECRET": "env-secret", "LOG_LEVEL": "warn",
	}
	// -a и -d заданы флагами → перекрывают ENV; остальные поля берутся из ENV.
	cfg, err := Parse([]string{"-a", "127.0.0.1:9000", "-d", "postgres://flag"}, lookup(env))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RunAddress != "127.0.0.1:9000" || cfg.DatabaseURI != "postgres://flag" {
		t.Fatalf("flags must override env: %+v", cfg)
	}
	if cfg.AccrualSystemAddress != "http://env:1" || cfg.AuthSecret != "env-secret" || cfg.LogLevel != "warn" {
		t.Fatalf("env must fill fields without flags: %+v", cfg)
	}
}
