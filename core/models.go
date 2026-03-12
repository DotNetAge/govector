package core

import "regexp"

// Payload mimics Qdrant's payload structure, storing metadata as a map of string keys to any values.
// It is used for filtering points based on their associated metadata.
type Payload map[string]interface{}

// PointStruct represents a single vector data point, compatible with Qdrant's data model.
// It contains a unique identifier, the vector embedding, and optional metadata.
type PointStruct struct {
	ID      string    `json:"id"`                // UUID or uint64 (using string for now)
	Version uint64    `json:"version"`           // Incremental or timestamp-based version
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

// ConditionType defines the type of filter condition
type ConditionType string

const (
	// MatchTypeExact exact value match
	MatchTypeExact ConditionType = "exact"
	// MatchTypeRange range match (greater than, less than, etc.)
	MatchTypeRange ConditionType = "range"
	// MatchTypePrefix prefix match
	MatchTypePrefix ConditionType = "prefix"
	// MatchTypeContains contains match (for arrays)
	MatchTypeContains ConditionType = "contains"
	// MatchTypeRegex regex match
	MatchTypeRegex ConditionType = "regex"
)

// Condition represents a single filter condition on a specific payload key.
type Condition struct {
	Key   string        `json:"key"`
	Type  ConditionType `json:"type"`
	Match MatchValue    `json:"match,omitempty"`
	Range *RangeValue   `json:"range,omitempty"`
}

// MatchValue holds the value to match against in a filter condition.
type MatchValue struct {
	Value interface{} `json:"value"`
}

// RangeValue holds range values for range conditions
type RangeValue struct {
	GT  interface{} `json:"gt,omitempty"`  // Greater than
	GTE interface{} `json:"gte,omitempty"` // Greater than or equal
	LT  interface{} `json:"lt,omitempty"`  // Less than
	LTE interface{} `json:"lte,omitempty"` // Less than or equal
}

// CollectionMeta stores the configuration and metadata for a vector collection.
// It is used for persisting collection settings and reloading them on server restart.
type CollectionMeta struct {
	Name       string     `json:"name"`
	VectorLen  int        `json:"vector_size"`
	Metric     Distance   `json:"distance"`
	UseHNSW    bool       `json:"hnsw"`
	HNSWParams HNSWParams `json:"parameters"`
}

// MatchFilter evaluates whether a given payload matches the filter criteria.
// It returns true if:
//   - The filter is nil (no filtering)
//   - All Must conditions are satisfied
//   - No MustNot conditions are satisfied
func MatchFilter(payload Payload, filter *Filter) bool {
	if filter == nil {
		return true // No filter means match all
	}

	// Must conditions (all must match)
	for _, cond := range filter.Must {
		if !matchCondition(payload, cond) {
			return false
		}
	}

	// MustNot conditions (none must match)
	for _, cond := range filter.MustNot {
		if matchCondition(payload, cond) {
			return false
		}
	}

	return true
}

// matchCondition evaluates a single condition against a payload
func matchCondition(payload Payload, cond Condition) bool {
	val, exists := payload[cond.Key]
	if !exists {
		return false
	}

	switch cond.Type {
	case MatchTypeExact:
		return val == cond.Match.Value
	case MatchTypeRange:
		return matchRange(val, cond.Range)
	case MatchTypePrefix:
		return matchPrefix(val, cond.Match.Value)
	case MatchTypeContains:
		return matchContains(val, cond.Match.Value)
	case MatchTypeRegex:
		return matchRegex(val, cond.Match.Value)
	default:
		return val == cond.Match.Value
	}
}

// matchRange evaluates a range condition
func matchRange(val interface{}, r *RangeValue) bool {
	if r == nil {
		return false
	}

	switch v := val.(type) {
	case float64:
		if r.GT != nil {
			if gt, ok := r.GT.(float64); ok && v <= gt {
				return false
			}
		}
		if r.GTE != nil {
			if gte, ok := r.GTE.(float64); ok && v < gte {
				return false
			}
		}
		if r.LT != nil {
			if lt, ok := r.LT.(float64); ok && v >= lt {
				return false
			}
		}
		if r.LTE != nil {
			if lte, ok := r.LTE.(float64); ok && v > lte {
				return false
			}
		}
	case int:
		if r.GT != nil {
			if gt, ok := r.GT.(int); ok && v <= gt {
				return false
			}
			if gt, ok := r.GT.(float64); ok && float64(v) <= gt {
				return false
			}
		}
		if r.GTE != nil {
			if gte, ok := r.GTE.(int); ok && v < gte {
				return false
			}
			if gte, ok := r.GTE.(float64); ok && float64(v) < gte {
				return false
			}
		}
		if r.LT != nil {
			if lt, ok := r.LT.(int); ok && v >= lt {
				return false
			}
			if lt, ok := r.LT.(float64); ok && float64(v) >= lt {
				return false
			}
		}
		if r.LTE != nil {
			if lte, ok := r.LTE.(int); ok && v > lte {
				return false
			}
			if lte, ok := r.LTE.(float64); ok && float64(v) > lte {
				return false
			}
		}
	}

	return true
}

// matchPrefix evaluates a prefix condition
func matchPrefix(val interface{}, prefix interface{}) bool {
	strVal, ok := val.(string)
	if !ok {
		return false
	}

	strPrefix, ok := prefix.(string)
	if !ok {
		return false
	}

	if len(strPrefix) > len(strVal) {
		return false
	}

	return strVal[:len(strPrefix)] == strPrefix
}

// matchContains evaluates a contains condition
func matchContains(val interface{}, target interface{}) bool {
	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			if item == target {
				return true
			}
		}
	case []string:
		if targetStr, ok := target.(string); ok {
			for _, item := range v {
				if item == targetStr {
					return true
				}
			}
		}
	case []int:
		if targetInt, ok := target.(int); ok {
			for _, item := range v {
				if item == targetInt {
					return true
				}
			}
		}
	case string:
		if targetStr, ok := target.(string); ok {
			return len(targetStr) > 0 && containsString(v, targetStr)
		}
	}

	return false
}

// containsString checks if a string contains a substring
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// matchRegex evaluates a regex condition
func matchRegex(val interface{}, pattern interface{}) bool {
	strVal, ok := val.(string)
	if !ok {
		return false
	}

	strPattern, ok := pattern.(string)
	if !ok {
		return false
	}

	// Compile and match the regex pattern
	re, err := regexp.Compile(strPattern)
	if err != nil {
		return false
	}

	return re.MatchString(strVal)
}
