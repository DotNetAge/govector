package core

import (
	"os"
	"testing"
)

func TestSQ8Quantizer(t *testing.T) {
	// Create a test vector
	vector := []float32{1.0, -0.5, 0.75, -1.0, 0.0}

	// Create SQ8 quantizer
	quantizer := NewSQ8Quantizer()

	// Quantize the vector
	compressed := quantizer.Quantize(vector)

	// Check compressed size
	expectedSize := 8 + len(vector) // 8 bytes for min/max, 1 byte per element
	if len(compressed) != expectedSize {
		t.Errorf("Expected compressed size %d, got %d", expectedSize, len(compressed))
	}

	// Dequantize the vector
	dequantized := quantizer.Dequantize(compressed)

	// Check dequantized vector length
	if len(dequantized) != len(vector) {
		t.Errorf("Expected dequantized length %d, got %d", len(vector), len(dequantized))
	}

	// Check if dequantized values are close to original
	for i, val := range dequantized {
		// Allow small error due to quantization
		if val < vector[i]-0.1 || val > vector[i]+0.1 {
			t.Errorf("Dequantized value at index %d: expected %f, got %f", i, vector[i], val)
		}
	}

	// Test GetCompressedSize
	expectedCompressedSize := quantizer.GetCompressedSize(len(vector))
	if expectedCompressedSize != expectedSize {
		t.Errorf("Expected compressed size %d, got %d", expectedSize, expectedCompressedSize)
	}
}

func TestSQ8QuantizerEdgeCases(t *testing.T) {
	quantizer := NewSQ8Quantizer()

	// Test case 1: Empty vector
	emptyVector := []float32{}
	compressed := quantizer.Quantize(emptyVector)
	if len(compressed) != 0 {
		t.Errorf("Expected empty compressed data for empty vector, got length %d", len(compressed))
	}

	dequantized := quantizer.Dequantize(compressed)
	if len(dequantized) != 0 {
		t.Errorf("Expected empty dequantized vector for empty data, got length %d", len(dequantized))
	}

	// Test case 2: Single element vector
	singleVector := []float32{1.0}
	compressed = quantizer.Quantize(singleVector)
	dequantized = quantizer.Dequantize(compressed)
	if len(dequantized) != 1 || dequantized[0] != 1.0 {
		t.Errorf("Expected [1.0], got %v", dequantized)
	}

	// Test case 3: Vector with all same values
	sameVector := []float32{5.0, 5.0, 5.0}
	compressed = quantizer.Quantize(sameVector)
	dequantized = quantizer.Dequantize(compressed)
	for _, v := range dequantized {
		if v != 5.0 {
			t.Errorf("Expected 5.0, got %f", v)
		}
	}

	// Test case 4: Dequantize with invalid data length
	invalidData := []byte{1, 2, 3}
	dequantized = quantizer.Dequantize(invalidData)
	if len(dequantized) != 0 {
		t.Errorf("Expected empty dequantized vector for invalid data length, got length %d", len(dequantized))
	}

	// Test case 5: Large range values
	largeRangeVector := []float32{-1000.0, 1000.0}
	compressed = quantizer.Quantize(largeRangeVector)
	dequantized = quantizer.Dequantize(compressed)
	if len(dequantized) != 2 {
		t.Errorf("Expected length 2, got %d", len(dequantized))
	}
	// Check if values are reasonably close
	if dequantized[0] < -1005.0 || dequantized[0] > -995.0 {
		t.Errorf("Dequantized[0] far from -1000.0: %f", dequantized[0])
	}
}

func TestQuantizationWithStorage(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := createTempFile()
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer removeTempFile(tempPath)

	// Create storage with quantization
	storage, err := NewStorageWithQuantization(tempPath, true, nil, false)
	if err != nil {
		t.Fatalf("Failed to create storage with quantization: %v", err)
	}
	defer storage.Close()

	// Create test points
	points := []PointStruct{
		{ID: "1", Vector: []float32{1.0, 0.0, 0.0}},
		{ID: "2", Vector: []float32{0.0, 1.0, 0.0}},
		{ID: "3", Vector: []float32{0.0, 0.0, 1.0}},
	}

	// Ensure collection exists
	err = storage.EnsureCollection("test-collection")
	if err != nil {
		t.Fatalf("Failed to ensure collection: %v", err)
	}

	// Upsert points
	err = storage.UpsertPoints("test-collection", points)
	if err != nil {
		t.Fatalf("Failed to upsert points: %v", err)
	}

	// Load collection
	loadedPoints, err := storage.LoadCollection("test-collection")
	if err != nil {
		t.Fatalf("Failed to load collection: %v", err)
	}

	// Check if points were loaded correctly
	if len(loadedPoints) != len(points) {
		t.Errorf("Expected %d points, got %d", len(points), len(loadedPoints))
	}

	// Check if vectors are close to original
	for id, loaded := range loadedPoints {
		// Find the original point with the same ID
		var original PointStruct
		found := false
		for _, p := range points {
			if p.ID == loaded.ID {
				original = p
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Original point not found for ID %s", loaded.ID)
			continue
		}
		
		if loaded.ID != original.ID {
			t.Errorf("Expected ID %s, got %s", original.ID, loaded.ID)
		}
		for j, val := range loaded.Vector {
			if val < original.Vector[j]-0.1 || val > original.Vector[j]+0.1 {
				t.Errorf("Vector for ID %s, element %d: expected %f, got %f", id, j, original.Vector[j], val)
			}
		}
	}
}

// Helper functions for temporary files
func createTempFile() (tempFile *os.File, err error) {
	return os.CreateTemp("", "quantization-test")
}

func removeTempFile(path string) {
	os.Remove(path)
}
