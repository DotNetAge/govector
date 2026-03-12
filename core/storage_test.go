package core

import (
	"os"
	"testing"
)

func TestStorage(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "storage-test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// Test NewStorage
	storage, err := NewStorage(tempPath)
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	// Test EnsureCollection
	err = storage.EnsureCollection("test-collection")
	if err != nil {
		t.Fatalf("Failed to ensure collection: %v", err)
	}

	// Test UpsertPoints
	points := []PointStruct{
		{ID: "1", Vector: []float32{1.0, 2.0, 3.0}, Payload: map[string]interface{}{"name": "test1"}},
		{ID: "2", Vector: []float32{4.0, 5.0, 6.0}, Payload: map[string]interface{}{"name": "test2"}},
	}
	err = storage.UpsertPoints("test-collection", points)
	if err != nil {
		t.Fatalf("Failed to upsert points: %v", err)
	}

	// Test LoadCollection
	loadedPoints, err := storage.LoadCollection("test-collection")
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	if len(loadedPoints) != 2 {
		t.Errorf("Expected 2 points, got %d", len(loadedPoints))
	}

	if loadedPoints["1"] == nil {
		t.Error("Expected point 1 to exist")
	}

	if loadedPoints["2"] == nil {
		t.Error("Expected point 2 to exist")
	}

	// Test DeletePoints
	err = storage.DeletePoints("test-collection", []string{"1"})
	if err != nil {
		t.Fatalf("Failed to delete points: %v", err)
	}

	// Test LoadCollection after deletion
	loadedPoints, err = storage.LoadCollection("test-collection")
	if err != nil {
		t.Fatalf("Failed to load collection after deletion: %v", err)
	}

	if len(loadedPoints) != 1 {
		t.Errorf("Expected 1 point after deletion, got %d", len(loadedPoints))
	}

	if loadedPoints["1"] != nil {
		t.Error("Expected point 1 to be deleted")
	}

	if loadedPoints["2"] == nil {
		t.Error("Expected point 2 to still exist")
	}

	// Test DeletePoints on non-existent collection
	err = storage.DeletePoints("non-existent", []string{"1"})
	if err != nil {
		t.Errorf("Expected no error when deleting from non-existent collection, got %v", err)
	}

	// Test UpsertPoints on non-existent collection
	err = storage.UpsertPoints("non-existent", points)
	if err == nil {
		t.Error("Expected error when upserting to non-existent collection")
	}
}
