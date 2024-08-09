package main

type Int int
type Int8 int8
type Int16 int16
type Int32 int32
type Int64 int64

type Uint uint
type Uint8 uint8
type Uint16 uint16
type Uint32 uint32
type Uint64 uint64

type Float32 float32
type Float64 float64

type String string

func (i Int) Cmp(other Comparable) int {
	otherInt := other.(Int)
	if i > otherInt {
		return 1
	}
	if i < otherInt {
		return -1
	}
	return 0
}

func (i8 Int8) Cmp(other Comparable) int {
	otherInt := other.(Int8)
	if i8 > otherInt {
		return 1
	}
	if i8 < otherInt {
		return -1
	}
	return 0
}

func (i16 Int16) Cmp(other Comparable) int {
	otherInt := other.(Int16)
	if i16 > otherInt {
		return 1
	}
	if i16 < otherInt {
		return -1
	}
	return 0
}

func (i Int32) Cmp(other Comparable) int {
	otherInt := other.(Int32)
	if i > otherInt {
		return 1
	}
	if i < otherInt {
		return -1
	}
	return 0
}

func (i Int64) Cmp(other Comparable) int {
	otherInt := other.(Int64)
	if i > otherInt {
		return 1
	}
	if i < otherInt {
		return -1
	}
	return 0
}

func (ui Uint) Cmp(other Comparable) int {
	otherInt := other.(Uint)
	if ui > otherInt {
		return 1
	}
	if ui < otherInt {
		return -1
	}
	return 0
}

func (ui Uint8) Cmp(other Comparable) int {
	otherInt := other.(Uint8)
	if ui > otherInt {
		return 1
	}
	if ui < otherInt {
		return -1
	}
	return 0
}

func (ui Uint16) Cmp(other Comparable) int {
	otherInt := other.(Uint16)
	if ui > otherInt {
		return 1
	}
	if ui < otherInt {
		return -1
	}
	return 0
}

func (ui Uint32) Cmp(other Comparable) int {
	otherInt := other.(Uint32)
	if ui > otherInt {
		return 1
	}
	if ui < otherInt {
		return -1
	}
	return 0
}

func (ui Uint64) Cmp(other Comparable) int {
	otherInt := other.(Uint64)
	if ui > otherInt {
		return 1
	}
	if ui < otherInt {
		return -1
	}
	return 0
}

func (f Float32) Cmp(other Comparable) int {
	otherFloat := other.(Float32)
	if f > otherFloat {
		return 1
	}
	if f < otherFloat {
		return -1
	}
	return 0
}

func (f Float64) Cmp(other Comparable) int {
	otherFloat := other.(Float64)
	if f > otherFloat {
		return 1
	}
	if f < otherFloat {
		return -1
	}
	return 0
}

func (s String) Cmp(other Comparable) int {
	otherString := other.(String)
	if s > otherString {
		return 1
	}
	if s < otherString {
		return -1
	}
	return 0
}
