package main

import "fmt"

type SLItem[K Comparable] struct {
	Key K
	Val any
}

type SLNode[K Comparable] struct {
	Key     K
	Val     any
	forward []*SLNode[K]
}

func (sn SLNode[K]) String() string {
	return fmt.Sprintf("{key: %v, val: %v}", sn.Key, sn.Val)
}

func (sn SLNode[K]) Item() SLItem[K] {
	return SLItem[K]{sn.Key, sn.Val}
}

func NewNode[K Comparable](level int, key K, val any) *SLNode[K] {
	return &SLNode[K]{key, val, make([]*SLNode[K], level+1)}
}
