package skiplist

type Iterator[K, V any] struct {
	next *SLNode[K, V]
}

func (it *Iterator[K, V]) skipTombstones() {
	for it.next != nil && it.next.markedDeleted {
		it.next = it.next.forward[0]
	}
}

func (it *Iterator[K, V]) Next() *SLItem[K, V] {
	it.skipTombstones()
	if it.next == nil {
		return nil
	}
	res := it.next.Item()
	it.skipTombstones()
	if it.next != nil {
		it.next = it.next.forward[0]
	}
	return res
}

func (it *Iterator[K, V]) HasNext() bool {
	return it.next != nil
}
