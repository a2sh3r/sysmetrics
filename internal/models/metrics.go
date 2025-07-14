// Package models defines data structures for metrics used in API requests and responses.
package models

// Metrics represents a metric in API requests and responses.
type Metrics struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
