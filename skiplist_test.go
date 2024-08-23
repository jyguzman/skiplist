package skiplist

import (
	"fmt"
	"slices"
	"testing"
)

func TestSkipListInsert(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	sl.Insert(2, "hello, world")
	sl.Insert(0, "bar")
	sl.Insert(-5, "foo")

	want := []SLItem[int, string]{
		{-5, "foo"},
		{0, "bar"},
		{2, "hello, world"},
	}
	var res []SLItem[int, string]
	h := sl.header
	for h.forward[0] != nil {
		res = append(res, *h.forward[0].Item())
		h = h.forward[0]
	}
	if !slices.Equal(res, want) {
		t.Errorf("insert: want: %v, got: %v", want, res)
	}
	if sl.size != len(res) {
		t.Errorf("insert: want: %v, got: %v", len(res), sl.size)
	}
}

func TestSkipList_InsertExistingKey(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	want := "bye, world"
	sl.Insert(2, "hello, world")
	sl.Insert(2, want)

	val, ok := sl.Search(2)
	if !ok {
		t.Errorf("test insert exisiting: search fail")
	}

	if val != want {
		t.Errorf("test insert exisiting: got %v, want %v", val, want)
	}
}

func TestSkipList_Search(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	items := []SLItem[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	for _, item := range items {
		sl.Insert(item.Key, item.Val)
	}

	for _, item := range items {
		val, ok := sl.Search(item.Key)
		if !ok {
			t.Errorf("search: key %d not found", item.Key)
		}
		if val != item.Val {
			t.Errorf("search: want value: %v, got value: %v", item.Val, val)
		}
	}

	_, ok := sl.Search(-10)
	if ok {
		t.Errorf("search: uninserted key %d found", -10)
	}
}

func TestSkipList_Delete(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	items := []SLItem[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	for _, item := range items {
		sl.Insert(item.Key, item.Val)
	}

	sl.Delete(-5)
	sl.Delete(2)

	_, ok := sl.Search(-5)
	if ok {
		t.Errorf("testing delete: deleted key %d found", -5)
	}

	_, ok = sl.Search(2)
	if ok {
		t.Errorf("testing delete: deleted key %d found", 2)
	}

	if sl.size != len(items)-2 {
		t.Errorf("delete: want: %v, got: %v", len(items), sl.size)
	}

	sl.Delete(-2)
	if sl.size != len(items)-2 {
		t.Errorf("deleting deleted affected size: want: %v, got: %v", len(items), sl.size)
	}
}

func TestSkipList_LazyDelete(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	items := []SLItem[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	for _, item := range items {
		sl.Insert(item.Key, item.Val)
	}

	sl.LazyDelete(-5)
	sl.LazyDelete(2)

	_, ok := sl.Search(-5)
	if ok {
		t.Errorf("testing lazy delete: deleted key %d found", -5)
	}

	_, ok = sl.Search(2)
	if ok {
		t.Errorf("testing lazy delete: deleted key %d found", 2)
	}

	if sl.size != len(items)-2 {
		t.Errorf("lazy delete: want: %v, got: %v", len(items), sl.size)
	}

	sl.Delete(-2)
	if sl.size != len(items)-2 {
		t.Errorf("deleting deleted affected size: want: %v, got: %v", len(items), sl.size)
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

	res := sl.Range(5, 40)
	fmt.Println(res)

}

func TestSkipList_Min(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	res := sl.Min()
	if res != nil {
		t.Error("Min on empty skip list failed")
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

	sl.Delete(1)
	res = sl.Min()
	if res.Key != 5 {
		t.Error("Min after deleting previous min failed")
	}
}

func TestSkipList_Max(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	res := sl.Max()
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

	res = sl.Max()
	if res.Key != 50 {
		t.Error("Max failed")
	}

	sl.Delete(50)
	res = sl.Max()
	if res.Key != 40 {
		t.Error("Max after deleting previous max failed")
	}
}

func TestSkipList_ToArray(t *testing.T) {
	sl := NewOrderedKeySkipList[int, string](16, 0.5)

	items := []SLItem[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	sl.InsertAll(items)

	want := []SLItem[int, string]{
		{-5, "beefcafe"}, {0, "foo"},
		{2, "bar"}, {5, "hello, world"},
		{10, "dijkstra"},
	}

	res := sl.ToArray()
	if !slices.Equal(res, want) {
		t.Errorf("want %v, got %v", want, res)
	}
}

func TestSkipList_Combine(t *testing.T) {
	sl1 := NewOrderedKeySkipList[int, string](16, 0.5)
	sl2 := NewOrderedKeySkipList[int, string](16, 0.5)

	items1 := []SLItem[int, string]{
		{6, "hello, world"},
		{4, "bar"},
		{2, "foo"},
		{0, "dijkstra"},
		{1, "bing"},
		{7, "hello, world"},
		{5, "bar"},
		{3, "foo"},
	}

	items2 := []SLItem[int, string]{
		{1, "bong"},
		{2, "boom"},
	}

	sl1.InsertAll(items1)
	sl2.InsertAll(items2)

	res := Combine(sl1, sl2)
	fmt.Println(res)
}

func Test_Example(t *testing.T) {

}
