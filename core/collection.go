package core

import (
	"fmt"
	"sync"
)

// Collection represents a single logical group of vectors (like a table in SQL)
type Collection struct {
	Name      string
	VectorLen int
	Metric    Distance

	mu      sync.RWMutex
	index   VectorIndex // Transparent index engine (Flat or HNSW)
	storage *Storage    // Embedded Persistence
}

// NewCollection initializes a new vector collection, optionally loading from storage
// If useHNSW is true, it uses the optimized graph search, otherwise flat memory search.
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

// Upsert adds or updates points in the collection (Memory + Disk)
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

// Search delegates search to the underlying VectorIndex (Flat or HNSW)
func (c *Collection) Search(queryVector []float32, filter *Filter, topK int) ([]ScoredPoint, error) {
	if len(queryVector) != c.VectorLen {
		return nil, fmt.Errorf("query vector length mismatch: expected %d, got %d", c.VectorLen, len(queryVector))
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.index.Search(queryVector, filter, topK)
}

// Count returns the number of points in the collection
func (c *Collection) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.index.Count()
}
