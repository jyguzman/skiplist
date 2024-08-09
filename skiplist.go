package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

type SLItem[K Comparable] struct {
	Key K
	Val any
}

type SLNode[K Comparable] struct {
	Key     K
	Val     any
	forward []*SLNode[K]
}

func NewNode[K Comparable](level int, key K, val any) *SLNode[K] {
	return &SLNode[K]{key, val, make([]*SLNode[K], level+1)}
}

func (sn SLNode[K]) String() string {
	return fmt.Sprintf("{key: %v, val: %v}", sn.Key, sn.Val)
}

func (sn SLNode[K]) Item() SLItem[K] {
	return SLItem[K]{sn.Key, sn.Val}
}

type SkipList[K Comparable] struct {
	m        sync.RWMutex
	MaxLevel int
	Level    int
	P        float64
	Size     int
	header   *SLNode[K]
}

func (sl *SkipList[K]) randomLevel() int {
	level := 0
	for i := 0; i < sl.MaxLevel && rand.Float64() < sl.P; i++ {
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
	header := NewNode[String](maxLevel, "", nil)
	NIL := NewNode[String](0, "\xff\xff\xff\xff", nil)
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = NIL
	}

	return &SkipList[String]{
		MaxLevel: maxLevel - 1,
		Level:    0,
		P:        p,
		Size:     0,
		header:   header,
	}
}

func NewIntSkipList(maxLevel int, p float64) *SkipList[Int] {
	header := NewNode[Int](maxLevel, -math.MaxInt64, nil)
	NIL := NewNode[Int](0, math.MaxInt64, nil)
	for i := 0; i < maxLevel; i++ {
		header.forward[i] = NIL
	}

	return &SkipList[Int]{
		MaxLevel: maxLevel - 1,
		Level:    0,
		P:        p,
		Size:     0,
		header:   header,
	}
}

func (sl *SkipList[K]) Insert(searchKey K, val any) {
	sl.m.Lock()
	update := make([]*SLNode[K], sl.MaxLevel)
	x := sl.header
	for i := sl.Level; i >= 0; i-- {
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
		if lvl > sl.Level {
			for i := sl.Level + 1; i <= lvl; i++ {
				update[i] = sl.header
			}
			sl.Level = lvl
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

	update, x := make([]*SLNode[K], sl.MaxLevel), sl.header
	for i := sl.Level; i >= 0; i-- {
		for sl.less(x.forward[i].Key, searchKey) {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if sl.equal(x.Key, searchKey) {
		for i := 0; i <= sl.Level; i++ {
			if update[i].forward[i] != x {
				break
			}
			update[i].forward[i] = x.forward[i]
		}
		x = nil
		for i := sl.Level; i > 0 && sl.header.forward[sl.Level] == nil; i-- {
			sl.Level = sl.Level - 1
		}
	}

	sl.m.Unlock()
}

func (sl *SkipList[K]) Search(searchKey K) any {
	sl.m.RLock()
	defer sl.m.RUnlock()

	x := sl.header
	for i := sl.Level; i >= 0; i-- {
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
	for i := sl.Level; i >= 0; i-- {
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
