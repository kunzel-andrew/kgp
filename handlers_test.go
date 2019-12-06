package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var a App

func TestMain(m *testing.M){
	a = App{}
	a.Initialize()

	code := m.Run()

	os.Exit(code)
}

func TestSearchIndexForWordHandler(t *testing.T){
	testCache := map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3, "Test Title 5": 1},
		"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
		"c": map[string]int{"Test Title 2": 1}}

	fixtures := []struct{
		cache 	map[string]map[string]int
		word  string
		expectedCode int
		expectedResult string
	}{
		{testCache, "a", http.StatusOK, "[{\"Title\":\"Test Title 4\",\"Count\":3},{\"Title\":\"Test Title 1\",\"Count\":2},{\"Title\":\"Test Title 5\",\"Count\":1},{\"Title\":\"Test Title 2\",\"Count\":1}]"},
		{testCache, "d", http.StatusOK, "null"},
		{testCache, "123", http.StatusNotFound, "404 page not found\n"},
		{testCache, "", http.StatusNotFound, "404 page not found\n"},
		{nil, "a", http.StatusOK, "null"},
	}
	for _, fixture := range fixtures {
		indexCache = fixture.cache
		req, _ := http.NewRequest("GET", "/search/"+fixture.word, nil)
		response := executeRequest(req)

		checkResponseCode(t, fixture.expectedCode, response.Code)

		if body := response.Body.String(); body != fixture.expectedResult {
			t.Errorf("Incorrect Response Body. Expected:%s Received%s", fixture.expectedResult, body)
		}
	}
}

func TestDeleteCache(t *testing.T){
	indexCache = map[string]map[string]int{"a": map[string]int{"Test Title 1": 2, "Test Title 2": 1, "Test Title 4": 3, "Test Title 5": 1},
		"b": map[string]int{"Test Title 1": 1, "Test Title 2": 1},
		"c": map[string]int{"Test Title 2": 1}}

	req, _ := http.NewRequest("DELETE", "/index", nil)
	response := executeRequest(req)

	checkResponseCode(t, http.StatusNoContent, response.Code)

	if len(indexCache) > 0 {
		t.Error("The Cache was not deleted")
	}
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	a.Router.ServeHTTP(rr, req)

	return rr
}

func checkResponseCode(t *testing.T, expected, actual int) {
	if expected != actual {
		t.Errorf("Expected response code %d. Got %d", expected, actual)
	}
}
