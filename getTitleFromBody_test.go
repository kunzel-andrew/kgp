package main

import (
	"testing"
)

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
		title := getTitleFromBody(fixture.body)
		if title != fixture.result {
			t.Errorf("Expected Title: %s but received %s", fixture.result, title)
		}
	}
}
