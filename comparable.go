package main

type Comparable interface {
	Compare(other Comparable) int
}

type Int int
type Float float64
type String string

func (ik Int) Compare(other Comparable) int {
	otherInt := other.(Int)
	if ik == otherInt {
		return 0
	}
	if ik < otherInt {
		return -1
	}
	return 1
}

func (fk Float) Compare(other Comparable) int {
	otherFloat := other.(Float)
	if fk == otherFloat {
		return 0
	}
	if fk < otherFloat {
		return -1
	}
	return 1
}

func (sk String) Compare(other Comparable) int {
	otherString := other.(String)
	if sk == otherString {
		return 0
	}
	if sk < otherString {
		return -1
	}
	return 1
}
