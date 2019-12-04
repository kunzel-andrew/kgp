package main

import (
	"testing"
)

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
		words := getWordsFromBody(fixture.body)
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
