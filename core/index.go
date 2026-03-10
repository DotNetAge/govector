package core

// VectorIndex defines the standard interface for vector search engines
// This allows us to transparently switch between Flat and HNSW strategies
type VectorIndex interface {
	// Upsert adds or updates points in the index
	Upsert(points []PointStruct) error
	
	// Search finds the nearest neighbors, optionally applying a filter
	Search(query []float32, filter *Filter, topK int) ([]ScoredPoint, error)
	
	// Delete removes a point from the index by its ID
	Delete(id string) error

	// Count returns the number of vectors in the index
	Count() int
}
