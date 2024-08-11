package main

import (
	"cmp"
	"fmt"
	"math/rand"
	"sync"
)

type SkipList[K, V any] struct {
	m           sync.RWMutex
	maxLevel    int            // the maximum number of levels a node can appear on
	level       int            // the current highest level
	p           float64        // the chance from 0 to 1 that a node can appear at higher levels
	size        int            // the current number of elements
	compareFunc func(K, K) int // function used to compare keys
	header      *SLNode[K, V]  // the header node
	max         *SLItem[K, V]  // the element with the maximum key
}

// NewOrderedKeySkipList initializes a skip list using a cmp.Ordered key type with a given maxLevel and p.
func NewOrderedKeySkipList[K cmp.Ordered, V any](maxLevel int, p float64) *SkipList[K, V] {
	return &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: cmp.Compare[K],
	}
}

// NewCustomKeySkipList initializes a skip list using a custom key type that must implement Comparable
// and with a given maxLevel and p.
func NewCustomKeySkipList[K Comparable, V any](maxLevel int, p float64) *SkipList[K, V] {
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

// Min returns the element with the minimum key
func (sl *SkipList[K, V]) Min() *SLItem[K, V] {
	if sl.size == 0 {
		return nil
	}
	return sl.header.forward[0].Item()
}

// Max returns the element with the maximum key
func (sl *SkipList[K, V]) Max() *SLItem[K, V] {
	if sl.size == 0 {
		return nil
	}
	return sl.max
}

// Insert adds a given key & value to the skip list.
func (sl *SkipList[K, V]) Insert(key K, val V) {
	sl.m.RLock()
	update := make([]*SLNode[K, V], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, key) {
			x = x.forward[i]
		}
		update[i] = x
	}
	sl.m.RUnlock()

	sl.m.Lock()
	x = x.forward[0]
	if x != nil && sl.equal(x.key, key) {
		x.val = val
	} else {
		lvl := sl.randomLevel()
		//lvl := rand.Intn(sl.maxLevel)
		if lvl > sl.level {
			for i := sl.level + 1; i <= lvl; i++ {
				update[i] = sl.header
			}
			sl.level = lvl
		}

		x = newNode[K](lvl, key, val)
		for i := 0; i <= lvl; i++ {
			x.forward[i] = update[i].forward[i]
			update[i].forward[i] = x
		}

		if sl.max == nil || sl.greater(key, sl.max.Key) {
			sl.max = x.Item()
		}

		sl.size++
	}

	sl.m.Unlock()
}

// InsertAll bulk inserts an array of key-value pairs
func (sl *SkipList[K, V]) InsertAll([]SLItem[K, V]) {}

// Delete removes a given key & value from the skip list.
func (sl *SkipList[K, V]) Delete(key K) {
	sl.m.RLock()
	update, x := make([]*SLNode[K, V], sl.maxLevel), sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, key) {
			x = x.forward[i]
		}
		update[i] = x
	}
	sl.m.RUnlock()

	sl.m.Lock()
	x = x.forward[0]
	if x != nil && sl.equal(x.key, key) {
		if sl.equal(x.key, sl.max.Key) {
			sl.max = update[0].Item()
		}
		for i := 0; i <= sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		x = nil
		sl.size--
		for i := sl.level; i > 0 && sl.header.forward[sl.level] == nil; i-- {
			sl.level -= 1
		}
	}
	sl.m.Unlock()
}

// Search returns a value given by the key if it exists and a bool indicating if it exists
func (sl *SkipList[K, V]) Search(key K) (V, bool) {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, key) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	var val V
	if x != nil && sl.equal(x.key, key) {
		val = x.val
		return val, true
	}
	return val, false
}

// Range returns a list of key-pairs sorted from a minimum key to a maximum key.
func (sl *SkipList[K, V]) Range(min K, max K, leftInclusive bool, rightInclusive bool) []*SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, min) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	var result []*SLItem[K, V]
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
func (sl *SkipList[K, V]) RangeInc(min K, max K) []*SLItem[K, V] {
	return sl.Range(min, max, true, true)
}

// Merge combines this skip list with another and returns the result
func (sl *SkipList[K, V]) Merge(other *SkipList[K, V]) *SkipList[K, V] {
	result := NewSkipList[K, V](sl.maxLevel, sl.p, sl.compareFunc)
	var smaller *SkipList[K, V]
	if sl.size >= other.size {
		smaller = other
	} else {
		smaller = sl
	}

	fmt.Println(smaller)
	return result
}

// Split removes all elements with keys >= pivot from this list and returns them in a new list
func (sl *SkipList[K, V]) Split(pivot K) *SkipList[K, V] {
	sl.m.RLock()
	update, x := make([]*SLNode[K, V], sl.maxLevel), sl.header

	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, pivot) {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	sl.m.RUnlock()

	sl.m.Lock()
	newListHeader, newListLvl := newHeader[K, V](sl.maxLevel), 0
	for i := 0; i <= sl.level; i++ {
		newListHeader.forward[i] = update[i].forward[i]
		newListLvl = i
		update[i].forward[i] = nil
	}
	sl.m.Unlock()

	l2 := &SkipList[K, V]{
		maxLevel:    sl.maxLevel,
		level:       newListLvl,
		p:           sl.p,
		compareFunc: sl.compareFunc,
		header:      newListHeader,
	}

	l2Size, x := 0, l2.header
	for x.forward[0] != nil {
		l2Size += 1
		x = x.forward[0]
	}
	l2.size = l2Size

	sl.size = sl.size - l2Size
	return l2
}

// Rank returns the number of elements with key equal to or less than the given key, or - 1
// if the list is empty
func (sl *SkipList[K, V]) Rank(key K) int {
	sl.m.RLock()
	defer sl.m.RUnlock()

	if sl == nil || sl.size == 0 {
		return -1
	}
	return len(sl.RangeInc(sl.header.forward[0].key, key))
}

// Select returns the element with given rank
func (sl *SkipList[K, V]) Select(rank int) *SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	if sl.size == 0 || sl.size > rank {
		return nil
	}
	x := sl.header
	pos := 0
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && pos <= rank {
			x = x.forward[i]
		}
	}
	return x.Item()
}

// Successor returns the next item in sorted order from the element with given key
func (sl *SkipList[K, V]) Successor(key K) *SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, key) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	if x != nil {
		if sl.equal(x.key, key) && x.forward[0] != nil {
			return x.forward[0].Item()
		}
		return x.Item()
	}
	return nil
}

// Predecessor returns the previous item in sorted order from the element with given key
func (sl *SkipList[K, V]) Predecessor(key K) *SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	previous := make([]*SLNode[K, V], sl.maxLevel+1)
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, key) {
			x = x.forward[i]
		}
		previous[i] = x
	}
	if previous[0].isHeader {
		return nil
	}
	return previous[0].Item()
}

// Iterator returns a snapshot iterator over the skip list
func (sl *SkipList[K, V]) Iterator() Iterator[K, V] {
	return Iterator[K, V]{curr: sl.header.forward[0]}
}

// LazyDelete marks a key for deletion but does not actually remove the element. It is treated as
// deleted, e.g. a search for this key will return nil
func (sl *SkipList[K, V]) LazyDelete(key K) {}

// Clear removes all elements from the skip list
func (sl *SkipList[K, V]) Clear() {
	sl.m.Lock()

	sl.size = 0
	sl.level = 0
	sl.header = newHeader[K, V](sl.maxLevel)

	sl.m.Unlock()
}

func (sl *SkipList[K, V]) String() string {
	sl.m.RLock()

	res := ""
	for i := sl.level; i >= 0; i-- {
		head := sl.header
		res += "HEAD -> "
		level := head.forward[i]
		if level != nil {
			levelString := level.String()
			res += levelString
			for i < len(level.forward) && level.forward[i] != nil {
				res += " -> " + level.forward[i].String()
				level = level.forward[i]
			}
			res += "\n"
		}
	}

	sl.m.RUnlock()
	return res
}

// randomLevel returns the highest level a node will be assigned
func (sl *SkipList[K, V]) randomLevel() int {
	level := 0
	for i := 0; i < sl.maxLevel && rand.Float64() < sl.p; i++ {
		level++
	}
	return level
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
