package skiplist

// Compare a comparator function for a user-defined Comparable type that
// returns 1 if a > b, 0 if a == b, or -1 if a < b
func Compare[K Comparable](a, b K) int {
	return a.Cmp(b)
}

// Comparable interface that user-defined key types should implement
type Comparable interface {
	// Cmp a comparator function that should return 1 if a > b, 0 if a == b, or -1 if a < b
	Cmp(Comparable) int
}
