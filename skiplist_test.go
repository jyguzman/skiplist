package main

import "testing"

func TestIntSkipListInsert(t *testing.T) {
	sl := NewIntSkipList(16, 0.5)

	sl.Insert(1, "hello, world")
	if sl.Search(1) != "hello, world" {
		t.Error("Insertion failed")
	}

	sl.Insert(1, "bye, world")
	if sl.Search(1) != "bye, world" {
		t.Error("Updating existing key failed")
	}
}
