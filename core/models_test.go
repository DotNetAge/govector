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
}
