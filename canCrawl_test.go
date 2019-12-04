package main

import (
	"github.com/jarcoal/httpmock"
	"testing"
)

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
	}

	for _, fixture := range fixtures {
		httpmock.RegisterResponder("GET", "http://www.test.com/robots.txt",fixture.robotsResponse)
		robotCrawl := canCrawl(fixture.URL)
		if robotCrawl != fixture.result {
			t.Error()
		}
	}
}
