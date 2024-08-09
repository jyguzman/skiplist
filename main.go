package main

import "fmt"

func main() {
	sl := SkipListWithIntKeys(15, 0.5)
	sl.Insert(10, "hello")
	sl.Insert(20, "world")
	sl.Insert(5, "before")
	sl.Insert(30, "thirty")
	fmt.Println(sl.Search(10))
	fmt.Println(sl.Search(20))
	fmt.Println(sl.Search(30))
	fmt.Println(sl.String())
	//sl.Delete(10)
	//fmt.Println(sl.Search(10))
	//sl.Insert(20, "beach")
	//fmt.Println(sl.Search(20))
	//fmt.Println(sl.Search(30))
}
