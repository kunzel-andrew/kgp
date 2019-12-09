package main

import (
	"encoding/json"
	"github.com/jarcoal/httpmock"
	"reflect"
	"testing"
)

func TestIndexPage(t *testing.T){
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "http://www.testError.test", httpmock.NewBytesResponder(500, nil))
	httpmock.RegisterResponder("GET", "http://www.test.test/a",httpmock.NewBytesResponder(200, []byte("<head><title>Test Title</title></head><a href=\"https://test.test/b\">Test Link</a>") ))

	fixtures := []struct{
		URI Crawler
		MaxDepth int
		expectedLinks []string
		expectedDepth int
		expectedIndex indexResponse
	} {
		{Crawler{"http://www.testError.test", 0}, 3, []string{}, 0, indexResponse{0, 0}},
		{Crawler{"http://www.test.test/a", 0}, 3, []string{"https://test.test/b"}, 1, indexResponse{1, 3}},
		{Crawler{"http://www.test.test/a", 0}, 0, []string{}, 1, indexResponse{1, 3}},
	}
	for _, fixture := range fixtures {
		var tokens = make(chan struct{}, 1)
		configuration.MaxDepth = fixture.MaxDepth

		links, depth, index := indexPage(fixture.URI, tokens)
		if !reflect.DeepEqual(links, fixture.expectedLinks){
			t.Errorf("Error in returned Links expected: %s received %s", links,fixture.expectedLinks)
		}
		if depth != fixture.expectedDepth {
			t.Errorf("Error in returned Depth expected %d received %d", depth, fixture.expectedDepth)
		}

		if !reflect.DeepEqual(index, fixture.expectedIndex) {
			indexJson, _ := json.Marshal(index)
			expectedIndexJson, _ := json.Marshal(fixture.expectedIndex)
			t.Errorf("Error in returned IndexResponse expected %s received %s", indexJson, expectedIndexJson)
		}


	}
}

/*
func TestCrawler(t *testing.T){
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder("GET", "http://www.testError.test", httpmock.NewBytesResponder(500, nil))
	httpmock.RegisterResponder("GET", "http://www.test.test/robots.txt", httpmock.NewBytesResponder(401, nil))
	httpmock.RegisterResponder("GET", "http://www.test.test/a",httpmock.NewBytesResponder(200, []byte("<head><title>Test Title a</title></head><a href=\"https://www.test.test/b\">Test Link</a>") ))
	httpmock.RegisterResponder("GET", "http://www.test.test/b",httpmock.NewBytesResponder(200, []byte("<head><title>Test Title b</title></head><a href=\"https://www.test.test/c\">Test Link</a>") ))
	httpmock.RegisterResponder("GET", "http://www.test.test/c",httpmock.NewBytesResponder(200, []byte("<head><title>Test Title c</title></head><a href=\"https://www.test.test/d\">Test Link</a>") ))
	httpmock.RegisterResponder("GET", "http://www.test.test/d",httpmock.NewBytesResponder(200, []byte("<head><title>Test Title d</title></head><a href=\"https://www.test.test/d\">Test Link</a>") ))

	fixtures := []struct{
		URI Crawler
		MaxDepth int
		expectedIndex []indexResponse
	} {
//		{Crawler{"http://www.testError.test", 0}, 3, []indexResponse{}},
		{Crawler{"http://www.test.test/a", 0}, 3, []indexResponse{indexResponse{1, 3}, indexResponse{1,3}, indexResponse{1,3}}},
//		{Crawler{"http://www.test.test/a", 0}, 0, []indexResponse{indexResponse{1, 3}}},
	}
	for _, fixture := range fixtures {
		configuration.MaxDepth = fixture.MaxDepth
		indexCache = map[string]map[string]int{}
		index := crawl(fixture.URI, 1)

		if !reflect.DeepEqual(index, fixture.expectedIndex) {
			indexJson, _ := json.Marshal(index)
			expectedIndexJson, _ := json.Marshal(fixture.expectedIndex)
			t.Errorf("Error in returned IndexResponse expected %s received %s", expectedIndexJson, indexJson)
		}


	}
}*/
func TestLinksFromBody(t *testing.T){
	fixtures := []struct{
		body 	string
		result  []string
	}{
		{"<a href=\"https://test.com/test\">Test Link</a>", []string{"https://test.com/test"}},
		{"", []string{}},
		{"<a href=\"https://test.com/test\">Test Link</a>\n<a href=\"https://test.com/test2\">Test2 Link</a>", []string{"https://test.com/test", "https://test.com/test2"}},
	}

	for _, fixture := range fixtures {
		links,  err := getLinksFromBody(fixture.body)
		if err != nil {
			t.Error(err)
		}

		if len(links) != len(fixture.result) {
			t.Error()
		}
		for i, v := range links{
			if v != fixture.result[i]{
				t.Error()
			}
		}
	}
}

func TestTitleFromBody(t *testing.T){
	fixtures := []struct{
		body 	string
		result  string
	}{
		{"<head><title>Test Title</title></head><a href=\"https://test.com/test\">Test Link</a>", "Test Title"},
		{"", ""},
		{"<a href=\"https://test.com/test\">Test Link</a>\n<a href=\"https://test.com/test2\">Test2 Link</a>", ""},
	}

	for _, fixture := range fixtures {
		title,  err := getTitleFromBody(fixture.body)
		if err != nil {
			t.Error(err)
		}
		if title != fixture.result {
			t.Errorf("Expected Title: %s but received %s", fixture.result, title)
		}
	}
}

func TestWordsFromBody(t *testing.T){
	fixtures := []struct{
		body 	string
		result  []string
	}{
		{"<a href=\"https://test.com/test\">Test Link</a><br>This Is Words<br>", []string{"Test", "Link", "This", "Is", "Words"}},
		{"<a href=\"https://test.com/test\">Test Link</a><br>Test Link<br>", []string{"Test", "Link", "Test","Link"}},
		{"<a href=\"https://test.com/test\"></a>", []string{}},
		{"", []string{}},
	}

	for _, fixture := range fixtures {
		words, err := getWordsFromBody(fixture.body)
		if err != nil {
			t.Error(err)
		}

		if len(words) != len(fixture.result) {
			t.Error()
		}
		for i, v := range words{
			if v != fixture.result[i]{
				t.Error()
			}
		}
	}
}

func TestCanCrawl(t *testing.T){
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	fixtures := []struct{
		URL 	string
		robotsResponse 	httpmock.Responder
		result  bool
	}{
		{"http://www.test.com/test",  httpmock.NewBytesResponder(200, []byte("User-Agent: * \nDisallow: /")), false},
		{"http://www.test.com/test",  httpmock.NewBytesResponder(200, []byte("User-Agent: * \nDisallow: */test")), false},
		{"http://www.test.com/test",  httpmock.NewBytesResponder(200, []byte("User-Agent: * \nAllow: */test")), true},
		{"http://www.test.com/test",  httpmock.NewBytesResponder(403, nil), true},
		{"http://www.test.com/test",  httpmock.NewBytesResponder(401, nil), true},
		{"http://www.test.com/test",  httpmock.NewBytesResponder(500, nil), false},
		//		{"https://test.com/test/test1.13.1.linux-arm64.tar.gz",  httpmock.NewBytesResponder(403, nil), false},
	}
	for _, fixture := range fixtures {
		httpmock.RegisterResponder("GET", "http://www.test.com/robots.txt",fixture.robotsResponse)
		robotCrawl := canCrawl(fixture.URL)
		if robotCrawl != fixture.result {
			t.Error()
		}
	}
}

func TestFormatURL(t *testing.T){
	fixtures := []struct{
		link 	string
		base string
		result  string
		err error
	}{
		{"/test",  "http://www.test.com", "http://www.test.com/test", nil},
		{"http://www.test.com/test",  "http://www.test.com", "http://www.test.com/test", nil},
		{"http://www.test2.com/test",  "http://www.test.com", "http://www.test2.com/test", nil},
		{"#",  "http://www.test.com", "http://www.test.com", nil},
		{"test.com/test",  "http://www.test.com", "http://www.test.com/test.com/test", nil},
	}

	for _, fixture := range fixtures {
		formattedURL, err := formatURL(fixture.link, fixture.base)
		if err != fixture.err || formattedURL != fixture.result {
			t.Error(fixture, formattedURL, err)
		}
	}
}

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
	indexCache = map[string]map[string]int{}

	for _, fixture := range fixtures {
		updatedCache := updateCache(fixture.data, fixture.title)
		if !reflect.DeepEqual(updatedCache, fixture.cache) {
			t.Error("cache: ", updatedCache, "does not match expected", fixture.cache)
		}
	}
}
