package main

type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

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
