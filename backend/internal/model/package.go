package model

// Package represents a Go module package.
type Package struct {
	ID   int64  `json:"id"`
	Path string `json:"path"`
}
