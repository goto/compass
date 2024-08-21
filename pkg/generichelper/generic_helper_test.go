package generichelper_test

import (
	"reflect"
	"testing"

	"github.com/goto/compass/pkg/generichelper"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		arr      []string
		target   string
		expected bool
	}{
		{"Found", []string{"Alice", "Bob", "Carol"}, "Bob", true},
		{"Not Found", []string{"Alice", "Bob", "Carol"}, "Dave", false},
		{"Empty Array", []string{}, "Bob", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generichelper.Contains(tt.arr, tt.target)
			if result != tt.expected {
				t.Errorf("Contains(%v, %v) = %v; want %v", tt.arr, tt.target, result, tt.expected)
			}
		})
	}
}

func TestGetJSONTags(t *testing.T) {
	type Person struct {
		ID        int    `json:"id"`
		Name      string `json:"name"`
		Age       int    `json:"age"`
		CreatedAt string `json:"created_at"`
	}

	p := Person{}
	expectedTags := []string{"id", "name", "age", "created_at"}

	result := generichelper.GetJSONTags(p)
	if !reflect.DeepEqual(result, expectedTags) {
		t.Errorf("GetJSONTags(%v) = %v; want %v", p, result, expectedTags)
	}
}

func TestGetMapKeys(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]int
		expected []string
	}{
		{"Simple Map", map[string]int{"Alice": 30, "Bob": 25, "Carol": 27}, []string{"Alice", "Bob", "Carol"}},
		{"Empty Map", map[string]int{}, nil},
		{"Single Element", map[string]int{"Alice": 30}, []string{"Alice"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generichelper.GetMapKeys(tt.input)
			if !CompareSlices(result, tt.expected) {
				t.Errorf("GetMapKeys(%v) = %v; want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// CompareSlices checks if two slices contain the same elements, regardless of order.
func CompareSlices[T comparable](a, b []T) bool {
	if a == nil && b == nil {
		return true
	}
	if len(a) != len(b) {
		return false
	}

	counts := make(map[T]int)

	for _, item := range a {
		counts[item]++
	}

	for _, item := range b {
		if counts[item] == 0 {
			return false
		}
		counts[item]--
	}

	return true
}
