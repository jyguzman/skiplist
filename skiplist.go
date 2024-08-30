package skiplist

import (
	"cmp"
	"math/bits"
	"math/rand"
	"strings"
	"sync"
)

type SkipList[K, V any] struct {
	rw          sync.RWMutex
	maxLevel    int             // the maximum number of levels a node can appear on
	level       int             // the current highest level
	size        int             // the current number of elements
	compareFunc func(K, K) int  // function used to compare keys
	header      *SLNode[K, V]   // the header node
	min         *SLItem[K, V]   // the element with the minimum key
	max         *SLItem[K, V]   // the element with the maximum key
	tombstones  []*SLNode[K, V] // the nodes of elements that had been marked deleted
}

// NewSkipList initializes a skip list using a cmp.Ordered key type with a given maximum
// number of levels from 1 to maxLevel. Optionally include items with which to initialize the list.
func NewSkipList[K cmp.Ordered, V any](maxLevel int, items ...SLItem[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: cmp.Compare[K],
	}
	if items != nil && len(items) > 0 {
		sl.InsertAll(items)
	}
	return sl
}

// NewCustomKeySkipList initializes a skip list using a user-defined custom key type that must implement Comparable
// and with a given maximum number of levels from 1 to maxLevel.
// Use this when you have a key type that isn't a cmp.Ordered type. Your key type must implement Comparable.
// Optionally include items with which to initialize the list.
func NewCustomKeySkipList[K Comparable, V any](maxLevel int, items ...SLItem[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: Compare[K],
	}
	if items != nil && len(items) > 0 {
		sl.InsertAll(items)
	}
	return sl
}

// Size returns the number of elements in the skip list
func (sl *SkipList[K, V]) Size() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.size
}

// IsEmpty returns true if the skip list has no elements
func (sl *SkipList[K, V]) IsEmpty() bool {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.Size() == 0
}

// MaxLevel returns the maximum number of levels any node in the skip list can be on.
func (sl *SkipList[K, V]) MaxLevel() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.maxLevel
}

// Level returns the current highest level of any node in the list.
func (sl *SkipList[K, V]) Level() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.level
}

// Min returns the element with the minimum key.
func (sl *SkipList[K, V]) Min() *SLItem[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.min
}

// Max returns the element with the maximum key.
func (sl *SkipList[K, V]) Max() *SLItem[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.max
}

// Insert adds a key-value pair to the skip list.
func (sl *SkipList[K, V]) Insert(key K, val V) {
	sl.rw.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.rw.RUnlock()

	sl.rw.Lock()
	if x != nil && sl.equal(x.key, key) {
		x.val = val
		x.markedDeleted = false
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

		if sl.IsEmpty() {
			sl.min, sl.max = x.Item(), x.Item()
		}
		if sl.greater(x.key, sl.max.Key) {
			sl.max = x.Item()
		}
		if sl.less(x.key, sl.min.Key) {
			sl.min = x.Item()
		}

		sl.size++
	}
	sl.rw.Unlock()
}

// InsertAll bulk inserts each element in an array of key-value pairs.
func (sl *SkipList[K, V]) InsertAll(items []SLItem[K, V]) {
	sl.rw.Lock()
	for _, item := range items {
		sl.insert(item.Key, item.Val)
	}
	sl.rw.Unlock()
}

// Delete removes the element with given key from the skip list.
func (sl *SkipList[K, V]) Delete(key K) {
	sl.rw.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.rw.RUnlock()

	sl.rw.Lock()
	if x != nil && sl.equal(x.key, key) {
		if sl.equal(x.key, sl.max.Key) && update[0] != nil {
			sl.max = update[0].Item()
		}
		if sl.equal(x.key, sl.min.Key) && update[0].forward[0] != nil {
			sl.min = update[0].forward[0].Item()
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
	sl.rw.Unlock()
}

// LazyDelete marks a key as deleted but does not actually remove the element. It is treated as
// deleted, i.e. searches for this key will return nil, and it will be skipped in queries.
// If you are not using the skip list in a situation where lazy deletion provides better efficiency
// or consistency, such as in an LSM engine, you should prefer Delete instead.
func (sl *SkipList[K, V]) LazyDelete(key K) {
	sl.rw.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.rw.RUnlock()

	sl.rw.Lock()
	if x != nil {
		x.markedDeleted = true
		sl.size--
		if sl.size <= 1 {
			sl.max, sl.min = nil, nil
		} else if sl.equal(x.key, sl.min.Key) {
			if update[0].isHeader {
				sl.min = x.forward[0].Item()
			} else {
				sl.min = update[0].Item()
			}
		}
		if sl.equal(x.key, sl.max.Key) {
			if x.forward[0] != nil {
				sl.max = x.forward[0].Item()
			} else {
				sl.max = update[0].Item()
			}
		}
	}
	sl.rw.Unlock()
}

// DeleteAll elements with the given keys.
func (sl *SkipList[K, V]) DeleteAll(keys []K) {
	sl.rw.Lock()
	for _, key := range keys {
		sl.delete(key)
	}
	sl.rw.Unlock()
}

// Search returns a value given by the key if the key exists and a bool indicating if it does.
func (sl *SkipList[K, V]) Search(key K) (V, bool) {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	_, x := sl.searchNode(key)
	x = x.forward[0]
	var val V
	if x != nil && sl.equal(x.key, key) && !x.markedDeleted {
		val = x.val
		return val, true
	}
	return val, false
}

// Range returns a sorted list of elements from a minimum key to a maximum key (inclusive).
func (sl *SkipList[K, V]) Range(start, end K) []SLItem[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	_, startNode := sl.searchNode(start)
	startNode = startNode.forward[0]
	if startNode != nil && sl.geq(startNode.key, start) {
		return sl.iterator(startNode).UpTo(end)
	}
	return []SLItem[K, V]{}
}

// Combine returns a new skip list with the elements from both lists. For any keys that are
// in both of the lists, the result will use the value from the second list.
// The maxLevel of the result will be the greater maxLevel of the inputs.
// Tombstones will be combined. The min of the result will be the  smaller min
// of the inputs, and the max will be the greater max from the inputs.
func Combine[K, V any](sl1, sl2 *SkipList[K, V]) *SkipList[K, V] {
	sl1.rw.Lock()
	sl2.rw.Lock()

	newMaxLevel := sl1.maxLevel
	if sl2.maxLevel > newMaxLevel {
		newMaxLevel = sl2.maxLevel
	}

	p1, p2 := sl1.header.forward[0], sl2.header.forward[0]
	var tombstones []*SLNode[K, V]

	newHead := newHeader[K, V](newMaxLevel)
	previous := make([]*SLNode[K, V], newMaxLevel)
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
		var node *SLNode[K, V]
		if sl1.less(k1, k2) {
			if p1.markedDeleted {
				tombstones = append(tombstones, p1)
			} else {
				newSize++
			}
			node = newNode[K, V](level, k1, p1.val)
			p1 = p1.forward[0]
		} else {
			if p2.markedDeleted {
				tombstones = append(tombstones, p2)
			} else {
				newSize++
			}
			node = newNode[K, V](level, k2, p2.val)
			p2 = p2.forward[0]
			if sl1.equal(k1, k2) {
				p1 = p1.forward[0]
			}
		}
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
	}

	for p1 != nil {
		level := randomLevel(newMaxLevel)
		if p1.markedDeleted {
			tombstones = append(tombstones, p1)
		} else {
			newSize++
		}
		node := newNode[K, V](level, p1.key, p1.val)
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
		p1 = p1.forward[0]
	}

	for p2 != nil {
		level := randomLevel(newMaxLevel)
		if p2.markedDeleted {
			tombstones = append(tombstones, p2)
		} else {
			newSize++
		}
		node := newNode[K, V](level, p2.key, p2.val)
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
		p2 = p2.forward[0]
	}

	sl1.rw.Unlock()
	sl2.rw.Unlock()

	sl1.rw.RLock()
	sl2.rw.RLock()

	newMin := sl1.min
	if sl1.less(sl2.min.Key, sl1.min.Key) {
		newMin = sl2.min
	}

	newMax := sl1.max
	if sl1.greater(sl2.max.Key, sl1.max.Key) {
		newMin = sl2.max
	}

	sl1.rw.RUnlock()
	sl2.rw.RUnlock()

	return &SkipList[K, V]{
		maxLevel:    newMaxLevel,
		level:       newLevel,
		size:        newSize,
		compareFunc: sl1.compareFunc,
		header:      newHead,
		min:         newMin,
		max:         newMax,
		tombstones:  tombstones,
	}
}

// Copy returns a new skip list containing the same elements as the original.
func (sl *SkipList[K, V]) Copy() *SkipList[K, V] {
	newHead := newHeader[K, V](sl.maxLevel)
	previous := make([]*SLNode[K, V], sl.maxLevel)
	for i := 0; i < sl.maxLevel; i++ {
		previous[i] = newHead
	}

	sl.rw.RLock()
	newLvl, p := sl.level, sl.header.forward[0]
	for p != nil {
		level := sl.randomLevel()
		if level > newLvl {
			newLvl = level
		}
		node := newNode(level, p.key, p.val)
		for i := 0; i <= level; i++ {
			node.forward[i] = previous[i].forward[i]
			previous[i].forward[i] = node
			previous[i] = node
		}
		p = p.forward[0]
	}
	sl.rw.RUnlock()

	return &SkipList[K, V]{
		maxLevel:   sl.maxLevel,
		level:      newLvl,
		size:       sl.size,
		header:     newHead,
		min:        sl.min,
		max:        sl.max,
		tombstones: sl.tombstones,
	}
}

// Iterator returns a snapshot iterator over the skip list.
func (sl *SkipList[K, V]) Iterator() *Iterator[K, V] {
	return sl.iterator(sl.header.forward[0])
}

// ToArray returns a sorted array of all elements of the skip list.
func (sl *SkipList[K, V]) ToArray() []SLItem[K, V] {
	return sl.Iterator().All()
}

// Clean officially removes the nodes marked for deletion with LazyDelete
func (sl *SkipList[K, V]) Clean() {
	sl.rw.Lock()
	for _, t := range sl.tombstones {
		// there may have been an insert for this key, so check to see it's still marked deleted
		if t.markedDeleted {
			sl.delete(t.key)
		}
	}
	sl.tombstones = nil
	sl.rw.Unlock()
}

// Clear removes all elements from the skip list
func (sl *SkipList[K, V]) Clear() {
	sl.rw.Lock()

	sl.size = 0
	sl.level = 0
	sl.tombstones = nil
	sl.max = nil
	sl.min = nil
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
		for node != nil && node.markedDeleted {
			node = node.forward[0]
		}
		if node == nil {
			break
		}
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

// randomLevel returns highest level to which a node will be promoted, up to maxLevel.
func randomLevel(maxLevel int) int {
	return bits.TrailingZeros64(uint64(rand.Int63()) & ((1 << maxLevel) - 1))
}

// randomLevel returns the highest level a node will be promoted on insertion.
func (sl *SkipList[K, V]) randomLevel() int {
	return randomLevel(sl.maxLevel - 1)
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

// searchNode returns the node with the given key and an array containing the last
// node that comes before the target node at each level of the list.
func (sl *SkipList[K, V]) searchNode(searchKey K) ([]*SLNode[K, V], *SLNode[K, V]) {
	previous := make([]*SLNode[K, V], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i] != nil && sl.less(x.forward[i].key, searchKey) {
			x = x.forward[i]
		}
		previous[i] = x
	}
	return previous, x
}

// Inserts a key-value pair but doesn't use locks; this is used for the InsertAll() method
// to acquire a single lock for the bulk insertion
func (sl *SkipList[K, V]) insert(key K, val V) {
	update, x := sl.searchNode(key)
	x = x.forward[0]
	if x != nil && sl.equal(x.key, key) {
		x.val = val
		x.markedDeleted = false
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

		if sl.min == nil || sl.max == nil {
			sl.max = x.Item()
			sl.min = x.Item()
		}
		if sl.greater(x.key, sl.max.Key) {
			sl.max = x.Item()
		}
		if sl.less(x.key, sl.min.Key) {
			sl.min = x.Item()
		}

		sl.size++
	}
}

// Deletes a key-value pair but doesn't use locks; this is used for the DeleteAll()
// method to acquire a single lock for the bulk deletion
func (sl *SkipList[K, V]) delete(key K) {
	update, x := sl.searchNode(key)
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
}

// iterator returns an iterator beginning at the given node
func (sl *SkipList[K, V]) iterator(node *SLNode[K, V]) *Iterator[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return &Iterator[K, V]{compareFunc: sl.compareFunc, curr: node}
}
