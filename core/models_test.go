package core

import (
	"testing"
)

func TestMatchFilter(t *testing.T) {
	// Test payload
	payload := Payload{
		"category": "electronics",
		"price":    100,
		"in_stock": true,
		"tags":     []string{"gadget", "device"},
		"name":     "smartphone",
	}

	// Test case 1: No filter (should match all)
	if !MatchFilter(payload, nil) {
		t.Error("Expected no filter to match all")
	}

	// Test case 2: Must condition - match
	filter := &Filter{
		Must: []Condition{
			{Key: "category", Match: MatchValue{Value: "electronics"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected must condition to match")
	}

	// Test case 3: Must condition - no match
	filter = &Filter{
		Must: []Condition{
			{Key: "category", Match: MatchValue{Value: "clothing"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected must condition to not match")
	}

	// Test case 4: MustNot condition - no match (should pass)
	filter = &Filter{
		MustNot: []Condition{
			{Key: "category", Match: MatchValue{Value: "clothing"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected mustNot condition to pass when value doesn't match")
	}

	// Test case 5: MustNot condition - match (should fail)
	filter = &Filter{
		MustNot: []Condition{
			{Key: "category", Match: MatchValue{Value: "electronics"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected mustNot condition to fail when value matches")
	}

	// Test case 6: Combined must and mustNot conditions - match
	filter = &Filter{
		Must: []Condition{
			{Key: "category", Match: MatchValue{Value: "electronics"}},
			{Key: "in_stock", Match: MatchValue{Value: true}},
		},
		MustNot: []Condition{
			{Key: "price", Match: MatchValue{Value: 200}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected combined conditions to match")
	}

	// Test case 7: Combined must and mustNot conditions - no match (must condition fails)
	filter = &Filter{
		Must: []Condition{
			{Key: "category", Match: MatchValue{Value: "electronics"}},
			{Key: "price", Match: MatchValue{Value: 200}},
		},
		MustNot: []Condition{
			{Key: "in_stock", Match: MatchValue{Value: false}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected combined conditions to not match when must condition fails")
	}

	// Test case 8: Combined must and mustNot conditions - no match (mustNot condition fails)
	filter = &Filter{
		Must: []Condition{
			{Key: "category", Match: MatchValue{Value: "electronics"}},
		},
		MustNot: []Condition{
			{Key: "price", Match: MatchValue{Value: 100}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected combined conditions to not match when mustNot condition fails")
	}

	// Test case 9: Non-existent key in must condition (should fail)
	filter = &Filter{
		Must: []Condition{
			{Key: "non_existent", Match: MatchValue{Value: "test"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected must condition with non-existent key to fail")
	}

	// Test case 10: Non-existent key in mustNot condition (should pass)
	filter = &Filter{
		MustNot: []Condition{
			{Key: "non_existent", Match: MatchValue{Value: "test"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected mustNot condition with non-existent key to pass")
	}

	// Test case 11: Range match - within range
	filter = &Filter{
		Must: []Condition{
			{Key: "price", Type: MatchTypeRange, Range: &RangeValue{GTE: 50, LTE: 150}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected range condition to match")
	}

	// Test case 12: Range match - outside range
	filter = &Filter{
		Must: []Condition{
			{Key: "price", Type: MatchTypeRange, Range: &RangeValue{GTE: 150, LTE: 200}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected range condition to not match")
	}

	// Test case 13: Prefix match - match
	filter = &Filter{
		Must: []Condition{
			{Key: "name", Type: MatchTypePrefix, Match: MatchValue{Value: "smart"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected prefix condition to match")
	}

	// Test case 14: Prefix match - no match
	filter = &Filter{
		Must: []Condition{
			{Key: "name", Type: MatchTypePrefix, Match: MatchValue{Value: "dumb"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected prefix condition to not match")
	}

	// Test case 15: Contains match - match
	filter = &Filter{
		Must: []Condition{
			{Key: "tags", Type: MatchTypeContains, Match: MatchValue{Value: "gadget"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected contains condition to match")
	}

	// Test case 16: Contains match - no match
	filter = &Filter{
		Must: []Condition{
			{Key: "tags", Type: MatchTypeContains, Match: MatchValue{Value: "accessory"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected contains condition to not match")
	}

	// Test case 17: Regex match - match
	filter = &Filter{
		Must: []Condition{
			{Key: "name", Type: MatchTypeRegex, Match: MatchValue{Value: "^smart.*"}},
		},
	}

	if !MatchFilter(payload, filter) {
		t.Error("Expected regex condition to match")
	}

	// Test case 18: Regex match - no match
	filter = &Filter{
		Must: []Condition{
			{Key: "name", Type: MatchTypeRegex, Match: MatchValue{Value: "^dumb.*"}},
		},
	}

	if MatchFilter(payload, filter) {
		t.Error("Expected regex condition to not match")
	}
}

func TestMatchRange(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		rangeVal *RangeValue
		expected bool
	}{
		{"nil range", 10.0, nil, false},
		{"float GT match", 10.0, &RangeValue{GT: 5.0}, true},
		{"float GT no match", 10.0, &RangeValue{GT: 15.0}, false},
		{"float GTE match", 10.0, &RangeValue{GTE: 10.0}, true},
		{"float GTE no match", 10.0, &RangeValue{GTE: 11.0}, false},
		{"float LT match", 10.0, &RangeValue{LT: 15.0}, true},
		{"float LT no match", 10.0, &RangeValue{LT: 5.0}, false},
		{"float LTE match", 10.0, &RangeValue{LTE: 10.0}, true},
		{"float LTE no match", 10.0, &RangeValue{LTE: 9.0}, false},
		{"int GT match", 10, &RangeValue{GT: 5}, true},
		{"int GT float match", 10, &RangeValue{GT: 5.0}, true},
		{"int GT no match", 10, &RangeValue{GT: 15}, false},
		{"int GTE match", 10, &RangeValue{GTE: 10}, true},
		{"int GTE float match", 10, &RangeValue{GTE: 10.0}, true},
		{"int GTE no match", 10, &RangeValue{GTE: 11}, false},
		{"int LT match", 10, &RangeValue{LT: 15}, true},
		{"int LT float match", 10, &RangeValue{LT: 15.0}, true},
		{"int LT no match", 10, &RangeValue{LT: 5}, false},
		{"int LTE match", 10, &RangeValue{LTE: 10}, true},
		{"int LTE float match", 10, &RangeValue{LTE: 10.0}, true},
		{"int LTE no match", 10, &RangeValue{LTE: 9}, false},
		{"unsupported type", "string", &RangeValue{GT: 5}, true}, // matchRange returns true for unsupported types if r != nil and it falls through
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchRange(tt.val, tt.rangeVal); got != tt.expected {
				t.Errorf("matchRange() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchPrefix(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		prefix   interface{}
		expected bool
	}{
		{"string match", "hello world", "hello", true},
		{"string no match", "hello world", "world", false},
		{"prefix longer than val", "hi", "hello", false},
		{"non-string val", 123, "12", false},
		{"non-string prefix", "hello", 12, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchPrefix(tt.val, tt.prefix); got != tt.expected {
				t.Errorf("matchPrefix() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchContains(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		target   interface{}
		expected bool
	}{
		{"[]interface{} match", []interface{}{"a", "b", "c"}, "b", true},
		{"[]interface{} no match", []interface{}{"a", "b", "c"}, "d", false},
		{"[]string match", []string{"a", "b", "c"}, "b", true},
		{"[]string no match", []string{"a", "b", "c"}, "d", false},
		{"[]int match", []int{1, 2, 3}, 2, true},
		{"[]int no match", []int{1, 2, 3}, 4, false},
		{"string match", "hello world", "world", true},
		{"string no match", "hello world", "hi", false},
		{"string empty target", "hello", "", false},
		{"unsupported type", 123, "target", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchContains(tt.val, tt.target); got != tt.expected {
				t.Errorf("matchContains() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMatchRegex(t *testing.T) {
	tests := []struct {
		name     string
		val      interface{}
		pattern  interface{}
		expected bool
	}{
		{"regex match", "hello world", "^hello.*", true},
		{"regex no match", "hello world", "^world.*", false},
		{"invalid regex", "hello", "[", false},
		{"non-string val", 123, "^123", false},
		{"non-string pattern", "hello", 123, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchRegex(tt.val, tt.pattern); got != tt.expected {
				t.Errorf("matchRegex() = %v, want %v", got, tt.expected)
			}
		})
	}
}
