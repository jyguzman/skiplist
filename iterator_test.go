package skiplist

import (
	"slices"
	"testing"
)

func TestIterator_Next(t *testing.T) {
	sl := NewSkipList[int, string](16)

	items := []SLItem[int, string]{
		{5, "hello, world"},
		{2, "bar"},
		{0, "foo"},
		{-5, "beefcafe"},
	}

	sorted := []SLItem[int, string]{
		{-5, "beefcafe"},
		{0, "foo"},
		{2, "bar"},
		{5, "hello, world"},
	}

	for _, item := range items {
		sl.Insert(item.Key, item.Val)
	}

	it := sl.Iterator()
	for i := range sorted {
		next := it.Next()
		if next.Key != sorted[i].Key || next.Val != sorted[i].Val {
			t.Errorf("iterator next: got %v, want %v", next.Key, sorted[i].Key)
		}
	}
}

func TestIterator_All(t *testing.T) {
	sl := NewSkipList[int, string](16)

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

	want := []SLItem[int, string]{
		{-5, "beefcafe"}, {0, "foo"},
		{2, "bar"}, {5, "hello, world"},
		{10, "dijkstra"},
	}

	res := sl.Iterator().All()
	if !slices.Equal(res, want) {
		t.Errorf("All: got %v\n want %v", res, want)
	}
}

func TestIterator_All_WithoutTombstones(t *testing.T) {
	sl := NewSkipList[int, string](16)

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
	sl.LazyDelete(10)

	want := []SLItem[int, string]{
		{0, "foo"}, {5, "hello, world"},
	}

	res := sl.Iterator().All()
	if !slices.Equal(res, want) {
		t.Errorf("All: got %v\n want %v", res, want)
	}
}

func TestIterator_UpTo(t *testing.T) {
	sl := NewSkipList[int, string](16)

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

	want := []SLItem[int, string]{
		{-5, "beefcafe"}, {0, "foo"},
		{2, "bar"},
	}

	res := sl.Iterator().UpTo(2)
	if !slices.Equal(res, want) {
		t.Errorf("All: got %v\n want %v", res, want)
	}
}
