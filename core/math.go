package core

import "math"

// Distance represents the metric used for vector comparison
type Distance string

const (
	Cosine  Distance = "Cosine"
	Euclid  Distance = "Euclid"
	Dot     Distance = "Dot"
)

// CalculateDistance computes the similarity/distance based on the metric
func CalculateDistance(metric Distance, a, b []float32) float32 {
	switch metric {
	case Cosine:
		return cosineSimilarity(a, b)
	case Euclid:
		return euclideanDistance(a, b)
	case Dot:
		return dotProduct(a, b)
	default:
		return cosineSimilarity(a, b)
	}
}

func dotProduct(a, b []float32) float32 {
	var sum float32
	for i := range a {
		sum += a[i] * b[i]
	}
	return sum
}

func cosineSimilarity(a, b []float32) float32 {
	var dot, normA, normB float32
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / float32(math.Sqrt(float64(normA))*math.Sqrt(float64(normB)))
}

func euclideanDistance(a, b []float32) float32 {
	var sum float32
	for i := range a {
		diff := a[i] - b[i]
		sum += diff * diff
	}
	// For ranking, we often skip the Sqrt for efficiency, but let's keep it standard.
	// Note: For Euclid, smaller is better. We'll handle this in sorting.
	return float32(math.Sqrt(float64(sum)))
}
