package main

import (
	"reflect"
	"testing"
)


func TestUpdateCache(t *testing.T){
	fixtures := []struct{
		data map[string]int
		title string
		cache map[string]map[string]int
	}{
		{ map[string]int{"a": 2, "b": 1}, "Test Title 1", map[string]map[string]int{"a": map[string]int{"Test Title 1": 2},
																									  "b": map[string]int{"Test Title 1": 1}}},
		{ map[string]int{"a": 1, "b": 1, "c": 1}, "Test Title 2", map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1},
																							"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
																							"c": map[string]int{"Test Title 2": 1}}},
		{ map[string]int{},"Test Title 3", map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1},
																						"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
																						"c": map[string]int{"Test Title 2": 1}}},
		{ map[string]int{"a": 3}, "Test Title 4", map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3},
																								"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
																								"c": map[string]int{"Test Title 2": 1}}},
		{ map[string]int{"a": 1}, "Test Title 5", map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3, "Test Title 5": 1},
																								"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
																								"c": map[string]int{"Test Title 2": 1}}},
	}

	for _, fixture := range fixtures {
		updatedCache := updateCache(fixture.data, fixture.title)
		if !reflect.DeepEqual(updatedCache, fixture.cache) {
			t.Error("cache: ", updatedCache, "does not match expected", fixture.cache)
		}
	}
}
