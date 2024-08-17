package main

import (
	"fmt"
	"github.com/jyguzman/skiplist/skiplist"
)

func main() {
	sl := skiplist.NewOrderedKeySkipList[int, string](16, 0.5)
	sl.Insert(0, "zero")
	sl.Insert(1, "one")
	sl.Insert(2, "two")
	sl.Insert(3, "three")
	sl.Insert(4, "four")
	sl.Insert(5, "five")
	sl.LazyDelete(2)
	sl.LazyDelete(5)
	//it := sl.Iterator()
	fmt.Println(sl.Range(1, 5))
	//sl.Insert(5, "five")
	//sl.Insert(10, "ten")
	//sl.Insert(-10, "negTen")
	//sl.Insert(15, "fifteen")
	//sl.Insert(20, "twenty")
	//sl.Insert(40, "forty")
	//sl.Insert(-2, "negTwo")
	//fmt.Println(sl.Max())
	//sl.LazyDelete(40)
	//fmt.Println(sl.Max())
	//items := []skiplist.SLItem[int, string]{
	//	skiplist.NewItem(0, "hello"),
	//}
	//fmt.Println(items)
	//it := sl.Iterator()
	//fmt.Println(it.Next())
	//fmt.Println(it.Next())
	//fmt.Println(it.Next())
	//fmt.Println(it.HasNext())
	//fmt.Println(sl.String())
	//fmt.Println("original:")
	//fmt.Println(sl.String())
	//fmt.Println("original size:", sl.Size())
	//fmt.Println(sl.Range(-10, 4, false, false))

	//res := sl.Split(5)
	//fmt.Println("res:")
	//fmt.Println(res.String())
	//fmt.Println("sl:")
	//fmt.Println(sl.String())
	//fmt.Println("res size:", res.Size(), "sl size:", sl.Size())
	//fmt.Println(l1.String())
	//fmt.Println(l2)
	//slCustom := NewCustomSkipList[Person, int](16, 0.5)
	//slCustom.Insert(Person{"jordie"}, 26)
	//slCustom.Insert(Person{"mary"}, 52)
	//fmt.Println(slCustom.Search(Person{name: "mary"}))

	//slF := NewSkipList[Person, int](16, 0.5, Compare[Person])

}
