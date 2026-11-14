package database

import (
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestLoadConfigFromEnvSSLMode(t *testing.T) {
	tests := []struct {
		name    string
		env     string
		wantDSN string
	}{
		{
			name:    "defaults to require when unset",
			env:     "",
			wantDSN: "sslmode='require'",
		},
		{
			name:    "disable for local postgres",
			env:     "disable",
			wantDSN: "sslmode='disable'",
		},
		{
			name:    "verify-full passes through",
			env:     "verify-full",
			wantDSN: "sslmode='verify-full'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("POSTGRES_SSLMODE", tt.env)

			dsn := LoadConfigFromEnv().dsn()

			if !strings.Contains(dsn, tt.wantDSN) {
				t.Errorf("dsn = %q, want it to contain %q", dsn, tt.wantDSN)
			}
			if !strings.Contains(dsn, "default_query_exec_mode=simple_protocol") {
				t.Errorf("dsn = %q, lost simple_protocol (needed for Supabase pooler)", dsn)
			}
		})
	}
}

func TestDSNPasswordSurvivesParsing(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{name: "plain", password: "s3cret"},
		{name: "hash", password: "abc#def"},
		{name: "space", password: "two words"},
		{name: "single quote", password: "it's"},
		{name: "backslash", password: `back\slash`},
		{name: "equals", password: "a=b"},
		{name: "empty", password: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Host: "db.example.com", Port: "6543", User: "postgres.ref",
				Password: tt.password, DBName: "postgres", SSLMode: "require",
			}

			parsed, err := pgx.ParseConfig(cfg.dsn())
			if err != nil {
				t.Fatalf("ParseConfig: %v", err)
			}
			if parsed.Password != tt.password {
				t.Errorf("password = %q, want %q", parsed.Password, tt.password)
			}
			if parsed.User != cfg.User || parsed.Database != cfg.DBName || parsed.Host != cfg.Host {
				t.Errorf("got user=%q db=%q host=%q, want %q %q %q",
					parsed.User, parsed.Database, parsed.Host, cfg.User, cfg.DBName, cfg.Host)
			}
			if parsed.DefaultQueryExecMode != pgx.QueryExecModeSimpleProtocol {
				t.Errorf("exec mode = %v, want simple protocol", parsed.DefaultQueryExecMode)
			}
		})
	}
}
