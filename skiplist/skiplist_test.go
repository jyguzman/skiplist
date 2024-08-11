package skiplist

import (
	"slices"
	"testing"
)

func TestSkipListInsert(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(1, "hello, world")
	res, ok := sl.Search(1)
	if !ok || res != "hello, world" {
		t.Error("Insertion failed")
	}

	sl.Insert(1, "bye, world")
	res, ok = sl.Search(1)
	if !ok || res != "bye, world" {
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

func TestSkipList_Delete(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(1, "hello, world")
	sl.Delete(1)

	_, ok := sl.Search(1)
	if ok {
		t.Error("Deletion failed")
	}
}

func TestSkipList_LazyDelete(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(1, "hello, world")
	sl.LazyDelete(1)

	_, ok := sl.Search(1)
	if ok {
		t.Error("Lazy deletion failed")
	}
}

func TestSkipListRange(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(10, "ten")
	sl.Insert(20, "twenty")
	sl.Insert(40, "forty")
	sl.Insert(50, "fifty")
	sl.Insert(8, "eight")
	sl.Insert(5, "five")
	sl.Insert(30, "thirty")
	sl.Insert(1, "hello, world")

	res := sl.RangeInc(-5, 42)
	var strings []string
	for _, v := range res {
		strings = append(strings, v.Val)
	}
	inclusiveExpected := []string{"five", "eight", "ten", "forty"}

	if !slices.Equal(strings, inclusiveExpected) {
		t.Errorf("Range failed")
	}

}

func TestSkipListRank(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(10, "ten")
	sl.Insert(20, "twenty")
	sl.Insert(40, "forty")
	sl.Insert(50, "fifty")
	sl.Insert(8, "eight")
	sl.Insert(5, "five")
	sl.Insert(30, "thirty")
	sl.Insert(1, "hello, world")

	res := sl.Rank(20)
	if res != 5 {
		t.Error("Rank failed: expected 5, got ", res)
	}
}

func TestSkipListMinMax(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	res := sl.Min()
	if res != nil {
		t.Error("Min on empty skip list failed")
	}

	res = sl.Max()
	if res != nil {
		t.Error("Max on empty skip list failed")
	}

	sl.Insert(10, "ten")
	sl.Insert(20, "twenty")
	sl.Insert(40, "forty")
	sl.Insert(50, "fifty")
	sl.Insert(8, "eight")
	sl.Insert(5, "five")
	sl.Insert(30, "thirty")
	sl.Insert(1, "hello, world")

	res = sl.Min()
	if res.Key != 1 {
		t.Error("Min failed")
	}

	res = sl.Max()
	if res.Key != 50 {
		t.Error("Max failed")
	}

	sl.Delete(1)
	sl.Delete(50)

	res = sl.Min()
	if res.Key != 5 {
		t.Error("Min after deleting previous min failed")
	}

	res = sl.Max()
	if res.Key != 40 {
		t.Error("Max after deleting previous max failed")
	}
}
