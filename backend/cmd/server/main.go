package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/anphx/gopackage-chronicles/internal/api"
	"github.com/anphx/gopackage-chronicles/internal/database"
	"github.com/anphx/gopackage-chronicles/internal/migrations"
	"github.com/pressly/goose/v3"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// Connect to database
	cfg := database.LoadConfigFromEnv()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			slog.Error("failed to close database connection", "error", err)
		}
	}()

	// Run migrations
	goose.SetBaseFS(migrations.FS)

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}

	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	slog.Info("migrations completed successfully")

	// Start HTTP server
	port := envOrDefault("PORT", "8080")
	server := api.NewServer(db)

	addr := ":" + port
	slog.Info("starting HTTP server", "addr", addr)

	if err := http.ListenAndServe(addr, server); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

func envOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
