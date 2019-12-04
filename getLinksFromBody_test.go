package main

import (
	"testing"
)

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
		links := getLinksFromBody(fixture.body)
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
