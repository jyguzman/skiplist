package main

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
)

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

type SkipList[K Comparable] struct {
	m        sync.RWMutex
	maxLevel int
	level    int
	p        float64
	size     int
	header   *SLNode[K]
}

func (sl *SkipList[K]) randomLevel() int {
	level := 0
	for i := 0; i < sl.maxLevel && rand.Float64() < sl.p; i++ {
		level++
	}
	return level
}

func NewStringSkipList(maxLevel int, p float64) *SkipList[String] {
	header := NewNode[String](maxLevel, "", nil)
	NIL := NewNode[String](0, "\xff\xff\xff\xff", nil)
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

func NewIntSkipList(maxLevel int, p float64) *SkipList[Int] {
	header := NewNode[Int](maxLevel, -math.MaxInt64, nil)
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
	update := make([]*SLNode[K], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i].Key.CompareTo(searchKey) == -1 {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if x.Key.CompareTo(searchKey) == 0 {
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
}

func (sl *SkipList[K]) Delete(searchKey K) {
	update := make([]*SLNode[K], sl.maxLevel)
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i].Key.CompareTo(searchKey) == -1 {
			x = x.forward[i]
		}
		update[i] = x
	}
	x = x.forward[0]
	if x.Key.CompareTo(searchKey) == 0 {
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
}

func (sl *SkipList[K]) Search(searchKey K) any {
	x := sl.header
	for i := sl.level; i >= 0; i-- {
		for x.forward[i].Key.CompareTo(searchKey) == -1 {
			x = x.forward[i]
		}
	}
	x = x.forward[0]
	if x.Key.CompareTo(searchKey) == 0 {
		return x.Val
	}
	return nil
}

func (sl *SkipList[K]) String() string {
	fmt.Println(sl.header.forward[1], sl.header.forward[1])
	//return ""
	res := ""
	for i := sl.level; i >= 0; i-- {
		level := sl.header.forward[i]
		fmt.Println(i, level, len(level.forward))
		line := level.String()
		//fmt.Println("level:", level, level.forward)
		//for level.forward[i] != nil {
		//	line += " -> " + strings.Repeat(" ", i) + level.forward[i].String()
		//	level = level.forward[i]
		//}
		for _, node := range level.forward {
			line += " -> " + node.String()
			//level = level.forward[i]
		}
		res += line + "\n"
	}
	return res
}
