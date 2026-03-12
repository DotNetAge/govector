package core

import (
	"math"
	"testing"
)

func TestCalculateDistance(t *testing.T) {
	// Test Cosine similarity
	t.Run("CosineSimilarity", func(t *testing.T) {
		// Identical vectors should have cosine similarity 1.0
		a := []float32{1.0, 0.0, 0.0}
		b := []float32{1.0, 0.0, 0.0}
		similarity := CalculateDistance(Cosine, a, b)
		if math.Abs(float64(similarity-1.0)) > 1e-6 {
			t.Errorf("Expected cosine similarity 1.0, got %f", similarity)
		}

		// Orthogonal vectors should have cosine similarity 0.0
		a = []float32{1.0, 0.0, 0.0}
		b = []float32{0.0, 1.0, 0.0}
		similarity = CalculateDistance(Cosine, a, b)
		if math.Abs(float64(similarity)) > 1e-6 {
			t.Errorf("Expected cosine similarity 0.0, got %f", similarity)
		}

		// Opposite vectors should have cosine similarity -1.0
		a = []float32{1.0, 0.0, 0.0}
		b = []float32{-1.0, 0.0, 0.0}
		similarity = CalculateDistance(Cosine, a, b)
		if math.Abs(float64(similarity+1.0)) > 1e-6 {
			t.Errorf("Expected cosine similarity -1.0, got %f", similarity)
		}

		// Test with zero vectors
		a = []float32{0.0, 0.0, 0.0}
		b = []float32{1.0, 0.0, 0.0}
		similarity = CalculateDistance(Cosine, a, b)
		if similarity != 0.0 {
			t.Errorf("Expected cosine similarity 0.0 for zero vector, got %f", similarity)
		}
	})

	// Test Euclidean distance
	t.Run("EuclideanDistance", func(t *testing.T) {
		// Identical vectors should have distance 0.0
		a := []float32{1.0, 2.0, 3.0}
		b := []float32{1.0, 2.0, 3.0}
		distance := CalculateDistance(Euclid, a, b)
		if math.Abs(float64(distance)) > 1e-6 {
			t.Errorf("Expected Euclidean distance 0.0, got %f", distance)
		}

		// Test distance between (0,0) and (3,4) should be 5.0
		a = []float32{0.0, 0.0}
		b = []float32{3.0, 4.0}
		distance = CalculateDistance(Euclid, a, b)
		if math.Abs(float64(distance-5.0)) > 1e-6 {
			t.Errorf("Expected Euclidean distance 5.0, got %f", distance)
		}
	})

	// Test Dot product
	t.Run("DotProduct", func(t *testing.T) {
		// Test dot product of [1,2,3] and [4,5,6] = 1*4 + 2*5 + 3*6 = 4+10+18=32
		a := []float32{1.0, 2.0, 3.0}
		b := []float32{4.0, 5.0, 6.0}
		product := CalculateDistance(Dot, a, b)
		if math.Abs(float64(product-32.0)) > 1e-6 {
			t.Errorf("Expected dot product 32.0, got %f", product)
		}

		// Test dot product of orthogonal vectors should be 0.0
		a = []float32{1.0, 0.0, 0.0}
		b = []float32{0.0, 1.0, 0.0}
		product = CalculateDistance(Dot, a, b)
		if math.Abs(float64(product)) > 1e-6 {
			t.Errorf("Expected dot product 0.0 for orthogonal vectors, got %f", product)
		}
	})

	// Test default metric (should be Cosine)
	t.Run("DefaultMetric", func(t *testing.T) {
		a := []float32{1.0, 0.0, 0.0}
		b := []float32{1.0, 0.0, 0.0}
		similarity := CalculateDistance("Unknown", a, b)
		if math.Abs(float64(similarity-1.0)) > 1e-6 {
			t.Errorf("Expected default metric (Cosine) similarity 1.0, got %f", similarity)
		}
	})
}
