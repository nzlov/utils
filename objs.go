// Package utils provides utility functions for common operations.
package utils

import "sync"

// ObjsTrans concurrently transforms a slice of type K to a slice of type T
// using the provided transformation function fn.
//
// This function processes each element of the input slice in parallel using goroutines,
// waits for all transformations to complete, and returns the resulting slice.
//
// Parameters:
//   - objs: Input slice of type K to be transformed
//   - fn: Transformation function that converts a single element of type K to type T
//
// Returns:
//   - []T: Slice containing the transformed elements in the same order as the input
//
// Example:
//
//	// Convert []int to []string by converting each integer to its string representation
//	nums := []int{1, 2, 3, 4, 5}
//	strs := ObjsTrans(nums, func(n int) string {
//	    return fmt.Sprintf("Number: %d", n)
//	})
func ObjsTrans[K, T any](objs []K, fn func(K) T) []T {
	wg := &sync.WaitGroup{}
	wg.Add(len(objs))
	resps := make([]T, len(objs))
	for i := range objs {
		go func(i int) {
			defer wg.Done()
			v := objs[i]
			resps[i] = fn(v)
		}(i)
	}
	wg.Wait()
	return resps
}
