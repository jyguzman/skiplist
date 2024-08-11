package skiplist

type Iterator[K, V any] struct {
	curr *SLNode[K, V]
}

func (iter *Iterator[K, V]) Next() *SLItem[K, V] {
	if !iter.HasNext() {
		return nil
	}
	res := iter.curr.Item()
	iter.curr = iter.curr.forward[0]
	return res
}

func (iter *Iterator[K, V]) HasNext() bool {
	return iter.curr != nil
}
