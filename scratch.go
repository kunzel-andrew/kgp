package main

import (
	"testing"
)

func TestScratch(t *testing.T){
	extractConfig("config.json")
	indexPage(Crawler{"https://www.golang.org/", 0})
}
