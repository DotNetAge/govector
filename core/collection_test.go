package core

import (
	"os"
	"testing"
)

func TestCollection(t *testing.T) {
	// Test 1: Collection without storage
	t.Run("CollectionWithoutStorage", func(t *testing.T) {
		col, err := NewCollection("test-collection", 3, Cosine, nil, false)
		if err != nil {
			t.Fatalf("Failed to create collection: %v", err)
		}

		// Test Upsert
		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 0.0, 0.0}},
			{ID: "2", Vector: []float32{0.0, 1.0, 0.0}},
			{ID: "3", Vector: []float32{0.0, 0.0, 1.0}},
		}

		err = col.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		// Test Count
		if col.Count() != 3 {
			t.Errorf("Expected count 3, got %d", col.Count())
		}

		// Test Search
		query := []float32{1.0, 0.0, 0.0}
		results, err := col.Search(query, nil, 2)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}

		// Test Delete by IDs
		deleted, err := col.Delete([]string{"1", "2"}, nil)
		if err != nil {
			t.Fatalf("Failed to delete: %v", err)
		}

		if deleted != 2 {
			t.Errorf("Expected to delete 2 points, got %d", deleted)
		}

		if col.Count() != 1 {
			t.Errorf("Expected count 1 after deletion, got %d", col.Count())
		}
	})

	// Test 2: Collection with storage
	t.Run("CollectionWithStorage", func(t *testing.T) {
		// Create a temporary file for testing
		tempFile, err := os.CreateTemp("", "collection-test")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		tempPath := tempFile.Name()
		tempFile.Close()
		defer os.Remove(tempPath)

		// Create storage
		storage, err := NewStorage(tempPath)
		if err != nil {
			t.Fatalf("Failed to create storage: %v", err)
		}
		defer storage.Close()

		// Create collection with HNSW index
		col, err := NewCollection("test-collection", 3, Euclid, storage, true)
		if err != nil {
			t.Fatalf("Failed to create collection with storage: %v", err)
		}

		// Test Upsert
		points := []PointStruct{
			{ID: "1", Vector: []float32{1.0, 2.0, 3.0}},
			{ID: "2", Vector: []float32{4.0, 5.0, 6.0}},
		}

		err = col.Upsert(points)
		if err != nil {
			t.Fatalf("Failed to upsert points: %v", err)
		}

		if col.Count() != 2 {
			t.Errorf("Expected count 2, got %d", col.Count())
		}

		// Test Search
		query := []float32{1.0, 2.0, 3.0}
		results, err := col.Search(query, nil, 1)
		if err != nil {
			t.Fatalf("Failed to search: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(results))
		}

		if results[0].ID != "1" {
			t.Errorf("Expected first result to be ID 1, got %s", results[0].ID)
		}

		// Test Delete by filter
		filter := &Filter{
			Must: []Condition{
				{Key: "non-existent", Match: MatchValue{Value: "test"}},
			},
		}

		deleted, err := col.Delete(nil, filter)
		if err != nil {
			t.Fatalf("Failed to delete by filter: %v", err)
		}

		if deleted != 0 {
			t.Errorf("Expected to delete 0 points with non-existent filter, got %d", deleted)
		}

		// Test error cases
		t.Run("ErrorCases", func(t *testing.T) {
			// Test invalid vector length
			invalidPoints := []PointStruct{
				{ID: "3", Vector: []float32{1.0, 2.0}}, // Wrong length
			}

			err := col.Upsert(invalidPoints)
			if err == nil {
				t.Error("Expected error for invalid vector length")
			}

			// Test invalid query vector length
			invalidQuery := []float32{1.0, 2.0} // Wrong length
			_, err = col.Search(invalidQuery, nil, 1)
			if err == nil {
				t.Error("Expected error for invalid query vector length")
			}

			// Test delete with no points or filter
			_, err = col.Delete(nil, nil)
			if err == nil {
				t.Error("Expected error when deleting with no points or filter")
			}
		})
	})
}
