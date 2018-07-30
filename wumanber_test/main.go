 package main

 import (
 	"fmt"
 	"wumanber"
 )

 func main() {
	var w wumanber.WuManber
	//var patterns []string = []string{"abc", "bcd", "SoulSense"}
	//err := w.Init(patterns)
	//if err != nil {
	//	fmt.Println("errors")
	//}
	//hits := w.Search("abcedSoul呵呵SoulSense")
	err := w.Deserialize("model.bin")
	if err != nil {
		fmt.Println("serialize error")
		return
	}
	hits := w.Search("abcedSoul呵呵SoulSense")
	fmt.Println(hits)
}