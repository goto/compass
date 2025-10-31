package generichelper_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

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

func TestChunkSlice(t *testing.T) {
	tests := []struct {
		name      string
		input     []int
		chunkSize int
		expected  [][]int
	}{
		{
			name:      "Simple chunking",
			input:     []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			chunkSize: 3,
			expected:  [][]int{{1, 2, 3}, {4, 5, 6}, {7, 8, 9}, {10}},
		},
		{
			name:      "Exact division",
			input:     []int{1, 2, 3, 4, 5, 6},
			chunkSize: 2,
			expected:  [][]int{{1, 2}, {3, 4}, {5, 6}},
		},
		{
			name:      "Chunk size larger than slice",
			input:     []int{1, 2, 3},
			chunkSize: 10,
			expected:  [][]int{{1, 2, 3}},
		},
		{
			name:      "Empty slice",
			input:     []int{},
			chunkSize: 3,
			expected:  [][]int{},
		},
		{
			name:      "Chunk size of 1",
			input:     []int{1, 2, 3},
			chunkSize: 1,
			expected:  [][]int{{1}, {2}, {3}},
		},
		{
			name:      "Invalid chunk size (zero)",
			input:     []int{1, 2, 3},
			chunkSize: 0,
			expected:  nil,
		},
		{
			name:      "Invalid chunk size (negative)",
			input:     []int{1, 2, 3},
			chunkSize: -1,
			expected:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generichelper.ChunkSlice(tt.input, tt.chunkSize)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ChunkSlice(%v, %d) = %v; want %v", tt.input, tt.chunkSize, result, tt.expected)
			}
		})
	}
}

func TestChunkSliceWithStrings(t *testing.T) {
	input := []string{"a", "b", "c", "d", "e"}
	chunkSize := 2
	expected := [][]string{{"a", "b"}, {"c", "d"}, {"e"}}

	result := generichelper.ChunkSlice(input, chunkSize)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("ChunkSlice(%v, %d) = %v; want %v", input, chunkSize, result, expected)
	}
}

func TestProcessInChunksConcurrently(t *testing.T) {
	t.Run("Successful concurrent processing", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		chunkSize := 3
		maxConcurrent := 3

		var mu sync.Mutex
		var processedChunks [][]int
		var concurrentCount int32
		var maxConcurrentReached int32

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(chunk []int) error {
			// Track concurrent goroutines
			current := atomic.AddInt32(&concurrentCount, 1)
			defer atomic.AddInt32(&concurrentCount, -1)

			// Track max concurrent
			for {
				maxConcurrent := atomic.LoadInt32(&maxConcurrentReached)
				if current <= maxConcurrent || atomic.CompareAndSwapInt32(&maxConcurrentReached, maxConcurrent, current) {
					break
				}
			}

			// Simulate work
			time.Sleep(50 * time.Millisecond)

			mu.Lock()
			processedChunks = append(processedChunks, chunk)
			mu.Unlock()

			return nil
		})
		if err != nil {
			t.Errorf("ProcessInChunksConcurrently returned unexpected error: %v", err)
		}

		if len(processedChunks) != 4 {
			t.Errorf("Expected 4 chunks to be processed, got %d", len(processedChunks))
		}

		maxReached := atomic.LoadInt32(&maxConcurrentReached)
		if maxReached > int32(maxConcurrent) {
			t.Errorf("Max concurrent exceeded: expected <= %d, got %d", maxConcurrent, maxReached)
		}
	})

	t.Run("Error handling in concurrent processing", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		chunkSize := 3
		maxConcurrent := 3
		expectedError := errors.New("processing error")

		var processedCount int32

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(chunk []int) error {
			atomic.AddInt32(&processedCount, 1)
			time.Sleep(10 * time.Millisecond)

			// Fail on chunk containing 5
			for _, v := range chunk {
				if v == 5 {
					return expectedError
				}
			}
			return nil
		})

		if !errors.Is(err, expectedError) {
			t.Errorf("ProcessInChunksConcurrently error = %v; want %v", err, expectedError)
		}

		// Should have processed at least 2 chunks (one with error)
		if atomic.LoadInt32(&processedCount) < 2 {
			t.Errorf("Expected at least 2 chunks processed, got %d", processedCount)
		}
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		input := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		chunkSize := 3
		maxConcurrent := 3

		var processedCount int32

		// Cancel context immediately to test cancellation
		cancel()

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(_ []int) error {
			atomic.AddInt32(&processedCount, 1)
			time.Sleep(50 * time.Millisecond)
			return nil
		})

		if err == nil {
			t.Error("Expected context cancellation error, got nil")
		}

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled error, got %v", err)
		}
	})

	t.Run("Empty slice", func(t *testing.T) {
		ctx := context.Background()
		var input []int
		chunkSize := 3
		maxConcurrent := 3
		callCount := 0

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(_ []int) error {
			callCount++
			return nil
		})
		if err != nil {
			t.Errorf("ProcessInChunksConcurrently returned unexpected error: %v", err)
		}

		if callCount != 0 {
			t.Errorf("ProcessInChunksConcurrently called function %d times for empty slice; want 0", callCount)
		}
	})

	t.Run("MaxConcurrent zero defaults to 1", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3, 4, 5, 6}
		chunkSize := 2
		maxConcurrent := 0 // Should default to 1

		var concurrentCount int32
		var maxConcurrentReached int32

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(_ []int) error {
			current := atomic.AddInt32(&concurrentCount, 1)
			defer atomic.AddInt32(&concurrentCount, -1)

			// Track max concurrent
			for {
				maxConcurrent := atomic.LoadInt32(&maxConcurrentReached)
				if current <= maxConcurrent || atomic.CompareAndSwapInt32(&maxConcurrentReached, maxConcurrent, current) {
					break
				}
			}

			time.Sleep(50 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Errorf("ProcessInChunksConcurrently returned unexpected error: %v", err)
		}

		maxReached := atomic.LoadInt32(&maxConcurrentReached)
		if maxReached > 1 {
			t.Errorf("Max concurrent should default to 1, got %d", maxReached)
		}
	})

	t.Run("Single chunk processes correctly", func(t *testing.T) {
		ctx := context.Background()
		input := []int{1, 2, 3}
		chunkSize := 10
		maxConcurrent := 3
		processed := false

		err := generichelper.ProcessInChunksConcurrently(ctx, input, chunkSize, maxConcurrent, func(chunk []int) error {
			if !reflect.DeepEqual(chunk, input) {
				t.Errorf("Expected chunk %v, got %v", input, chunk)
			}
			processed = true
			return nil
		})
		if err != nil {
			t.Errorf("ProcessInChunksConcurrently returned unexpected error: %v", err)
		}

		if !processed {
			t.Error("Chunk was not processed")
		}
	})
}
