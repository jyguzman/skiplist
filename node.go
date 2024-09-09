package skiplist

import (
	"fmt"
)

// SLItem a key-value pair in the skip list
type SLItem[K, V any] struct {
	key K
	val V
}

func (sl SLItem[K, V]) Key() K {
	return sl.key
}

func (sl SLItem[K, V]) Value() V {
	return sl.val
}

func (sl SLItem[K, V]) String() string {
	return fmt.Sprintf("Item{K: %v, V: %v}", sl.key, sl.val)
}

func NewItem[K, V any](key K, val V) SLItem[K, V] {
	return SLItem[K, V]{key: key, val: val}
}

// SLNode a node in the skip list that contains a key, value, and list of forward pointers
type SLNode[K, V any] struct {
	key      K
	val      V
	isHeader bool
	forward  []*SLNode[K, V]
	backward *SLNode[K, V] // a pointer to the previous node only on the bottom level
}

// Level return the highest level this node is in
func (sn *SLNode[K, V]) Level() int {
	return len(sn.forward) - 1
}

func (sn *SLNode[K, V]) String() string {
	return fmt.Sprintf("{key: %v, val: %v}", sn.key, sn.val)
}

// Item returns the key-value pair from this node
func (sn *SLNode[K, V]) Item() *SLItem[K, V] {
	return &SLItem[K, V]{sn.key, sn.val}
}

func newHeader[K, V any](maxLevel int) *SLNode[K, V] {
	header := &SLNode[K, V]{isHeader: true, forward: make([]*SLNode[K, V], maxLevel)}
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = nil
	}
	return header
}

func newNode[K, V any](level int, key K, val V) *SLNode[K, V] {
	return &SLNode[K, V]{
		key:     key,
		val:     val,
		forward: make([]*SLNode[K, V], level+1),
	}
}
