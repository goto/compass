package generichelper

import (
	"context"
	"reflect"

	"golang.org/x/sync/errgroup"
)

// Contains checks if a target item exists in an array of any type.
//
// Example:
//
//	names := []string{"Alice", "Bob", "Carol"}
//	result := Contains(names, "Bob")
//
// Result:
//
//	true
func Contains[T comparable](arr []T, target T) bool {
	for _, item := range arr {
		if item == target {
			return true
		}
	}
	return false
}

// GetJSONTags retrieves all JSON tag values from a struct.
// It returns a slice of JSON tag values extracted from the struct's fields.
//
// Example:
//
//		type Person struct {
//	  ID        int    `json:"id"`
//	  Name      string `json:"name"`
//	  Age       int    `json:"age"`
//	  CreatedAt string `json:"created_at"`
//		}
//
//		p := Person{}
//	 jsonTags := GetJSONTags(p)
//
// Result:
//
//	["id", "name", "age", "created_at"]
func GetJSONTags(v interface{}) []string {
	var tags []string
	val := reflect.ValueOf(v)
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			tags = append(tags, jsonTag)
		}
	}

	return tags
}

// GetMapKeys is a generic function that extracts all keys from a map and returns them in a slice.
//
// Example:
//
//	ageMap := map[string]int{"Alice": 30, "Bob": 25, "Carol": 27}
//	keys := GetMapKeys(ageMap)
//
// Result:
//
//	["Alice", "Bob", "Carol"]
func GetMapKeys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

// ChunkSlice splits a slice into smaller chunks of the specified size.
// It returns a slice of slices, where each inner slice has at most chunkSize elements.
// This is useful for processing large datasets in batches to avoid resource limits.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
//	chunks := ChunkSlice(numbers, 3)
//
// Result:
//
//	[[1, 2, 3], [4, 5, 6], [7, 8, 9], [10]]
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	if chunkSize <= 0 {
		return nil
	}

	if len(slice) == 0 {
		return [][]T{}
	}

	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}

	return chunks
}

// ProcessInChunksConcurrently processes a slice in chunks concurrently with a limit on concurrent goroutines.
// It returns an error if any chunk processing fails. Processing stops immediately on first error via context cancellation.
// The maxConcurrent parameter controls how many chunks can be processed simultaneously.
//
// Example:
//
//	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
//	err := ProcessInChunksConcurrently(ctx, numbers, 3, 3, func(chunk []int) error {
//	    fmt.Printf("Processing chunk: %v\n", chunk)
//	    time.Sleep(100 * time.Millisecond) // Simulate work
//	    return nil
//	})
//
// Result: Up to 3 chunks processed concurrently
//
//	Processing chunk: [1 2 3]
//	Processing chunk: [4 5 6]
//	Processing chunk: [7 8 9]
//	Processing chunk: [10 11 12]
func ProcessInChunksConcurrently[T any](
	ctx context.Context,
	slice []T,
	chunkSize int,
	maxConcurrent int,
	processFn func(chunk []T) error,
) error {
	if maxConcurrent <= 0 {
		maxConcurrent = 1
	}

	chunks := ChunkSlice(slice, chunkSize)
	if len(chunks) == 0 {
		return nil
	}

	// Use errgroup with context for proper error handling and cancellation
	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(maxConcurrent)

	// Process chunks concurrently
	for _, chunk := range chunks {
		chunk := chunk // Capture loop variable
		eg.Go(func() error {
			// Check if context is already canceled
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
			}

			// Process chunk
			return processFn(chunk)
		})
	}

	// Wait for all goroutines to complete and return first error if any
	return eg.Wait()
}
