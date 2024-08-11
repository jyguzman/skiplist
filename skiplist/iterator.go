package skiplist

type Iterator[K, V any] struct {
	curr *SLNode[K, V]
}

func (it *Iterator[K, V]) skipTombstones() {
	for it.curr != nil && it.curr.markedDeleted {
		it.curr = it.curr.forward[0]
	}
}

func (it *Iterator[K, V]) Next() *SLItem[K, V] {
	it.skipTombstones()
	if it.curr == nil {
		return nil
	}
	res := it.curr.Item()
	it.skipTombstones()
	if it.curr != nil {
		it.curr = it.curr.forward[0]
	}
	return res
}

func (it *Iterator[K, V]) HasNext() bool {
	return it.curr != nil
}
