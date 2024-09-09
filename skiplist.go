package skiplist

import (
	"cmp"
	"log"
	"math/bits"
	"math/rand"
	"strings"
	"sync"
)

const DefaultMaxLevel = 32
const AbsoluteMaxLevel = 64

type SkipList[K, V any] struct {
	rw       sync.RWMutex
	maxLevel int             // the maximum number of levels a node can appear on
	level    int             // the current highest level
	size     int             // the current number of elements
	lessThan func(K, K) bool // function used to compare keys
	header   *slNode[K, V]   // the header node
	max      *slNode[K, V]   // the node with the maximum key, which can also be considered the "end" or "back" of the list
}

// NewSkipList initializes a skip list using a cmp.Ordered key type and with a default max level of 32.
// Optionally include items with which to initialize the list.
func NewSkipList[K cmp.Ordered, V any](items ...Pair[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel: DefaultMaxLevel - 1,
		level:    0,
		size:     0,
		header:   newHeader[K, V](DefaultMaxLevel),
		lessThan: func(k1, k2 K) bool { return cmp.Compare[K](k1, k2) == -1 },
	}
	if items != nil && len(items) > 0 {
		sl.SetAll(items)
	}
	return sl
}

// NewCustomSkipList initializes a skip list using a custom key type, which means there must be
// a function that defines a linear ordering of keys, i.e. for two keys X & Y the function must
// define how X is less than Y. Optionally include items with which to initialize the list. Uses
// default max level of 32.
func NewCustomSkipList[K, V any](lessThan func(K, K) bool, items ...Pair[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel: DefaultMaxLevel - 1,
		level:    0,
		size:     0,
		header:   newHeader[K, V](DefaultMaxLevel),
		lessThan: lessThan,
	}
	if items != nil && len(items) > 0 {
		sl.SetAll(items)
	}
	return sl
}

// Len returns the number of elements in the skip list.
func (sl *SkipList[K, V]) Len() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.size
}

// IsEmpty returns true if the skip list has no elements.
func (sl *SkipList[K, V]) IsEmpty() bool {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.size == 0
}

// MaxLevel returns the maximum number of levels any node in the skip list can be on.
func (sl *SkipList[K, V]) MaxLevel() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.maxLevel + 1
}

// First returns the first element of the skip list. This is the element with the minimum key.
func (sl *SkipList[K, V]) First() *Pair[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	if sl.header.forward[0] != nil {
		return sl.header.forward[0].Pair()
	}
	return nil
}

// Last returns the last element of the skip list. This is the element with the maximum key.
func (sl *SkipList[K, V]) Last() *Pair[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	if sl.max == nil {
		return nil
	}
	return sl.max.Pair()
}

// SetMaxLevel sets the max level of the skip list, up to 64. Inputs greater than 64 are clamped
// to 64. If the new max level is less than the level of the highest node in the list,
// the new max level will instead be that node's level.
func (sl *SkipList[K, V]) SetMaxLevel(newMaxLevel int) {
	sl.rw.RLock()

	if newMaxLevel < 0 {
		newMaxLevel = 0
	}
	if newMaxLevel > AbsoluteMaxLevel {
		newMaxLevel = AbsoluteMaxLevel
		log.Printf("Warning: maximum level clamped to %d\n", AbsoluteMaxLevel)
	}
	if newMaxLevel < sl.level {
		newMaxLevel = sl.level
	}
	for i := sl.maxLevel + 1; i < newMaxLevel; i++ {
		sl.header.forward = append(sl.header.forward, nil)
	}
	sl.maxLevel = newMaxLevel

	sl.rw.RUnlock()
}

// Set sets a key to a value in the skip list. Returns true if this pair was newly inserted, or false
// if this was an update.
func (sl *SkipList[K, V]) Set(key K, val V) bool {
	sl.rw.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.rw.RUnlock()

	sl.rw.Lock()
	defer sl.rw.Unlock()

	if x != nil && !sl.lessThan(key, x.key) {
		x.val = val
		return false
	}

	lvl := sl.randomLevel()
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
	x.backward = update[0]

	if sl.max == nil || sl.lessThan(sl.max.key, x.key) {
		sl.max = x
	}

	sl.size++
	return true
}

// SetAll inserts each key-value pair in an array of pairs into the skip list.
func (sl *SkipList[K, V]) SetAll(items []Pair[K, V]) {
	sl.rw.Lock()
	for _, item := range items {
		sl.set(item.key, item.val)
	}
	sl.rw.Unlock()
}

// Delete removes the element with given key from the skip list. Returns the deleted value if it
// existed and a bool indicating if it did.
func (sl *SkipList[K, V]) Delete(key K) (V, bool) {
	sl.rw.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.rw.RUnlock()

	sl.rw.Lock()
	defer sl.rw.Unlock()

	var val V
	if x != nil && !sl.lessThan(key, x.key) {
		if x.forward[0] == nil {
			sl.max = update[0]
		}
		if sl.max.isHeader {
			sl.max = nil
		}
		for i := 0; i <= sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		if x.forward[0] != nil {
			x.forward[0].backward = update[0]
		}
		val = x.val
		x = nil
		sl.size--
		for i := sl.level; i > 0 && sl.header.forward[sl.level] == nil; i-- {
			sl.level -= 1
		}
		return val, true
	}
	return val, false
}

// DeleteAll elements with the given keys.
func (sl *SkipList[K, V]) DeleteAll(keys []K) {
	sl.rw.Lock()
	for _, key := range keys {
		sl.delete(key)
	}
	sl.rw.Unlock()
}

// Get returns the value associated with the key if the key exists and a bool indicating if it does.
func (sl *SkipList[K, V]) Get(key K) (V, bool) {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	_, x := sl.searchNode(key)
	x = x.forward[0]
	var val V
	if x != nil && !sl.lessThan(key, x.key) {
		val = x.val
		return val, true
	}
	return val, false
}

// Range returns a bidirectional iterator beginning at the first node with key greater than or
// equal to start (inclusive) to the node with key end (exclusive), or nil if the list is empty.
func (sl *SkipList[K, V]) Range(start, end K) Iterator[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	update, startNode := sl.searchNode(start)
	startNode = startNode.forward[0]
	if startNode != nil && !sl.lessThan(startNode.key, start) {
		return sl.iterator(update[0], &end)
	}
	return nil
}

// Iterator returns a bidirectional iterator starting from the first node of the skip list,
// or nil if the list is empty.
func (sl *SkipList[K, V]) Iterator() Iterator[K, V] {
	return sl.iterator(sl.header, nil)
}

// IteratorFromEnd returns a bidirectional iterator starting from the last node of the skip list, or nil if
// the list is empty.
func (sl *SkipList[K, V]) IteratorFromEnd() Iterator[K, V] {
	var k K
	var v V
	dummy := newNode(1, k, v)
	sl.rw.RLock()
	dummy.backward = sl.max
	sl.rw.RUnlock()
	return sl.iterator(dummy, nil)
}

// IteratorFrom returns a bidirectional iterator starting from the first node with key equal to
// or greater than start, or nil if the list is empty.
func (sl *SkipList[K, V]) IteratorFrom(start K) Iterator[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	update, startNode := sl.searchNode(start)
	startNode = startNode.forward[0]
	if startNode != nil && !sl.lessThan(startNode.key, start) {
		return sl.iterator(update[0], nil)
	}
	return nil
}

// Clear resets the state of the skip list, removing all elements from the skip list.
func (sl *SkipList[K, V]) Clear() {
	sl.rw.Lock()

	sl.size = 0
	sl.level = 0
	sl.max = nil
	sl.header = newHeader[K, V](sl.maxLevel)

	sl.rw.Unlock()
}

// String returns a string representing a visualization of the skip list.
func (sl *SkipList[K, V]) String() string {
	sl.rw.RLock()

	mat := make([][]string, sl.level+1)
	for i := range mat {
		mat[i] = make([]string, sl.size)
	}

	rowHasNode := make([]bool, sl.level+1)

	for column, node := 0, sl.header.forward[0]; node != nil; column, node = column+1, node.forward[0] {
		for row, lvl := sl.level, 0; row >= 0 && lvl <= node.Level(); row, lvl = row-1, lvl+1 {
			mat[row][column] = node.String()
			rowHasNode[row] = true
		}
	}

	bldr := strings.Builder{}
	for level, row := range mat {
		if !rowHasNode[level] {
			continue
		}
		bldr.WriteString("-INF ")
		for column, str := range row {
			if str != "" {
				bldr.WriteString(str)
				bldr.WriteString(" ")
			} else {
				bldr.WriteString(strings.Repeat("-", len(mat[sl.level][column])))
				if column+1 < len(row) && row[column+1] != "" {
					bldr.WriteString(" ")
				} else {
					bldr.WriteString("-")
				}
			}
		}
		bldr.WriteString(" +INF")
		if level != sl.level {
			bldr.WriteString("\n")
		}
	}

	sl.rw.RUnlock()
	return bldr.String()
}

// Merge returns a new skip list with the elements from both lists. For any keys that are
// in both of the lists, the result will use the value from the second list.
// The maxLevel of the result will be the greater maxLevel of the inputs.
func Merge[K, V any](sl1, sl2 *SkipList[K, V]) *SkipList[K, V] {
	sl1.rw.Lock()
	sl2.rw.Lock()

	newMaxLevel := sl1.maxLevel
	if sl2.maxLevel > newMaxLevel {
		newMaxLevel = sl2.maxLevel
	}

	newMax := sl1.max
	if sl1.lessThan(sl1.max.key, sl2.max.key) {
		newMax = sl2.max
	}

	p1, p2 := sl1.header.forward[0], sl2.header.forward[0]

	newHead := newHeader[K, V](newMaxLevel)
	previous := make([]*slNode[K, V], newMaxLevel)
	for i := 0; i < newMaxLevel; i++ {
		previous[i] = newHead
	}

	newLevel, newSize := 0, 0

	for p1 != nil && p2 != nil {
		k1, k2 := p1.key, p2.key

		level := randomLevel(newMaxLevel)
		if level > newLevel {
			newLevel = level
		}

		var node *slNode[K, V]
		if sl1.lessThan(k1, k2) {
			newSize++
			node = newNode[K, V](level, k1, p1.val)
			p1 = p1.forward[0]
		} else if sl1.lessThan(k2, k1) {
			newSize++
			node = newNode[K, V](level, k2, p2.val)
			p2 = p2.forward[0]
		} else {
			p1 = p1.forward[0]
			continue
		}

		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
	}

	for p1 != nil {
		level := randomLevel(newMaxLevel)
		node := newNode[K, V](level, p1.key, p1.val)
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
		newSize++
		p1 = p1.forward[0]
	}

	for p2 != nil {
		level := randomLevel(newMaxLevel)
		node := newNode[K, V](level, p2.key, p2.val)
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
		newSize++
		p2 = p2.forward[0]
	}

	sl1.rw.Unlock()
	sl2.rw.Unlock()

	return &SkipList[K, V]{
		maxLevel: newMaxLevel,
		level:    newLevel,
		size:     newSize,
		lessThan: sl1.lessThan,
		header:   newHead,
		max:      newMax,
	}
}

// randomLevel returns highest level to which a node will be promoted, up to maxLevel.
func randomLevel(maxLevel int) int {
	return bits.TrailingZeros64(uint64(rand.Int63()) & ((1 << maxLevel) - 1))
}

// randomLevel returns the highest level a node will be promoted on insertion.
func (sl *SkipList[K, V]) randomLevel() int {
	return randomLevel(sl.maxLevel - 1)
}

// searchNode returns the node with the given key and an array containing the last
// node that comes before the target node at each level of the list.
func (sl *SkipList[K, V]) searchNode(searchKey K) ([]*slNode[K, V], *slNode[K, V]) {
	previous := make([]*slNode[K, V], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.lessThan(x.forward[i].key, searchKey) {
			x = x.forward[i]
		}
		previous[i] = x
	}
	return previous, x
}

// Inserts a key-value pair but doesn't use locks; this is used for the InsertAll() method
// to acquire a single lock for the bulk insertion
func (sl *SkipList[K, V]) set(key K, val V) {
	update, x := sl.searchNode(key)
	x = x.forward[0]
	if x != nil && !sl.lessThan(key, x.key) {
		x.val = val
	} else {
		lvl := sl.randomLevel()
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
		x.backward = update[0]

		if sl.max == nil || sl.lessThan(sl.max.key, x.key) {
			sl.max = x
		}

		sl.size++
	}
}

// Deletes a key-value pair but doesn't use locks; this is used for the DeleteAll()
// method to acquire a single lock for the bulk deletion
func (sl *SkipList[K, V]) delete(key K) {
	update, x := sl.searchNode(key)
	x = x.forward[0]
	if x != nil && !sl.lessThan(key, x.key) {
		if x.forward[0] == nil {
			sl.max = update[0]
		}
		if sl.max.isHeader {
			sl.max = nil
		}
		for i := 0; i <= sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		if x.forward[0] != nil {
			x.forward[0].backward = update[0]
		}
		x = nil
		sl.size--
		for i := sl.level; i > 0 && sl.header.forward[sl.level] == nil; i-- {
			sl.level -= 1
		}
	}
}

// iterator returns an Iterator beginning at the given node and ending at node with the given endKey (exclusive).
// If endKey is nil, the iterator goes until the end of the list. If start is nil, this would suggest the list
// is empty, so it returns nil.
func (sl *SkipList[K, V]) iterator(start *slNode[K, V], endKey *K) Iterator[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	if start == nil {
		return nil
	}

	return &iter[K, V]{
		lessThan:    sl.lessThan,
		curr:        start,
		rangeEndKey: endKey,
	}
}
