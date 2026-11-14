package database

import (
	"strings"
	"testing"
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
			wantDSN: "sslmode=require",
		},
		{
			name:    "disable for local postgres",
			env:     "disable",
			wantDSN: "sslmode=disable",
		},
		{
			name:    "verify-full passes through",
			env:     "verify-full",
			wantDSN: "sslmode=verify-full",
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
