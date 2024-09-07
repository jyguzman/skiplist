package skiplist

type iter[K, V any] struct {
	lessThan func(K, K) bool
	curr     *SLNode[K, V]
	seen     []*SLNode[K, V]
	seenIdx  int
	end      *K
}

type Iterator[K, V any] interface {
	Next() bool
	Prev() bool
	Key() K
	Value() V
	All() []SLItem[K, V] // returns an array of all key-value pairs covered by this iterator
}

func (it *iter[K, V]) hasNext() bool {
	if it.curr.forward[0] == nil {
		return false
	}
	if it.end != nil {
		return it.lessThan(it.curr.forward[0].key, *it.end)
	}
	return true
}

func (it *iter[K, V]) Next() bool {
	if it.seenIdx < len(it.seen)-1 {
		it.seenIdx++
		it.curr = it.seen[it.seenIdx]
		return true
	}
	if it.hasNext() {
		it.curr = it.curr.forward[0]
		it.seenIdx++
		if it.seenIdx >= len(it.seen) {
			it.seen = append(it.seen, it.curr)
		}
		return true
	}
	return false
}

func (it *iter[K, V]) Prev() bool {
	it.seenIdx--
	if it.seenIdx >= 0 {
		it.curr = it.seen[it.seenIdx]
		return true
	} else {
		it.seenIdx = 0
	}
	return false
}

func (it *iter[K, V]) Key() K {
	var key K
	if it.curr != nil {
		key = it.curr.key
	}
	return key
}

func (it *iter[K, V]) Value() V {
	var val V
	if it.curr != nil {
		val = it.curr.val
	}
	return val
}

// All returns an array of the key-value pairs covered by this iterator
func (it *iter[K, V]) All() []SLItem[K, V] {
	pairs := make([]SLItem[K, V], len(it.seen))
	for i, node := range it.seen {
		pairs[i] = *node.Item()
	}
	originalIdx := it.seenIdx
	it.seenIdx = len(it.seen) - 1
	for it.Next() {
		pairs = append(pairs, *it.curr.Item())
	}
	it.seenIdx = originalIdx
	return pairs
}
