package main

import (
	"testing"
)


func TestDeleteCache(t *testing.T){
	indexCache = map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3, "Test Title 5": 1},
											"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
											"c": map[string]int{"Test Title 2": 1}}

	deleteIndexHandler(nil, nil)

	if len(indexCache) > 0 {
		t.Error("The Cache was not deleted")
	}
}
