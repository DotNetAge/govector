package core

// Payload mimics Qdrant's payload structure, storing metadata as a map of string keys to any values.
// It is used for filtering points based on their associated metadata.
type Payload map[string]interface{}

// PointStruct represents a single vector data point, compatible with Qdrant's data model.
// It contains a unique identifier, the vector embedding, and optional metadata.
type PointStruct struct {
	ID      string    `json:"id"`                // UUID or uint64 (using string for now)
	Vector  []float32 `json:"vector"`            // The actual embeddings
	Payload Payload   `json:"payload,omitempty"` // Metadata for filtering
}

// ScoredPoint is returned by search queries, containing the distance score and point data.
// The Score field represents the computed similarity/distance based on the collection's metric.
type ScoredPoint struct {
	ID      string  `json:"id"`
	Version uint64  `json:"version"`
	Score   float32 `json:"score"`
	Payload Payload `json:"payload,omitempty"`
}

// Filter defines conditions for querying or deleting points based on their payload.
// It supports Must (all conditions must match) and MustNot (all conditions must not match) clauses.
type Filter struct {
	Must    []Condition `json:"must,omitempty"`
	MustNot []Condition `json:"must_not,omitempty"`
}

// Condition represents a single filter condition on a specific payload key.
type Condition struct {
	Key   string     `json:"key"`
	Match MatchValue `json:"match"`
}

// MatchValue holds the value to match against in a filter condition.
type MatchValue struct {
	Value interface{} `json:"value"`
}

// MatchFilter evaluates whether a given payload matches the filter criteria.
// It returns true if:
//   - The filter is nil (no filtering)
//   - All Must conditions are satisfied (payload contains key with matching value)
//   - No MustNot conditions are satisfied (payload does not contain key with matching value)
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
