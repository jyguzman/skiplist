package main

import (
	"math"
	"math/rand"
	"sync"
)

type SkipList[K Comparable] struct {
	m        sync.RWMutex
	maxLevel int
	level    int
	p        float64
	size     int
	header   *SLNode[K]
}

func (sl *SkipList[K]) Size() int {
	return sl.size
}

func (sl *SkipList[K]) MaxLevel() int {
	return sl.maxLevel
}

func (sl *SkipList[K]) Level() int {
	return sl.maxLevel
}

func (sl *SkipList[K]) P() float64 {
	return sl.p
}

func (sl *SkipList[K]) randomLevel() int {
	level := 0
	for i := 0; i < sl.maxLevel && rand.Float64() < sl.p; i++ {
		level++
	}
	return level
}

func (sl *SkipList[K]) less(x, y Comparable) bool {
	return x.CompareTo(y) == -1
}

func (sl *SkipList[K]) equal(x, y Comparable) bool {
	return x.CompareTo(y) == 0
}

func NewStringSkipList(maxLevel int, p float64) *SkipList[String] {
	header := NewNode[String](maxLevel+1, "", nil)
	NIL := NewNode[String](0, "\uffff", nil)
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = NIL
	}

	return &SkipList[String]{
		maxLevel: maxLevel,
		level:    0,
		p:        p,
		size:     0,
		header:   header,
	}
}

func SkipListWithIntKeys(maxLevel int, p float64) *SkipList[Int] {
	header := NewNode[Int](maxLevel+1, -math.MaxInt64, nil)
	NIL := NewNode[Int](0, math.MaxInt64, nil)
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = NIL
	}

	return &SkipList[Int]{
		maxLevel: maxLevel,
		level:    0,
		p:        p,
		size:     0,
		header:   header,
	}
}

func (sl *SkipList[K]) Insert(searchKey K, val any) {
	sl.m.Lock()

	update, x := make([]*SLNode[K], sl.maxLevel), sl.header
	for i := sl.level; i >= 0; i-- {
		for sl.less(x.forward[i].Key, searchKey) {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if sl.equal(x.Key, searchKey) {
		x.Val = val
	} else {
		lvl := sl.randomLevel()
		if lvl > sl.level {
			for i := sl.level + 1; i <= lvl; i++ {
				update[i] = sl.header
			}
			sl.level = lvl
		}

		x = NewNode[K](lvl, searchKey, val)
		for i := 0; i <= lvl; i++ {
			x.forward[i] = update[i].forward[i]
			update[i].forward[i] = x
		}
	}

	sl.m.Unlock()
}

func (sl *SkipList[K]) Delete(searchKey K) {
	sl.m.Lock()

	update, x := make([]*SLNode[K], sl.maxLevel), sl.header
	for i := sl.level; i >= 0; i-- {
		for sl.less(x.forward[i].Key, searchKey) {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if sl.equal(x.Key, searchKey) {
		for i := 0; i <= sl.level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		x = nil
		for i := sl.level; i > 0 && sl.header.forward[sl.level] == nil; i-- {
			sl.level = sl.level - 1
		}
	}

	sl.m.Unlock()
}

func (sl *SkipList[K]) Search(searchKey K) any {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for sl.less(x.forward[i].Key, searchKey) {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	if sl.equal(x.Key, searchKey) {
		return x.Val
	}
	return nil
}

func (sl *SkipList[K]) String() string {
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
