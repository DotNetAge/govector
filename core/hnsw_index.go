package core

import (
	"fmt"
	"math"
	"sync"

	"github.com/coder/hnsw"
)

// HNSWParams contains configurable parameters for HNSW index
// See https://github.com/coder/hnsw for more details on these parameters
type HNSWParams struct {
	// M is the maximum number of connections per node
	// Default: 16
	M int

	// EfConstruction is the size of the dynamic candidate list during construction
	// Default: 200
	EfConstruction int

	// EfSearch is the size of the dynamic candidate list during search
	// Default: 64
	EfSearch int

	// K is the number of nearest neighbors to return
	// Default: 10
	K int
}

// DefaultHNSWParams returns default HNSW parameters
func DefaultHNSWParams() HNSWParams {
	return HNSWParams{
		M:              16,
		EfConstruction: 200,
		EfSearch:       64,
		K:              10,
	}
}

// HNSWIndex wraps the coder/hnsw graph to provide an approximate nearest neighbor search.
// It uses Hierarchical Navigable Small World graphs for efficient similarity search
// with sub-linear complexity, making it suitable for large datasets.
type HNSWIndex struct {
	graph  *hnsw.Graph[string]
	points map[string]*PointStruct // Fast Metadata / Payload lookup
	metric Distance
	params HNSWParams
	mu     sync.RWMutex
}

// NewHNSWIndex creates a new HNSW index engine with the specified distance metric.
// It configures the underlying HNSW graph with appropriate distance functions
// for Cosine, Euclidean, or Dot product metrics.
func NewHNSWIndex(metric Distance) *HNSWIndex {
	return NewHNSWIndexWithParams(metric, DefaultHNSWParams())
}

// NewHNSWIndexWithParams creates a new HNSW index engine with custom parameters.
// It allows fine-tuning of HNSW parameters for specific use cases.
func NewHNSWIndexWithParams(metric Distance, params HNSWParams) *HNSWIndex {
	g := hnsw.NewGraph[string]()

	// Configure Distance Function based on our definitions
	switch metric {
	case Euclid:
		g.Distance = func(a, b []float32) float32 {
			var sum float32
			for i := range a {
				diff := a[i] - b[i]
				sum += diff * diff
			}
			return float32(math.Sqrt(float64(sum)))
		}
	case Dot:
		g.Distance = func(a, b []float32) float32 {
			var sum float32
			for i := range a {
				sum += a[i] * b[i]
			}
			// hnsw minimizes distance, so we invert dot product
			return -sum
		}
	case Cosine:
		fallthrough
	default:
		// Default is Cosine Distance (1 - CosineSimilarity)
		// hnsw seeks minimum distance, so 1 - sim is correct
		g.Distance = hnsw.CosineDistance
	}

	// Set HNSW parameters
	if params.M > 0 {
		g.M = params.M
	}
	if params.EfSearch > 0 {
		g.EfSearch = params.EfSearch
	}

	return &HNSWIndex{
		graph:  g,
		points: make(map[string]*PointStruct),
		metric: metric,
		params: params,
	}
}

// Upsert adds or updates points in the HNSW graph.
// Points are added to both the HNSW graph for search and a local map for payload lookup.
func (h *HNSWIndex) Upsert(points []PointStruct) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	var nodes []hnsw.Node[string]
	for i := range points {
		p := &points[i]
		h.points[p.ID] = p
		nodes = append(nodes, hnsw.MakeNode(p.ID, p.Vector))
	}

	// Batch add to graph
	h.graph.Add(nodes...)
	return nil
}

// Search performs an approximate nearest neighbor search using the HNSW algorithm.
// It uses a post-filtering strategy: over-fetches results to account for filtered points,
// then applies the payload filter and returns the topK matches.
func (h *HNSWIndex) Search(query []float32, filter *Filter, topK int) ([]ScoredPoint, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// If filtering is needed, we must over-fetch from HNSW (Post-Filtering Strategy for MVP)
	// In a heavily filtered scenario, this would ideally be pushed into the graph traversal.
	fetchK := topK
	if filter != nil {
		fetchK = topK * 10 // Heuristic: fetch more to account for dropped points
		// Cap it to the total graph size just in case
		if fetchK > len(h.points) {
			fetchK = len(h.points)
		}
	}

	// 1. Search the HNSW Graph
	neighbors := h.graph.Search(query, fetchK)

	// 2. Filter & Format Results
	var results []ScoredPoint
	for _, neighbor := range neighbors {
		point, exists := h.points[neighbor.Key]
		if !exists {
			continue // Should not happen if graph and map are in sync
		}

		// Apply Payload Filter
		if !MatchFilter(point.Payload, filter) {
			continue
		}

		// Calculate precise Score based on our metric definition
		// (hnsw returns nodes, we recalculate original score format)
		score := CalculateDistance(h.metric, query, point.Vector)

		results = append(results, ScoredPoint{
			ID:      point.ID,
			Score:   score,
			Payload: point.Payload,
		})

		if len(results) == topK {
			break
		}
	}

	return results, nil
}

// Delete removes a point from both the HNSW graph and the local points map.
// Returns an error if the point does not exist in the graph.
func (h *HNSWIndex) Delete(id string) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.graph.Delete(id) {
		delete(h.points, id)
		return nil
	}
	return fmt.Errorf("point %s not found", id)
}

// Count returns the number of vectors currently stored in the index.
func (h *HNSWIndex) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.points)
}

// GetIDsByFilter returns all point IDs that match the given filter.
func (h *HNSWIndex) GetIDsByFilter(filter *Filter) []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var ids []string
	for id, point := range h.points {
		if MatchFilter(point.Payload, filter) {
			ids = append(ids, id)
		}
	}
	return ids
}

// DeleteByFilter removes all points that match the given filter from both
// the HNSW graph and the local points map. Returns the IDs of deleted points.
func (h *HNSWIndex) DeleteByFilter(filter *Filter) ([]string, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	var deleted []string
	for id, point := range h.points {
		if MatchFilter(point.Payload, filter) {
			if h.graph.Delete(id) {
				delete(h.points, id)
				deleted = append(deleted, id)
			}
		}
	}
	return deleted, nil
}
