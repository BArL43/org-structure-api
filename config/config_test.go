package config

import (
	"strings"
	"testing"
)

func TestLoadAndDSN(t *testing.T) {
	t.Setenv("SERVER_PORT", "8080")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "p@ss word")
	t.Setenv("DB_NAME", "org_structure")
	t.Setenv("DB_SSLMODE", "require")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	dsn := cfg.GetDSN()
	if !strings.Contains(dsn, "p%40ss%20word") || !strings.Contains(dsn, "sslmode=require") {
		t.Fatalf("dsn is not safely encoded: %s", dsn)
	}
}

func TestLoadRejectsInvalidPort(t *testing.T) {
	t.Setenv("SERVER_PORT", "70000")
	if _, err := Load(); err == nil {
		t.Fatal("expected invalid port error")
	}
}
