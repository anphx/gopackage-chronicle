package main

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Configure a database connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create an HTTP server with timeouts
	port := envOrDefault("PORT", "8080")
	handler := api.NewServer(db)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel that receives OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Start the HTTPserver in a goroutine so main isn't blocked.
	go func() {
		slog.Info("starting HTTP server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	slog.Info("server started")

	// Block here until signal received
	<-quit

	slog.Info("shutting down server...")

	// Give outstanding requests 10 seconds to complete.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server exited gracefully")
}

func envOrDefault(key, fallback string) string {
	val := os.Getenv(key)
	if val == "" {
		return fallback
	}
	return val
}
