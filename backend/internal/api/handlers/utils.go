package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON", "error", err)
	}
}

func setHourlyCache(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "public, max-age=3600, stale-while-revalidate=3600")
}
