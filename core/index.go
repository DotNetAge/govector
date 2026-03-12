// Package core provides the core functionality for GoVector, a vector database library.
// It includes vector indexing, storage, and search capabilities compatible with Qdrant.
package core

// VectorIndex defines the standard interface for vector search engines.
// This allows transparent switching between Flat (brute-force) and HNSW (approximate) strategies.
type VectorIndex interface {
	// Upsert adds or updates points in the index.
	// If a point with the same ID already exists, it will be overwritten.
	Upsert(points []PointStruct) error

	// Search finds the nearest neighbors to the query vector, optionally applying a filter.
	// Returns up to topK results sorted by relevance.
	Search(query []float32, filter *Filter, topK int) ([]ScoredPoint, error)

	// Delete removes a point from the index by its ID.
	// Returns an error if the point does not exist.
	Delete(id string) error

	// Count returns the number of vectors currently stored in the index.
	Count() int

	// DeleteByFilter removes all points that match the given filter and returns their IDs.
	DeleteByFilter(filter *Filter) ([]string, error)
}
