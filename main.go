package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/tkanos/gonfig"
	"log"
	"net/http"
	"sync"
)

type Configuration struct {
	MaxDepth     int
	MaxParallel  int
	Port         int
	CrawlerAgent string
}

var configuration Configuration

var seenMapMutex = sync.RWMutex{}
var indexCashMutex = sync.RWMutex{}

type Crawler struct {
	URI   string
	depth int
}

var indexCache = map[string]map[string]int{}

var sitesIndexed int
var wordsIndexed int

type wordCount struct {
	word  string
	count int
}
type indexResponse struct {
	SitesIndexed int
	WordsIndexed int
}

func main() {
	extractConfig("config.json")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/index", indexPageHandler).Methods("POST")
	router.HandleFunc("/index", deleteIndexHandler).Methods("DELETE")
	router.HandleFunc("/search/{word}", searchIndexForWordHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func extractConfig(filename string) {
	err := gonfig.GetConf(filename, &configuration)
	if err != nil {
		panic(err)
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println("Error writing Payload: ", payload)
	}
}
