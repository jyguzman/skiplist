package skiplist

import (
	"cmp"
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
	min         *SLItem[K, V]
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
	sl.m.RLock()
	defer sl.m.RUnlock()

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
	sl.m.RLock()
	defer sl.m.RUnlock()

	return sl.level
}

// P returns the chance that a node is inserted into a higher level
func (sl *SkipList[K, V]) P() float64 {
	return sl.p
}

// Min returns the element with the minimum key
func (sl *SkipList[K, V]) Min() *SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	return sl.min
}

// Max returns the element with the maximum key
func (sl *SkipList[K, V]) Max() *SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	return sl.max
}

// Insert adds a key-value pair to the skip list.
func (sl *SkipList[K, V]) Insert(key K, val V) {
	sl.m.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.m.RUnlock()

	sl.m.Lock()
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
			sl.min = x.Item()
			sl.max = x.Item()
		}
		if sl.greater(x.key, sl.max.Key) {
			sl.max = x.Item()
		}
		if sl.less(x.key, sl.min.Key) {
			sl.min = x.Item()
		}

		sl.size++
	}
	sl.m.Unlock()
}

// InsertAll bulk inserts an array of key-value pairs
func (sl *SkipList[K, V]) InsertAll(items []SLItem[K, V]) {
	sl.m.Lock()
	for _, item := range items {
		sl.insert(item.Key, item.Val)
	}
	sl.m.Unlock()
}

// Delete removes a given key & value from the skip list and locks the list.
func (sl *SkipList[K, V]) Delete(key K) {
	sl.m.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.m.RUnlock()

	sl.m.Lock()
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

// DeleteAll bulk deletes an array of key-value pairs given by the keys
func (sl *SkipList[K, V]) DeleteAll(keys []K) {
	sl.m.Lock()
	for _, key := range keys {
		sl.delete(key)
	}
	sl.m.Unlock()
}

// Search returns a value given by the key if it exists and a bool indicating if it exists
func (sl *SkipList[K, V]) Search(key K) (V, bool) {
	sl.m.RLock()
	defer sl.m.RUnlock()

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
func (sl *SkipList[K, V]) Range(start, end K) []*SLItem[K, V] {
	sl.m.RLock()
	defer sl.m.RUnlock()

	_, startNode := sl.searchNode(start)
	startNode = startNode.forward[0]
	if startNode != nil && sl.geq(startNode.key, start) {
		return sl.iterator(startNode).UpTo(end)
	}
	return []*SLItem[K, V]{}
}

func merge[K, V any](sl1, sl2 *SkipList[K, V]) *SkipList[K, V] {
	sl1.m.Lock()
	sl2.m.Lock()

	sl1.maxLevel = sl2.maxLevel

	p1, p2 := sl1.header, sl2.header

	for p1 != nil && p2 != nil {

	}

	sl1.m.Unlock()
	sl2.m.Unlock()

	return nil
}

// Merge combines this skip list with another
func (sl *SkipList[K, V]) Merge(other *SkipList[K, V]) {
	sl.m.Lock()
	other.m.Lock()

	defer sl.m.Unlock()
	defer other.m.Unlock()

	//p1, p2 := sl.header, other.header
	//
	//for p1 != nil && p2 != nil {
	//	key1, key2 := p1.key, p2.key
	//	if sl.less(key1, key2) {
	//
	//	} else if sl.greater(key1, key2) {
	//		next := p1.forward[0]
	//		fmt.Println(next)
	//	} else {
	//		lvls1, lvls2 := p1.Level(), p2.Level()
	//		fmt.Println(lvls1, lvls2)
	//	}
	//}
	//
	//fmt.Println(p1, p2)
}

// Iterator returns a snapshot iterator over the skip list
func (sl *SkipList[K, V]) Iterator() *Iterator[K, V] {
	return sl.iterator(sl.header)
}

// ToArray returns a sorted array of all elements of the skip list
func (sl *SkipList[K, V]) ToArray() []*SLItem[K, V] {
	return sl.Iterator().All()
}

// LazyDelete marks a key as deleted but does not actually remove the element. It is treated as
// deleted, i.e. searches for this key will return nil, and it will be skipped in queries
func (sl *SkipList[K, V]) LazyDelete(key K) {
	sl.m.RLock()
	update, x := sl.searchNode(key)
	x = x.forward[0]
	sl.m.RUnlock()

	sl.m.Lock()
	if x != nil {
		x.markedDeleted = true
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
	sl.m.Unlock()
}

// Clean officially removes the lazily "deleted" elements
func (sl *SkipList[K, V]) Clean() {
	sl.m.Lock()
	for _, t := range sl.tombstones {
		// there may have been an insert for this key, so check to see it's still marked deleted
		if t.markedDeleted {
			sl.delete(t.key)
		}
	}
	sl.tombstones = nil
	sl.m.Unlock()
}

// Clear removes all elements from the skip list
func (sl *SkipList[K, V]) Clear() {
	sl.m.Lock()

	sl.size = 0
	sl.level = 0
	sl.tombstones = nil
	sl.max = nil
	sl.min = nil
	sl.header = newHeader[K, V](sl.maxLevel)

	sl.m.Unlock()
}

func (sl *SkipList[K, V]) String() string {
	sl.m.RLock()

	res := ""
	p := sl.header
	var ranks []K
	for p.forward[0] != nil {
		ranks = append(ranks, p.forward[0].key)
	}
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

func (sl *SkipList[K, V]) isMin(sn *SLNode[K, V]) bool {
	sn.lock()
	defer sn.unlock()

	return sl.equal(sn.key, sl.min.Key)
}

func (sl *SkipList[K, V]) isMax(sn *SLNode[K, V]) bool {
	sn.lock()
	defer sn.unlock()

	return sl.equal(sn.key, sl.max.Key)
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
	sl.m.RLock()
	defer sl.m.RUnlock()

	return &Iterator[K, V]{compareFunc: sl.compareFunc, curr: node}
}

func (sl *SkipList[K, V]) skipTombstones(node *SLNode[K, V]) {
	for node.forward[0] != nil && node.forward[0].markedDeleted {
		node = node.forward[0]
	}
}
