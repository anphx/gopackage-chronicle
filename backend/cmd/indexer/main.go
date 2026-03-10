package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/anphx/gopackage-chronicles/internal/database"
	"github.com/anphx/gopackage-chronicles/internal/indexer"
	repo "github.com/anphx/gopackage-chronicles/internal/repository"
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

	// Initialize repositories
	packageRepo := repo.NewPackageRepository(db)
	releaseRepo := repo.NewReleaseRepository(db)
	cursorRepo := repo.NewSyncCursorRepository(db)

	// Initialize indexer client and syncer
	client := indexer.NewIndexClient()
	syncer := indexer.NewSyncer(client, packageRepo, releaseRepo, cursorRepo)

	// Run sync
	ctx := context.Background()
	if err := syncer.Run(ctx); err != nil {
		log.Fatalf("sync failed: %v", err)
	}

	slog.Info("indexer completed successfully")
}
