package database

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Config holds database connection configuration.
type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// LoadConfigFromEnv loads database configuration from environment variables.
func LoadConfigFromEnv() Config {
	return Config{
		Host:     envOrDefault("POSTGRES_HOST", "localhost"),
		Port:     envOrDefault("POSTGRES_PORT", "5432"),
		User:     envOrDefault("POSTGRES_USER", "postgres"),
		Password: envOrDefault("POSTGRES_PASSWORD", "postgres"),
		DBName:   envOrDefault("POSTGRES_DB", "gopackage_chronicle"),
	}
}

// Connect establishes a database connection using the provided configuration.
func Connect(cfg Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable default_query_exec_mode=simple_protocol",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := sql.Open("pgx", connStr)
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
	)

	return db, nil
}

func envOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
