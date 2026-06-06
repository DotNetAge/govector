package core

import (
	"strings"
	"testing"
)

func TestHNSWIndex(t *testing.T) {
	// Test Cosine distance
	t.Run("CosineDistance", func(t *testing.T) {
		index := NewHNSWIndex(Cosine)

		// Test Upsert
		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 0.0, 0.0}, Payload: map[string]interface{}{"category": "A"}},
			{ID: "2", Vector: []float32{0.0, 1.0, 0.0}, Payload: map[string]interface{}{"category": "B"}},
			{ID: "3", Vector: []float32{0.0, 0.0, 1.0}, Payload: map[string]interface{}{"category": "A"}},
			{ID: "4", Vector: []float32{0.5, 0.5, 0.0}, Payload: map[string]interface{}{"category": "A"}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		// Test Count
		if index.Count() != 4 {
			t.Errorf("Expected count 4, got %d", index.Count())
		}

		// Test Search
		query := []float32{1.0, 0.0, 0.0}
		results, err := index.Search(query, nil, 2)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}

		// Test Search with filter
		filter := &Filter{
			Must: []Condition{
				{Key: "category", Match: MatchValue{Value: "A"}},
			},
		}

		filteredResults, err := index.Search(query, filter, 2)
		if err != nil {
			t.Fatalf("Failed to search with filter: %v", err)
		}

		if len(filteredResults) != 2 {
			t.Errorf("Expected 2 filtered results, got %d", len(filteredResults))
		}

		// Test Delete
		err = index.Delete("1")
		if err != nil {
			t.Fatalf("Failed to delete point: %v", err)
		}

		if index.Count() != 3 {
			t.Errorf("Expected count 3 after deletion, got %d", index.Count())
		}

		// Test Delete non-existent point
		err = index.Delete("non-existent")
		if err == nil {
			t.Error("Expected error when deleting non-existent point")
		}

		// Test DeleteByFilter
		deletedIDs, err := index.DeleteByFilter(filter)
		if err != nil {
			t.Fatalf("Failed to delete by filter: %v", err)
		}

		if len(deletedIDs) != 2 {
			t.Errorf("Expected to delete 2 points by filter, got %d", len(deletedIDs))
		}

		if index.Count() != 1 {
			t.Errorf("Expected count 1 after filter deletion, got %d", index.Count())
		}
	})

	// Test Euclidean distance
	t.Run("EuclideanDistance", func(t *testing.T) {
		index := NewHNSWIndex(Euclid)

		points := []PointStruct{
			{ID: "1", Vector: []float32{0.0, 0.0}},
			{ID: "2", Vector: []float32{1.0, 0.0}},
			{ID: "3", Vector: []float32{1.0, 1.0}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		query := []float32{0.0, 0.0}
		results, err := index.Search(query, nil, 3)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 3 {
			t.Errorf("Expected 3 results, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}
	})

	// Test Dot product
	t.Run("DotProduct", func(t *testing.T) {
		index := NewHNSWIndex(Dot)

		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 0.0}},
			{ID: "2", Vector: []float32{0.0, 1.0}},
			{ID: "3", Vector: []float32{1.0, 1.0}},
		}

		err := index.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		query := []float32{1.0, 0.0}
		results, err := index.Search(query, nil, 2)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		// Check that ID 1 is in results (order may vary due to HNSW)
		found := false
		for _, r := range results {
			if r.ID == "1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected ID 1 to be in results, got %v", results)
		}
	})
}

// TestHNSWIndexNilGraph verifies that all public methods return an error
// instead of panicking when the underlying HNSW graph is nil.
// This covers the P0 fix: defensive nil checks in Upsert/Search/Delete/DeleteByFilter.
func TestHNSWIndexNilGraph(t *testing.T) {
	t.Run("NilGraph_UpsertReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		err := idx.Upsert([]PointStruct{{ID: "1", Vector: []float32{1.0}}})
		if err == nil {
			t.Fatal("expected error when upserting with nil graph")
		}
		if !strings.Contains(err.Error(), "nil") {
			t.Errorf("error should mention nil graph, got: %v", err)
		}
	})

	t.Run("NilGraph_SearchReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		_, err := idx.Search([]float32{1.0}, nil, 1)
		if err == nil {
			t.Fatal("expected error when searching with nil graph")
		}
	})

	t.Run("NilGraph_DeleteReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		err := idx.Delete("1")
		if err == nil {
			t.Fatal("expected error when deleting with nil graph")
		}
	})

	t.Run("NilGraph_DeleteByFilterReturnsError", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		filter := &Filter{
			Must: []Condition{{Key: "x", Match: MatchValue{Value: "y"}}},
		}
		_, err := idx.DeleteByFilter(filter)
		if err == nil {
			t.Fatal("expected error when DeleteByFilter with nil graph")
		}
	})

	t.Run("NilGraph_CountIsZero", func(t *testing.T) {
		idx := &HNSWIndex{points: make(map[string]*PointStruct)}
		if idx.Count() != 0 {
			t.Errorf("expected count 0 for nil graph, got %d", idx.Count())
		}
	})
}
