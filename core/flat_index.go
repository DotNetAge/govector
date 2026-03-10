package core

import (
	"fmt"
	"sort"
	"sync"
)

// FlatIndex implements VectorIndex using a brute-force search
type FlatIndex struct {
	points map[string]*PointStruct
	metric Distance
	mu     sync.RWMutex
}

// NewFlatIndex creates a new flat memory index
func NewFlatIndex(metric Distance) *FlatIndex {
	return &FlatIndex{
		points: make(map[string]*PointStruct),
		metric: metric,
	}
}

func (f *FlatIndex) Upsert(points []PointStruct) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i := range points {
		p := &points[i]
		f.points[p.ID] = p
	}
	return nil
}

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

func (f *FlatIndex) Delete(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.points[id]; exists {
		delete(f.points, id)
		return nil
	}
	return fmt.Errorf("point %s not found", id)
}

func (f *FlatIndex) Count() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.points)
}

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
