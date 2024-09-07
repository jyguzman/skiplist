package skiplist

import (
	"testing"
)

func TestIterator_Next(t *testing.T) {
	items := []SLItem[int, string]{
		{-5, "beefcafe"},
		{0, "foo"},
		{1, "bar"},
		{2, "bar"},
		{4, "bing"},
		{7, "bong"},
		{8, "hello, world"},
	}

	sl := NewSkipList(items...)

	it := sl.Iterator()
	i := 0
	for it.Next() {
		key, val := it.Key(), it.Value()
		if key != items[i].Key {
			t.Errorf("key mismatch, expected %v, got %v", items[i].Key, key)
		}
		if val != items[i].Val {
			t.Errorf("val mismatch, expected %v, got %v", items[i].Val, val)
		}
		i++
	}
}

func TestIterator_Prev(t *testing.T) {
	items := []SLItem[int, string]{
		{-5, "beefcafe"},
		{0, "foo"},
		{1, "bar"},
		{2, "bar"},
		{4, "bing"},
		{7, "bong"},
		{8, "hello, world"},
	}

	sl := NewSkipList(items...)

	it := sl.Iterator()
	for it.Next() {
	}

	key, val := it.Key(), it.Value()
	i := len(items) - 1
	if key != items[i].Key {
		t.Errorf("key mismatch, expected %v, got %v", items[i].Key, key)
	}
	if val != items[i].Val {
		t.Errorf("val mismatch, expected %v, got %v", items[i].Val, val)
	}
	for it.Prev() {
		i--
		key, val = it.Key(), it.Value()
		if key != items[i].Key {
			t.Errorf("key mismatch, expected %v, got %v", items[i].Key, key)
		}
		if val != items[i].Val {
			t.Errorf("val mismatch, expected %v, got %v", items[i].Val, val)
		}
	}
}

func TestIterator_Range(t *testing.T) {
	items := []SLItem[int, string]{
		{-5, "beefcafe"},
		{0, "foo"},
		{1, "bar"},
		{2, "bar"},
		{4, "bing"},
		{7, "bong"},
		{8, "hello, world"},
	}

	sl := NewSkipList(items...)

	it := sl.Range(0, 7)
	i := 1
	for it.Next() {
		key, val := it.Key(), it.Value()
		if key != items[i].Key {
			t.Errorf("key mismatch, expected %v, got %v", items[i].Key, key)
		}
		if val != items[i].Val {
			t.Errorf("val mismatch, expected %v, got %v", items[i].Val, val)
		}
		i++
	}

	ok := it.Next()
	if ok != false {
		t.Errorf("range: expected false but got %v", ok)
	}
}
