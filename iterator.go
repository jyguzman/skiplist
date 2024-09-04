package skiplist

type Iterator[K, V any] struct {
	less func(K, K) bool
	curr *SLNode[K, V]
}

func (it *Iterator[K, V]) Next() *SLItem[K, V] {
	defer it.advance()
	if it.HasNext() {
		return it.curr.Item()
	}
	return nil
}

func (it *Iterator[K, V]) Key() K {
	var key K
	if it.HasNext() {
		key = it.curr.Item().Key
	}
	return key
}

func (it *Iterator[K, V]) Value() V {
	var val V
	if it.HasNext() {
		val = it.curr.Item().Val
	}
	return val
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
		results = append(results, *it.curr.Item())
		it.advance()
	}
	return results
}

func (it *Iterator[K, V]) advance() {
	if it.curr != nil {
		it.curr = it.curr.forward[0]
	}
}

func (it *Iterator[K, V]) UpTo(stop K) []SLItem[K, V] {
	var results []SLItem[K, V]
	for it.HasNext() && it.less(it.curr.key, stop) {
		results = append(results, *it.curr.Item())
		it.advance()
	}
	return results
}
