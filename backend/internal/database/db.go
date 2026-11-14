package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// LoadConfigFromEnv loads database configuration from environment variables.
func LoadConfigFromEnv() Config {
	return Config{
		Host:     envOrDefault("POSTGRES_HOST", "localhost"),
		Port:     envOrDefault("POSTGRES_PORT", "5432"),
		User:     envOrDefault("POSTGRES_USER", "postgres"),
		Password: envOrDefault("POSTGRES_PASSWORD", "postgres"),
		DBName:   envOrDefault("POSTGRES_DB", "gopackage_chronicle"),
		// Encrypt by default. Set POSTGRES_SSLMODE=disable for a local Postgres without TLS.
		SSLMode: envOrDefault("POSTGRES_SSLMODE", "require"),
	}
}

// Connect establishes a database connection using the provided configuration.
func Connect(cfg Config) (*sql.DB, error) {
	db, err := sql.Open("pgx", cfg.dsn())
	if err != nil {
		return nil, fmt.Errorf("database: open connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	slog.Info("database connection established",
		"host", cfg.Host,
		"port", cfg.Port,
		"database", cfg.DBName,
		"sslmode", cfg.SSLMode,
	)

	return db, nil
}

// dsn builds the pgx connection string. simple_protocol is required by
// Supabase's connection pooler.
func (c Config) dsn() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s default_query_exec_mode=simple_protocol",
		quoteDSNValue(c.Host), quoteDSNValue(c.Port), quoteDSNValue(c.User),
		quoteDSNValue(c.Password), quoteDSNValue(c.DBName), quoteDSNValue(c.SSLMode),
	)
}

// quoteDSNValue single-quotes a keyword/value DSN value so spaces, quotes and
// backslashes (common in generated passwords) survive parsing.
func quoteDSNValue(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, `'`, `\'`)
	return "'" + v + "'"
}

func envOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
