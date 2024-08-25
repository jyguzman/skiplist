package skiplist

import (
	"cmp"
	"fmt"
	"math/rand"
	"strings"
	"sync"
)

type SkipList[K, V any] struct {
	rw          sync.RWMutex
	maxLevel    int             // the maximum number of levels a node can appear on
	level       int             // the current highest level
	p           float64         // the chance from 0 to 1 that a node can appear at higher levels
	size        int             // the current number of elements
	compareFunc func(K, K) int  // function used to compare keys
	header      *SLNode[K, V]   // the header node
	min         *SLItem[K, V]   // the element with the minimum key
	max         *SLItem[K, V]   // the element with the maximum key
	tombstones  []*SLNode[K, V] // the nodes of elements that had been marked deleted
}

// NewOrderedKeySkipList initializes a skip list using a cmp.Ordered key type with a given maxLevel and p.
// Optionally include items to initialize list with.
func NewOrderedKeySkipList[K cmp.Ordered, V any](maxLevel int, p float64, items ...SLItem[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: cmp.Compare[K],
	}
	if items != nil && len(items) > 0 {
		sl.InsertAll(items)
	}
	return sl
}

// NewCustomKeySkipList initializes a skip list using a custom key type that must implement Comparable
// and with a given maxLevel and p.
// Optionally include items to initialize list with.
func NewCustomKeySkipList[K Comparable, V any](maxLevel int, p float64, items ...SLItem[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: Compare[K],
	}
	if items != nil && len(items) > 0 {
		sl.InsertAll(items)
	}
	return sl
}

// NewSkipList initializes a skip list with a given maxLevel and p. Must supply a comparator function
// for the K type: cmp.Compare[K] for a primitive ordered type, or Compare[K] for a custom type.
// Optionally include items to initialize list with.
func NewSkipList[K, V any](maxLevel int, p float64, compareFunc func(K, K) int, items ...SLItem[K, V]) *SkipList[K, V] {
	sl := &SkipList[K, V]{
		maxLevel:    maxLevel - 1,
		level:       0,
		p:           p,
		size:        0,
		header:      newHeader[K, V](maxLevel),
		compareFunc: compareFunc,
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
	return sl.Size() == 0
}

// MaxLevel returns the maximum numbers of forward pointers a node can have
func (sl *SkipList[K, V]) MaxLevel() int {
	return sl.maxLevel
}

// Level returns the current highest level of the list
func (sl *SkipList[K, V]) Level() int {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.level
}

// P returns the chance that a node is inserted into a higher level
func (sl *SkipList[K, V]) P() float64 {
	return sl.p
}

// Min returns the element with the minimum key
func (sl *SkipList[K, V]) Min() *SLItem[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return sl.min
}

// Max returns the element with the maximum key
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

// InsertAll bulk inserts an array of key-value pairs
func (sl *SkipList[K, V]) InsertAll(items []SLItem[K, V]) {
	sl.rw.Lock()
	for _, item := range items {
		sl.insert(item.Key, item.Val)
	}
	sl.rw.Unlock()
}

// Delete removes a given key & value from the skip list and locks the list.
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

// DeleteAll bulk deletes an array of key-value pairs given by the keys
func (sl *SkipList[K, V]) DeleteAll(keys []K) {
	sl.rw.Lock()
	for _, key := range keys {
		sl.delete(key)
	}
	sl.rw.Unlock()
}

// Search returns a value given by the key if it exists and a bool indicating if it exists
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

// Range returns a list of elements sorted from a minimum key to a maximum key.
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

// Combine returns a new skip list with the keys from both lists. If elements from both have equal
// keys, the result will use the values from the second list. The P of the new list
// will be the average of the inputs. The level and maxLevel of the result will be the largest of inputs.
// Tombstones will be combined. The min of the result will be the
// smaller min of the inputs, and the max will be the greater max from the inputs.
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

	newP := (sl1.p + sl2.p) / 2.0
	newLevel := 0

	for p1 != nil && p2 != nil {
		k1, k2 := p1.key, p2.key
		level := randomLevel(newMaxLevel, newP)
		if level > newLevel {
			newLevel = level
		}
		var node *SLNode[K, V]
		if sl1.less(k1, k2) {
			if p1.markedDeleted {
				tombstones = append(tombstones, p1)
			}
			node = newNode[K, V](level, k1, p1.val)
			p1 = p1.forward[0]
		} else {
			if p2.markedDeleted {
				tombstones = append(tombstones, p2)
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
		level := randomLevel(newMaxLevel, newP)
		if p1.markedDeleted {
			tombstones = append(tombstones, p1)
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
		level := randomLevel(newMaxLevel, newP)
		if p2.markedDeleted {
			tombstones = append(tombstones, p2)
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
		p:           newP,
		size:        sl1.size + sl2.size,
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
		p:          sl.p,
		size:       sl.size,
		header:     newHead,
		min:        sl.min,
		max:        sl.max,
		tombstones: sl.tombstones,
	}
}

// Iterator returns a snapshot iterator over the skip list
func (sl *SkipList[K, V]) Iterator() *Iterator[K, V] {
	return sl.iterator(sl.header.forward[0])
}

// ToArray returns a sorted array of all elements of the skip list
func (sl *SkipList[K, V]) ToArray() []SLItem[K, V] {
	return sl.Iterator().All()
}

// LazyDelete marks a key as deleted but does not actually remove the element. It is treated as
// deleted, i.e. searches for this key will return nil, and it will be skipped in queries
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

func (sl *SkipList[K, V]) String() string {
	sl.rw.RLock()

	res := ""
	for i := sl.level; i >= 0; i-- {
		level := sl.header.forward[i]
		levelStr := ""
		if level != nil {
			levelStr += level.String()
			for i < len(level.forward) && level.forward[i] != nil {
				levelStr += " -> " + level.forward[i].String()
				level = level.forward[i]
			}
			levelStr = "-INF -> " + levelStr + " -> +INF"
			res += levelStr + "\n"
		}
	}

	sl.rw.RUnlock()
	return res
}

func (sl *SkipList[K, V]) StringAlt() string {
	sl.rw.RLock()

	fmt.Println("current level:", sl.level)
	mat := make([][]string, sl.level+1)
	for i := range mat {
		mat[i] = make([]string, sl.size)
	}

	for column, node := 0, sl.header.forward[0]; node != nil; column, node = column+1, node.forward[0] {
		for row, lvl := sl.level, 0; row >= 0 && lvl <= node.Level(); row, lvl = row-1, lvl+1 {
			mat[row][column] = node.String()
		}
	}

	buf := strings.Builder{}
	for _, row := range mat {
		buf.WriteString("-INF ")
		for col, str := range row {
			if str != "" {
				buf.WriteString(str)
				//buf.WriteString(" - ")
			} else {
				bottom := mat[sl.level][col]
				whitespace := strings.Repeat("-", len(bottom))
				//buf.WriteString(" ")
				buf.WriteString(whitespace) // + strings.Repeat( "-", len(bottom)) )
			}
		}
		buf.WriteString(" +INF")
		buf.WriteString("\n")
	}

	sl.rw.RUnlock()
	return buf.String()
}

// randomLevel returns the highest level a node will be assigned
func randomLevel(maxLevel int, p float64) int {
	//randBits := rand.Int63()

	level := 0
	for i := 0; i < maxLevel && rand.Float64() < p; i++ {
		level++
	}
	return level
}

// randomLevel returns the highest level a node will be assigned
func (sl *SkipList[K, V]) randomLevel() int {
	return randomLevel(sl.maxLevel-1, sl.p)
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
				fmt.Println("i:", i)
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

// Deletes a key-value pair but doesn't use locks; this is used for the DeleteAll() method to acquire a single
// lock for the bulk deletion
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

func (sl *SkipList[K, V]) iterator(node *SLNode[K, V]) *Iterator[K, V] {
	sl.rw.RLock()
	defer sl.rw.RUnlock()

	return &Iterator[K, V]{compareFunc: sl.compareFunc, curr: node}
}
