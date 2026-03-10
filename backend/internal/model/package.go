package model

import "time"

// Package represents a Go module package.
type Package struct {
	ID        int64     `json:"id"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}
