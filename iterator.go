package skiplist

type Iterator[K, V any] struct {
	compareFunc func(K, K) int
	curr        *SLNode[K, V]
}

func (it *Iterator[K, V]) Next() *SLItem[K, V] {
	defer it.advance()
	if it.HasNext() {
		return it.curr.Item()
	}
	return nil
}

func (it *Iterator[K, V]) next() *SLNode[K, V] {
	defer it.advance()
	if it.HasNext() {
		return it.curr
	}
	return nil
}

func (it *Iterator[K, V]) HasNext() bool {
	return it.curr != nil
}

func (it *Iterator[K, V]) All() []SLItem[K, V] {
	var results []SLItem[K, V]
	for it.HasNext() {
		it.skipTombstones()
		results = append(results, *it.curr.Item())
		it.advance()
	}
	return results
}

func (it *Iterator[K, V]) advance() {
	it.skipTombstones()
	if it.curr != nil {
		it.curr = it.curr.forward[0]
	}
	it.skipTombstones()
}

func (it *Iterator[K, V]) skipTombstones() {
	for it.curr != nil && it.curr.markedDeleted {
		it.curr = it.curr.forward[0]
	}
}

func (it *Iterator[K, V]) UpTo(stop K) []SLItem[K, V] {
	var results []SLItem[K, V]
	for it.HasNext() && it.compareFunc(it.curr.key, stop) <= 0 {
		it.skipTombstones()
		results = append(results, *it.curr.Item())
		it.advance()
	}
	return results
}
