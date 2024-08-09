package main

import "math"

type Comparable interface {
	CompareTo(other Comparable) int
	Inf(int) Comparable
}

type Comp[T any] interface {
}

func (i Int) CompareTo(other Comparable) int {
	otherInt := other.(Int)
	if i > otherInt {
		return 1
	}
	if i < otherInt {
		return -1
	}
	return 0
}

func (i Int) Inf(sign int) Comparable {
	if sign >= 0 {
		return Int(math.MaxInt64)
	}
	return Int(-math.MaxInt64)
}

func (s String) CompareTo(other Comparable) int {
	otherString := other.(String)
	if s > otherString {
		return 1
	}
	if s < otherString {
		return -1
	}
	return 0
}

func (s String) Inf(sign int) Comparable {
	if sign >= 0 {
		return String("\xff\xff\xff\xff")
	}
	return String("")
}
