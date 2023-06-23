package util

/**
 * Generic shared utilities
 */

// SliceIncludes returns true is string slice includes value
func SliceIncludes[T comparable](s []T, val T) bool {
	for _, v := range s {
		if v == val {
			return true
		}
	}
	return false
}
