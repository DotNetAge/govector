package core

import (
	"fmt"
	"sort"
	"sync"
)

// FlatIndex implements VectorIndex using a brute-force search algorithm.
// It stores all vectors in memory and performs exhaustive distance calculations
// for each query. This provides exact results but has O(n) search complexity,
// making it suitable for small to medium datasets.
type FlatIndex struct {
	points map[string]*PointStruct
	metric Distance
	mu     sync.RWMutex
}

// NewFlatIndex creates a new flat memory index with the specified distance metric.
func NewFlatIndex(metric Distance) *FlatIndex {
	return &FlatIndex{
		points: make(map[string]*PointStruct),
		metric: metric,
	}
}

// Upsert adds or updates points in the index.
// If a point with the same ID already exists, it will be overwritten.
func (f *FlatIndex) Upsert(points []PointStruct) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range points {
		p := &points[i]
		f.points[p.ID] = p
	}
	return nil
}

// Search performs a brute-force search for the nearest neighbors.
// It calculates the distance between the query vector and all stored vectors,
// filters by payload if specified, and returns the topK results.
// Results are sorted by relevance: descending for Cosine/Dot, ascending for Euclidean.
func (f *FlatIndex) Search(query []float32, filter *Filter, topK int) ([]ScoredPoint, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var results []ScoredPoint
	for id, point := range f.points {
		if !MatchFilter(point.Payload, filter) {
			continue
		}

		dist := CalculateDistance(f.metric, query, point.Vector)
		results = append(results, ScoredPoint{
			ID:      id,
			Score:   dist,
			Payload: point.Payload,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		if f.metric == Euclid {
			return results[i].Score < results[j].Score
		}
		return results[i].Score > results[j].Score
	})

	if topK > len(results) {
		topK = len(results)
	}

	return results[:topK], nil
}

// Delete removes a point from the index by its ID.
// Returns an error if the point does not exist.
func (f *FlatIndex) Delete(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.points[id]; exists {
		delete(f.points, id)
		return nil
	}
	return fmt.Errorf("point %s not found", id)
}

// Count returns the number of vectors currently stored in the index.
func (f *FlatIndex) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.points)
}

// GetIDsByFilter returns all point IDs that match the given filter.
func (f *FlatIndex) GetIDsByFilter(filter *Filter) []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	var ids []string
	for id, point := range f.points {
		if MatchFilter(point.Payload, filter) {
			ids = append(ids, id)
		}
	}
	return ids
}

// DeleteByFilter removes all points that match the given filter and returns their IDs.
func (f *FlatIndex) DeleteByFilter(filter *Filter) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	var deleted []string
	for id, point := range f.points {
		if MatchFilter(point.Payload, filter) {
			delete(f.points, id)
			deleted = append(deleted, id)
		}
	}
	return deleted, nil
}
