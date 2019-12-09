package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"strings"
)

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		URL string `json:"URL"`
	}
	var parsedBody body
	var response []indexResponse
	var totals indexResponse

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to read Body")
	}
	defer r.Body.Close()
	json.Unmarshal(reqBody, &parsedBody)

	if parsedBody.URL == "" {
		respondWithError(w, http.StatusUnprocessableEntity, "Please include URL in Body of Request")
	} else {
		fmt.Println("Beginning to index at:", parsedBody.URL)
		response = crawl(Crawler{parsedBody.URL, 0}, configuration.MaxParallel)
		for _, entity := range response {
			totals.SitesIndexed += entity.SitesIndexed
			totals.WordsIndexed += entity.WordsIndexed
		}
		respondWithJSON(w, http.StatusOK, totals)

	}
}

func deleteIndexHandler(w http.ResponseWriter, r *http.Request) {
	indexCache = make(map[string]map[string]int)

	respondWithJSON(w, http.StatusNoContent, "")
}

func searchIndexForWordHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	word, _ := params["word"]
	response := searchIndexForWord(strings.ToLower(word))
	respondWithJSON(w, http.StatusOK, response)
}
