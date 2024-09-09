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

// NewPair returns a new key-value pair.
func NewPair[K, V any](key K, val V) Pair[K, V] {
	return Pair[K, V]{key: key, val: val}
}

// slNode a node in the skip list that contains a key, value, and list of forward pointers
type slNode[K, V any] struct {
	key      K
	val      V
	isHeader bool
	forward  []*slNode[K, V]
	backward *slNode[K, V] // a pointer to the previous node only on the bottom level
}

// Level return the highest level this node is in
func (sn *slNode[K, V]) Level() int {
	return len(sn.forward) - 1
}

func (sn *slNode[K, V]) String() string {
	return fmt.Sprintf("{key: %v, val: %v}", sn.key, sn.val)
}

// Pair returns the key-value pair from this node
func (sn *slNode[K, V]) Pair() *Pair[K, V] {
	return &Pair[K, V]{sn.key, sn.val}
}

func newHeader[K, V any](maxLevel int) *slNode[K, V] {
	header := &slNode[K, V]{isHeader: true, forward: make([]*slNode[K, V], maxLevel)}
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = nil
	}
	return header
}

func newNode[K, V any](level int, key K, val V) *slNode[K, V] {
	return &slNode[K, V]{
		key:     key,
		val:     val,
		forward: make([]*slNode[K, V], level+1),
	}
}
