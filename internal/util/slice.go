package util

/**
 * Generic shared utilities
 */

// SliceIncludes helper for detecting if a slice includes a value
func SliceIncludes[T comparable](s []T, val T) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}
