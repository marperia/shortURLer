package tests

import (
	"github.com/marperia/shortURLer/controllers"
	"testing"
)

func TestBase16Decode(t *testing.T) {

	tables := []struct {
		firstParam  string
		secondParam string
		result      int
	}{
		{"2", controllers.Alphabet, 2,},
		{"b", controllers.Alphabet, 11,},
		{"19", controllers.Alphabet, 25,},
		{"20", controllers.Alphabet, 32,},
		{"21", controllers.Alphabet, 33,},
		{"22", controllers.Alphabet, 34,},
		{"5f3", controllers.Alphabet, 1523,},
		{"101", controllers.Alphabet, 257,},
		{"ff", controllers.Alphabet, 255,},
	}

	for _, table := range tables {
		total := controllers.BaseDecode(table.firstParam, table.secondParam)
		if table.result != total {
			t.Errorf("Test of %s and %s was incorrect, got: %d, want: %d.", table.firstParam, table.secondParam, total, table.result)
		}
	}
}

func TestBase16Encode(t *testing.T) {

	tables := []struct {
		firstParam  int
		secondParam string
		result      string
	}{
		{10, controllers.Alphabet, "a",},
		{100, controllers.Alphabet, "64",},
		{57, controllers.Alphabet, "39",},
		{58, controllers.Alphabet, "3a",},
		{59, controllers.Alphabet, "3b",},
		{599, controllers.Alphabet, "257",},
		{1523, controllers.Alphabet, "5f3",},
		{152323, controllers.Alphabet, "25303",},
	}

	for _, table := range tables {
		total := controllers.BaseEncode(table.firstParam, table.secondParam)
		if table.result != total {
			t.Errorf("Test of %d and %s was incorrect, got: %s, want: %s.", table.firstParam, table.secondParam, total, table.result)
		}
	}
}
