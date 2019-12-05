package main

import (
	"reflect"
	"testing"
)


func TestMapReduceWords(t *testing.T){
	fixtures := []struct{
		words []string
		cache map[string]int
	}{
		{[]string{"a","a","b"}, map[string]int{"a": 2, "b": 1}},
		{[]string{"a","b","c"}, map[string]int{"a": 1, "b": 1, "c": 1}},
		{[]string{}, map[string]int{}},
		{[]string{"a","A","A"}, map[string]int{"a": 3}},
		{[]string{",","<","a"}, map[string]int{"a": 1}},
	}

	for _, fixture := range fixtures {
		cache, numWords := mapReduceWords(fixture.words)
		if !reflect.DeepEqual(cache, fixture.cache) {
			t.Error("cache: ", cache, "does not match expected", fixture.cache)
		}
		if numWords != len(fixture.cache) {
			t.Error("The word count returned", numWords, "does not match expected", len(fixture.cache))
		}
	}
}
