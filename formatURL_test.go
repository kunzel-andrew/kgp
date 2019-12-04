package main

import (
	"testing"
)


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
