package skiplist

// Iterator is a bidirectional iterator over the skip list.
type Iterator[K, V any] interface {
	// Next returns true if there are further nodes over which to iterate and
	// advances the iterator if there are
	Next() bool

	// Prev returns true if there are previous elements over which to iterate
	// and rewinds to the previous node if possible
	Prev() bool

	// Key returns the current key
	Key() K

	// Value returns the current value
	Value() V
}

type iter[K, V any] struct {
	lessThan    func(K, K) bool
	start       *SLNode[K, V]
	curr        *SLNode[K, V]
	rangeEndKey *K
}

func (it *iter[K, V]) hasNext() bool {
	if it.curr.forward[0] == nil {
		return false
	}
	if it.rangeEndKey != nil {
		return it.lessThan(it.curr.forward[0].key, *it.rangeEndKey)
	}
	return true
}

func (it *iter[K, V]) Next() bool {
	if it.hasNext() {
		it.curr = it.curr.forward[0]
		return true
	}
	return false
}

func (it *iter[K, V]) Prev() bool {
	if !it.curr.backward.isHeader {
		it.curr = it.curr.backward
		return true
	}
	return false
}

func (it *iter[K, V]) Key() K {
	return it.curr.key
}

func (it *iter[K, V]) Value() V {
	return it.curr.val
}
