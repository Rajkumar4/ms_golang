package main

import "fmt"

func main() {
	var ls [][]int
	temp := [][]int{{1, 2, 4}}
	ls = append(ls, temp...)
	fmt.Printf("list %v", ls)
}
