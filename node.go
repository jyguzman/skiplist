package skiplist

import (
	"fmt"
)

// Pair is a key-value pair, or element, in the skip list.
type Pair[K, V any] struct {
	key K
	val V
}

// Key returns the key of this key-value pair.
func (p Pair[K, V]) Key() K {
	return p.key
}

// Value returns the value of this key-value pair.
func (p Pair[K, V]) Value() V {
	return p.val
}

// String returns a string representation of the key-value pair.
func (p Pair[K, V]) String() string {
	return fmt.Sprintf("Pair{K: %v, V: %v}", p.key, p.val)
}

// NewPair returns a key-value pair.
func NewPair[K, V any](key K, val V) Pair[K, V] {
	return Pair[K, V]{key: key, val: val}
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

// Pair returns the key-value pair from this node
func (sn *SLNode[K, V]) Pair() *Pair[K, V] {
	return &Pair[K, V]{sn.key, sn.val}
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
