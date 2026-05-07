package db

import (
	"testing"
	"time"
)

func TestResolveSSLModeUsesExplicitConfig(t *testing.T) {
	t.Setenv("DB_SSL_MODE", "disable")

	if got := resolveSSLMode(ClientConfig{Host: "db.example.com", SslMode: "verify-full"}); got != "verify-full" {
		t.Fatalf("expected explicit ssl mode to win, got %q", got)
	}
}

func TestResolveSSLModeUsesEnvironmentFallback(t *testing.T) {
	t.Setenv("DB_SSL_MODE", "require")

	if got := resolveSSLMode(ClientConfig{Host: "db.example.com"}); got != "require" {
		t.Fatalf("expected env ssl mode fallback, got %q", got)
	}
}

func TestResolveSSLModeDefaultsDisableForLocalhost(t *testing.T) {
	t.Setenv("DB_SSL_MODE", "")

	if got := resolveSSLMode(ClientConfig{Host: "localhost"}); got != "disable" {
		t.Fatalf("expected localhost default ssl mode to be disable, got %q", got)
	}
}

func TestResolveSSLModeDefaultsRequireForRemoteHosts(t *testing.T) {
	t.Setenv("DB_SSL_MODE", "")

	if got := resolveSSLMode(ClientConfig{Host: "poltergeist.c7pqjm5qhybz.us-east-1.rds.amazonaws.com"}); got != "require" {
		t.Fatalf("expected remote host default ssl mode to be require, got %q", got)
	}
}

func TestIntFromEnvFallsBackForMissingValues(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "")

	if got := intFromEnv("DB_MAX_OPEN_CONNS", 7); got != 7 {
		t.Fatalf("expected fallback value, got %d", got)
	}
}

func TestIntFromEnvUsesConfiguredValue(t *testing.T) {
	t.Setenv("DB_MAX_OPEN_CONNS", "11")

	if got := intFromEnv("DB_MAX_OPEN_CONNS", 7); got != 11 {
		t.Fatalf("expected configured value, got %d", got)
	}
}

func TestDurationMinutesFromEnvUsesConfiguredMinutes(t *testing.T) {
	t.Setenv("DB_CONN_MAX_LIFETIME_MINUTES", "45")

	if got := durationMinutesFromEnv("DB_CONN_MAX_LIFETIME_MINUTES", 30); got != 45*time.Minute {
		t.Fatalf("expected 45 minute duration, got %s", got)
	}
}
