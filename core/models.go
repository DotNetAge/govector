package core

// Payload mimics Qdrant's payload structure (map of string to any type)
type Payload map[string]interface{}

// PointStruct represents a vector data point compatible with Qdrant
type PointStruct struct {
	ID      string    `json:"id"`                // UUID or uint64 (using string for now)
	Vector  []float32 `json:"vector"`            // The actual embeddings
	Payload Payload   `json:"payload,omitempty"` // Metadata for filtering
}

// ScoredPoint is returned by search queries, containing the distance score
type ScoredPoint struct {
	ID      string  `json:"id"`
	Version uint64  `json:"version"`
	Score   float32 `json:"score"`
	Payload Payload `json:"payload,omitempty"`
}

// Qdrant Filtering Model (Simplified for MVP)
type Filter struct {
	Must    []Condition `json:"must,omitempty"`
	MustNot []Condition `json:"must_not,omitempty"`
}

type Condition struct {
	Key   string     `json:"key"`
	Match MatchValue `json:"match"`
}

type MatchValue struct {
	Value interface{} `json:"value"`
}

// Filter Matching Logic
func MatchFilter(payload Payload, filter *Filter) bool {
	if filter == nil {
		return true // No filter means match all
	}

	// Must conditions (all must match)
	for _, cond := range filter.Must {
		val, exists := payload[cond.Key]
		if !exists || val != cond.Match.Value {
			return false
		}
	}

	// MustNot conditions (none must match)
	for _, cond := range filter.MustNot {
		val, exists := payload[cond.Key]
		if exists && val == cond.Match.Value {
			return false
		}
	}

	return true
}
