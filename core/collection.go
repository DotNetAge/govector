package core

import (
	"fmt"
	"sync"
)

// Collection represents a single logical group of vectors, similar to a table in SQL databases.
// It provides thread-safe operations for inserting, searching, and deleting vectors.
// Each collection has a fixed vector dimension and uses a specific distance metric.
type Collection struct {
	Name      string      // Unique name of the collection
	VectorLen int         // Dimension of vectors in this collection
	Metric    Distance    // Distance metric used for similarity search

	mu      sync.RWMutex
	index   VectorIndex // Transparent index engine (Flat or HNSW)
	storage *Storage    // Embedded Persistence
}

// NewCollection initializes a new vector collection, optionally loading from storage.
// If useHNSW is true, it uses the optimized HNSW graph search; otherwise uses flat memory search.
// When storage is provided, existing points are automatically loaded into memory.
// Returns an error if the collection cannot be created or if loaded points have invalid dimensions.
func NewCollection(name string, vectorLen int, metric Distance, store *Storage, useHNSW bool) (*Collection, error) {
	var index VectorIndex
	if useHNSW {
		index = NewHNSWIndex(metric)
	} else {
		index = NewFlatIndex(metric)
	}

	col := &Collection{
		Name:      name,
		VectorLen: vectorLen,
		Metric:    metric,
		index:     index,
		storage:   store,
	}

	if store != nil {
		if err := store.EnsureCollection(name); err != nil {
			return nil, fmt.Errorf("failed to ensure storage collection %s: %w", name, err)
		}

		// Load existing points into memory index
		loadedPoints, err := store.LoadCollection(name)
		if err != nil {
			return nil, fmt.Errorf("failed to load points for %s: %w", name, err)
		}

		// Validate and prep array for Upsert
		var batch []PointStruct
		for _, p := range loadedPoints {
			if len(p.Vector) != vectorLen {
				return nil, fmt.Errorf("loaded point %s has invalid dimension %d (expected %d)", p.ID, len(p.Vector), vectorLen)
			}
			batch = append(batch, *p)
		}

		if len(batch) > 0 {
			if err := col.index.Upsert(batch); err != nil {
				return nil, fmt.Errorf("failed to load initial points into index: %w", err)
			}
		}
	}

	return col, nil
}

// Upsert adds or updates points in the collection.
// Points are first persisted to disk (if storage is configured), then updated in the memory index.
// Returns an error if any point has an invalid vector length or if persistence fails.
func (c *Collection) Upsert(points []PointStruct) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Validate
	for i := range points {
		p := &points[i]
		if len(p.Vector) != c.VectorLen {
			return fmt.Errorf("point %s has invalid vector length %d, expected %d", p.ID, len(p.Vector), c.VectorLen)
		}
	}

	// 2. Persist to Disk First
	if c.storage != nil {
		if err := c.storage.UpsertPoints(c.Name, points); err != nil {
			return fmt.Errorf("failed to persist points: %w", err)
		}
	}

	// 3. Update Memory Index Engine
	return c.index.Upsert(points)
}

// Search performs a similarity search using the underlying VectorIndex.
// It finds the topK nearest neighbors to the query vector, optionally filtered by payload.
// Returns an error if the query vector dimension doesn't match the collection.
func (c *Collection) Search(queryVector []float32, filter *Filter, topK int) ([]ScoredPoint, error) {
	if len(queryVector) != c.VectorLen {
		return nil, fmt.Errorf("query vector length mismatch: expected %d, got %d", c.VectorLen, len(queryVector))
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.index.Search(queryVector, filter, topK)
}

// Count returns the number of points currently stored in the collection.
func (c *Collection) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.index.Count()
}

// Delete removes points either by explicit IDs or by a filter match.
// If points slice is provided, those specific points are deleted.
// If filter is provided, all points matching the filter are deleted.
// Returns the number of points deleted and any error encountered.
// Returns an error if neither points nor filter is provided.
func (c *Collection) Delete(points []string, filter *Filter) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var deletedIDs []string

	if len(points) > 0 {
		for _, id := range points {
			if err := c.index.Delete(id); err == nil {
				deletedIDs = append(deletedIDs, id)
			}
		}
	} else if filter != nil {
		var err error
		deletedIDs, err = c.index.DeleteByFilter(filter)
		if err != nil {
			return 0, fmt.Errorf("failed to delete by filter: %w", err)
		}
	} else {
		return 0, fmt.Errorf("must provide points or filter")
	}

	if len(deletedIDs) > 0 && c.storage != nil {
		if err := c.storage.DeletePoints(c.Name, deletedIDs); err != nil {
			return len(deletedIDs), fmt.Errorf("deleted from memory but failed to persist: %w", err)
		}
	}

	return len(deletedIDs), nil
}
