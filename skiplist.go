package main

import (
	"cmp"
	"math/rand"
	"sync"
)

type SkipList[K, V any] struct {
	m           sync.RWMutex
	maxLevel    int
	level       int
	p           float64
	size        int
	compareFunc func(K, K) int
	header      *SLNode[K, V]
}

// NewBasicSkipList initializes a skip list using an ordered primitive key type with a given maxLevel and p.
func NewBasicSkipList[K cmp.Ordered, V any](maxLevel int, p float64) *SkipList[K, V] {
	return &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: cmp.Compare[K],
	}
}

// NewCustomSkipList initializes a skip list using a custom key type with a given maxLevel and p.
func NewCustomSkipList[K Comparable, V any](maxLevel int, p float64) *SkipList[K, V] {
	return &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: Compare[K],
	}
}

// NewSkipList initializes a skip list with a given maxLevel and p. Must supply a comparator function
// for the K type: cmp.Compare[K] for a primitive ordered type, or Compare[K] for a custom type
func NewSkipList[K, V any](maxLevel int, p float64, compareFunc func(K, K) int) *SkipList[K, V] {
	return &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: compareFunc,
	}
}

// Size returns the number of elements in the skip list
func (sl *SkipList[K, V]) Size() int {
	return sl.size
}

// MaxLevel returns the maximum numbers of forward pointers a node can have
func (sl *SkipList[K, V]) MaxLevel() int {
	return sl.maxLevel
}

// Level returns the current highest level of the list
func (sl *SkipList[K, V]) Level() int {
	return sl.level
}

// P returns the chance that a node is inserted into a higher level
func (sl *SkipList[K, V]) P() float64 {
	return sl.p
}

// randomLevel returns the highest level a node will be assigned
func (sl *SkipList[K, V]) randomLevel() int {
	level := 0
	for i := 0; i < sl.maxLevel && rand.Float64() < sl.p; i++ {
		level++
	}
	return level
}

// Insert adds a given key & value to the skip list.
func (sl *SkipList[K, V]) Insert(searchKey K, val V) {
	sl.m.Lock()

	update := make([]*SLNode[K, V], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, searchKey) {
			x = x.forward[i]
		}
		update[i] = x
	}

	x = x.forward[0]
	if x != nil && sl.equal(x.key, searchKey) {
		x.val = val
	} else {
		lvl := sl.randomLevel()
		if lvl > sl.level {
			for i := sl.level + 1; i <= lvl; i++ {
				update[i] = sl.header
			}
			sl.level = lvl
		}

		sl.size++
		x = newNode[K](lvl, searchKey, val)
		for i := 0; i <= lvl; i++ {
			x.forward[i] = update[i].forward[i]
			update[i].forward[i] = x
		}
	}

	sl.m.Unlock()
}

// Delete remove a given key & value from the skip list.
func (sl *SkipList[K, V]) Delete(searchKey K) {
	sl.m.Lock()

	update, x := make([]*SLNode[K, V], sl.maxLevel), sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, searchKey) {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if x != nil && sl.equal(x.key, searchKey) {
		for i := 0; i <= sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		x = nil
		sl.size--
		for i := sl.level; i > 0 && sl.header.forward[sl.level] == nil; i-- {
			sl.level = sl.level - 1
		}
	}

	sl.m.Unlock()
}

// Search returns a value given by the key if it exists and a bool indicating if it exists
func (sl *SkipList[K, V]) Search(searchKey K) (V, bool) {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, searchKey) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	var val V
	if x != nil && sl.equal(x.key, searchKey) {
		val = x.val
		return val, true
	}
	return val, false
}

// Range returns a list of key-pairs sorted from a minimum key to a maximum key.
func (sl *SkipList[K, V]) Range(min K, max K, leftInclusive bool, rightInclusive bool) []SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, min) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	var result []SLItem[K, V]
	if x != nil && sl.geq(x.key, min) {
		if leftInclusive && sl.equal(x.key, min) {
			result = append(result, x.Item())
		}
		for x.forward[0] != nil && sl.less(x.forward[0].key, max) {
			result = append(result, x.forward[0].Item())
			x = x.forward[0]
		}
		x = x.forward[0]
		if rightInclusive && x != nil && sl.equal(x.key, max) {
			result = append(result, x.Item())
		}
	}

	return result
}

// RangeInc Range inclusive of min and max
func (sl *SkipList[K, V]) RangeInc(min K, max K) []SLItem[K, V] {
	return sl.Range(min, max, true, true)
}

func (sl *SkipList[K, V]) Merge(other *SkipList[K, V]) {}

// Split splits the skip list into two: a skip list containing items with key up to pivot,
// and a skip list with items from the pivot
func (sl *SkipList[K, V]) Split(pivot K) (*SkipList[K, V], *SkipList[K, V]) {
	return nil, nil
}

// Rank returns the number of elements with key equal to or less than the given key
func (sl *SkipList[K, V]) Rank(key K) int {
	return len(sl.RangeInc(sl.header.forward[0].key, key))
}

// Select returns the element with given rank
func (sl *SkipList[K, V]) Select(rank int) *SLItem[K, V] {
	if sl.size == 0 || sl.size > rank {
		x := sl.header
		for i := 1; i < rank; i++ {
			x = x.forward[0]
		}
		return &SLItem[K, V]{x.key, x.val}
	}
	return nil
}

func (sl *SkipList[K, V]) Successor(key K) SLItem[K, V] { return SLItem[K, V]{} }

func (sl *SkipList[K, V]) Predecessor(key K) SLItem[K, V] { return SLItem[K, V]{} }

// LazyDelete marks a key for deletion but does not actually remove the element
func (sl *SkipList[K, V]) LazyDelete(key K) {}

func (sl *SkipList[K, V]) String() string {
	sl.m.RLock()

	res := ""
	for i := sl.level; i >= 0; i-- {
		level := sl.header.forward[i]
		res += level.String()
		for i < len(level.forward) && level.forward[i] != nil {
			res += " -> " + level.forward[i].String()
			level = level.forward[i]
		}
		res += "\n"
	}

	sl.m.RUnlock()
	return res
}

// less returns true if x < y
func (sl *SkipList[K, V]) less(x, y K) bool {
	return sl.compareFunc(x, y) == -1
}

// equal returns true if x == y
func (sl *SkipList[K, V]) equal(x, y K) bool {
	return sl.compareFunc(x, y) == 0
}

// greater returns true if x > y
func (sl *SkipList[K, V]) greater(x, y K) bool {
	return sl.compareFunc(x, y) == 1
}

// leq returns true if x <= y
func (sl *SkipList[K, V]) leq(x, y K) bool {
	return sl.less(x, y) || sl.equal(x, y)
}

// geq returns true if x >= y
func (sl *SkipList[K, V]) geq(x, y K) bool {
	return sl.greater(x, y) || sl.equal(x, y)
}
