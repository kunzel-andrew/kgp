package main

import (
	"reflect"
	"testing"
)

func TestSearchIndexForWord(t *testing.T) {
	testCache := map[string]map[indexCacheInfo]int{"a": {indexCacheInfo{"Test Title 1", "test.com"}: 2, indexCacheInfo{"Test Title 2", "test.com"}: 1, indexCacheInfo{"Test Title 4", "test.com"}: 3, indexCacheInfo{"Test Title 5", "test.com"}: 1},
		"b": {indexCacheInfo{"Test Title 1", "test.com"}: 1, indexCacheInfo{"Test Title 2", "test.com"}: 1},
		"c": {indexCacheInfo{"Test Title 2", "test.com"}: 1}}
	fixtures := []struct {
		word   string
		cache  map[string]map[indexCacheInfo]int
		result PairList
	}{
		{"a", testCache, []Pair{{indexCacheInfo{"Test Title 4", "test.com"}, 3},
			{indexCacheInfo{"Test Title 1", "test.com"}, 2},
			{indexCacheInfo{"Test Title 5", "test.com"}, 1},
			{indexCacheInfo{"Test Title 2", "test.com"}, 1}}},
		{"", testCache, nil},
		{"d", testCache, nil},
		{"a", nil, nil},
		{"b", testCache, []Pair{{indexCacheInfo{"Test Title 2", "test.com"}, 1},
			{indexCacheInfo{"Test Title 1", "test.com"}, 1}}},
	}

	for _, fixture := range fixtures {
		indexCache = fixture.cache
		testResult := searchIndexForWord(fixture.word)
		if !reflect.DeepEqual(testResult, fixture.result) {
			t.Error("Result: ", testResult, "does not match expected", fixture.result)
		}
	}
}
