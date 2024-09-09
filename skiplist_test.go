package skiplist

import (
	"bytes"
	"fmt"
	"slices"
	"testing"
)

func TestSkipList_Set(t *testing.T) {
	sl := NewSkipList[int, string]()

	sl.Set(2, "hello, world")
	sl.Set(0, "bar")
	sl.Set(-5, "foo")

	want := []Pair[int, string]{
		{-5, "foo"},
		{0, "bar"},
		{2, "hello, world"},
	}
	var res []Pair[int, string]
	h := sl.header
	for h.forward[0] != nil {
		res = append(res, *h.forward[0].Pair())
		h = h.forward[0]
	}
	if !slices.Equal(res, want) {
		t.Errorf("insert: want: %v, got: %v", want, res)
	}
	if sl.size != len(res) {
		t.Errorf("insert: want: %v, got: %v", len(res), sl.size)
	}
}

func TestSkipList_SetExistingKey(t *testing.T) {
	sl := NewSkipList[int, string]()

	want := "bye, world"
	sl.Set(2, "hello, world")
	sl.Set(2, want)

	val, ok := sl.Get(2)
	if !ok {
		t.Errorf("test insert exisiting: search fail")
	}

	if val != want {
		t.Errorf("test insert exisiting: got %v, want %v", val, want)
	}
}

func TestSkipList_Search(t *testing.T) {
	sl := NewSkipList[int, string]()

	items := []Pair[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	for _, item := range items {
		sl.Set(item.key, item.val)
	}

	for _, item := range items {
		val, ok := sl.Get(item.key)
		if !ok {
			t.Errorf("search: key %d not found", item.key)
		}
		if val != item.val {
			t.Errorf("search: want value: %v, got value: %v", item.val, val)
		}
	}

	_, ok := sl.Get(-10)
	if ok {
		t.Errorf("search: uninserted key %d found", -10)
	}
}

func TestSkipList_Delete(t *testing.T) {
	sl := NewSkipList[int, string]()

	items := []Pair[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
		{10, "dijkstra"},
	}

	for _, item := range items {
		sl.Set(item.key, item.val)
	}

	val, ok := sl.Delete(-5)
	if !ok {
		t.Errorf("delete: fail for key -5")
	}
	if val != "beefcafe" {
		t.Errorf("delete: want value: %v, got value: %v", "beefcafe", val)
	}

	val, ok = sl.Delete(2)
	if !ok {
		t.Errorf("delete: fail for key 2")
	}
	if val != "bar" {
		t.Errorf("delete: want value: %v, got value: %v", "bar", val)
	}

	_, ok = sl.Get(-5)
	if ok {
		t.Errorf("testing delete: deleted key %d found", -5)
	}

	_, ok = sl.Get(2)
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

//func TestSkipList_Range(t *testing.T) {
//	sl := NewSkipList[int, string]()
//
//	sl.Set(10, "ten")
//	sl.Set(20, "twenty")
//	sl.Set(40, "forty")
//	sl.Set(50, "fifty")
//	sl.Set(8, "eight")
//	sl.Set(5, "five")
//	sl.Set(30, "thirty")
//	sl.Set(1, "hello, world")
//
//	res := sl.Range(5, 40)
//	fmt.Println(res)
//
//}

func TestSkipList_Min(t *testing.T) {
	sl := NewSkipList[int, string]()

	res := sl.First()
	if res != nil {
		t.Error("Min on empty skip list failed")
	}

	sl.Set(10, "ten")
	sl.Set(20, "twenty")
	sl.Set(40, "forty")
	sl.Set(50, "fifty")
	sl.Set(8, "eight")
	sl.Set(5, "five")
	sl.Set(30, "thirty")
	sl.Set(1, "hello, world")

	res = sl.First()
	if res.key != 1 {
		t.Errorf("Initial min failed, got %d but wanted %d", res.key, 1)
	}

	sl.Delete(1)

	res = sl.First()
	if res.key != 5 {
		t.Errorf("Min after deleting previous min failed, got %d but wanted %d", res.key, 5)
	}
}

func TestSkipList_Max(t *testing.T) {
	sl := NewSkipList[int, string]()

	res := sl.Last()
	if res != nil {
		t.Error("Max on empty skip list failed")
	}

	sl.Set(10, "ten")
	sl.Set(20, "twenty")
	sl.Set(40, "forty")
	sl.Set(50, "fifty")
	sl.Set(8, "eight")
	sl.Set(5, "five")
	sl.Set(30, "thirty")
	sl.Set(1, "hello, world")

	res = sl.Last()
	want := 50
	if res.key != 50 {
		t.Errorf("Initial max call failed, got %d but want %d", res.key, want)
	}

	sl.Delete(50)
	res = sl.Last()
	if res.key != 40 {
		t.Error("Max after deleting previous max failed")
	}
}

func TestSkipList_Merge(t *testing.T) {
	sl1 := NewSkipList[int, string]()
	sl2 := NewSkipList[int, string]()

	items1 := []Pair[int, string]{
		{6, "hello, world"},
		{4, "bar"},
		{2, "foo"},
		{7, "bye, world"},
		{0, "dijkstra"},
		{1, "bing"},
	}

	items2 := []Pair[int, string]{
		{-1, "negOne"},
		{2, "boom"},
		{5, "bar"},
		{1, "bong"},
		{3, "foo"},
	}

	sl1.SetAll(items1)
	sl2.SetAll(items2)

	res := Merge(sl1, sl2)
	fmt.Println(res)
	fmt.Println(res.size)
	fmt.Println(res.First(), res.max.backward)
}

func TestNewCustomKeySkipList(t *testing.T) {
	sl := NewCustomSkipList[[]byte, string](func(bs1, bs2 []byte) bool {
		return bytes.Compare(bs1, bs2) < 0
	})
	bs1 := []byte("foo")
	bs2 := []byte("bar")
	sl.Set(bs1, "hello, world")
	sl.Set(bs2, "bye, world")

	val, ok := sl.Get(bs1)
	if !ok {
		t.Errorf("key %v not found", string(bs1))
	}
	if val != "hello, world" {
		t.Errorf("wanted val %s but got %s", "hello, world", val)
	}

	val, ok = sl.Delete(bs2)
	if !ok {
		t.Errorf("key %v not found", string(bs2))
	}
	if val != "bye, world" {
		t.Errorf("wanted val %s but got %s", "bye, world", val)
	}
}

func TestSkipList_DeleteIterator(t *testing.T) {
	items := []Pair[int, string]{
		{-5, "beefcafe"},
		{0, "foo"},
		{1, "bar"},
		{2, "barTwo"},
		{4, "bing"},
		{7, "bong"},
		{8, "hello, world"},
	}

	sl := NewSkipList(items...)

	sl.Delete(4)
	sl.Delete(8)

	it := sl.IteratorFromEnd()
	for it.Prev() {
		fmt.Println(it.Key(), it.Value())
	}
}

func TestSkipList_Concurrent(t *testing.T) {
	sl := NewSkipList[int, string]()
	fmt.Println(sl)
}
