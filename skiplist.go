package main

type SLNode[K Comparable] struct {
	Key      K
	Val      any
	Pointers []*SLNode[K]
}

type SkipList[K Comparable] struct {
	maxLevel int
	size     int
	header   []*SLNode[K]
}

func NewIntSkipList() *SkipList[Int] {
	return &SkipList[Int]{}
}

func NewFloatSkipList() *SkipList[Float] {
	return &SkipList[Float]{}
}

func NewStringSkipList() *SkipList[String] {
	return &SkipList[String]{}
}

func NewSkipList[K Comparable]() *SkipList[K] {
	return &SkipList[K]{}
}

func (sl *SkipList[K]) Insert(key K, val any) {}

func (sl *SkipList[K]) Delete(key K) {}

func (sl *SkipList[K]) Search(key K) {}
