package wumanber

import (
	"testing"
	"fmt"
	"strconv"
)

func TestInit(t *testing.T)  {
	var patterns []string = []string{"你好", "世界"}
	var wm WuManber
	err := wm.Init(patterns)
	if err != nil {
		t.Error("init wumanber error")
	}
	if 3 != wm.Block {
		t.Error("block size is not 3")
	}
	if 1003 != wm.TableSize {
		t.Error("table size if not 2")
	}

	if 1003 != len(wm.ShiftTable) {
		t.Error("shift table size if not 2")
	}
	if 1003 != len(wm.HashTable) {
		t.Error("hash table size if not 2")
	}
}

func TestSearch(t *testing.T)  {
	var patterns []string = []string{"你好", "世界", "abc"}
	var wm WuManber
	err := wm.Init(patterns)
	if err != nil {
		t.Error("init wumanber error")
	}

	fmt.Println("patterns: ")
	fmt.Println(patterns)
	var testStrings []string = []string{"你们好", "abcdefg", "o世界很大", "北京你好，世界很大啊"}
	var hitArr []int = []int{0, 1, 1, 2}
	for i, str := range testStrings {

		hits := wm.Search(str)
		fmt.Println("str: " + str + " hits: " + strconv.Itoa(hits))
		if hits != hitArr[i] {
			t.Error("match error" + str)
		}
	}
}