package main

import (
	"reflect"
	"testing"
)


func TestSearchIndexForWord(t *testing.T){
	testCache := map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3, "Test Title 5": 1},
												"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
												"c": map[string]int{"Test Title 2": 1}}
	fixtures := []struct{
		word string
		cache map[string]map[string]int
		result PairList
	}{
		{"a", testCache, []Pair{{"Test Title 4", 3},
														{"Test Title 1", 2},
														{"Test Title 5", 1},
														{"Test Title 2", 1}}},
		{"", testCache, nil},
		{"d", testCache, nil},
		{"a", nil, nil},
		{"b", testCache,[]Pair{{"Test Title 2", 1},
													{"Test Title 1", 1 }} },
	}

	for _, fixture := range fixtures {
		indexCache = fixture.cache
		testResult := searchIndexForWord(fixture.word)
		if !reflect.DeepEqual(testResult, fixture.result) {
			t.Error("Result: ", testResult, "does not match expected", fixture.result)
		}
	}
}
