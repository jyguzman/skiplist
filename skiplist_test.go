package main

import "testing"

func TestIntSkipListInsert(t *testing.T) {
	sl := SkipListWithIntKeys(16, 0.5)

	sl.Insert(1, "hello, world")
	if sl.Search(1) != "hello, world" {
		t.Error("Insertion failed")
	}

	sl.Insert(1, "bye, world")
	if sl.Search(1) != "bye, world" {
		t.Error("Updating existing key failed")
	}

	sl.Insert(0, "beefcafe")
	sl.Insert(2, "sushibar")
	sl.Insert(3, "porkclub")
	//levelZero := sl.header.forward[0]
	//firstThree
	//for i := 0; i < len(levelZero.forward) && levelZero.forward[i] != nil {
	//
	//}
}
