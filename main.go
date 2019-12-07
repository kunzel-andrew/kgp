package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/temoto/robotstxt"
	"github.com/tkanos/gonfig"
)
type Configuration struct{
	MaxDepth	int
	MaxParallel int
	Port		int
	CrawlerAgent string
}
var configuration Configuration

var seenMapMutex = sync.RWMutex{}
var indexCashMutex = sync.RWMutex{}

type Crawler struct {
	URI string
	depth int
}

var indexCache = map[string]map[string]int{}

var sitesIndexed int
var wordsIndexed int
type wordCount struct {
	word string
	count int
}
type indexResponse struct{
	SitesIndexed int
	WordsIndexed int
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		URL			string `json:"URL"`
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
		response = crawl(Crawler{parsedBody.URL, 0}, configuration.MaxParallel,)
		for _, entity := range response {
			totals.SitesIndexed += entity.SitesIndexed
			totals.WordsIndexed += entity.WordsIndexed
		}
		respondWithJSON(w, http.StatusOK, totals)

	}
}
func crawl(startLink Crawler, concurrency int) []indexResponse {
	results := []indexResponse{}
	type linkList struct {
		linkList 	[]string
		depth 		int
	}
	worklist := make(chan linkList)
	n := 1

	var tokens = make(chan struct{}, concurrency)
	go func() {worklist <- linkList{[]string{startLink.URI}, startLink.depth}}()
	seen := make(map[string]bool)

	finish := false
	list := linkList{}
	ok := false
	for ; n > 0; n-- {
		n ++
		select {
		case list, ok = <-worklist:
			if ok {
				fmt.Println("Got a new list")
			}
		default:
			finish = true
		}
		depth := list.depth
		for _, link := range list.linkList {
			absoluteLink, err := formatURL(link, startLink.URI)
			seenMapMutex.RLock()
			seenLink, ok := seen[absoluteLink]
			seenMapMutex.RUnlock()
			if !ok{
				seenLink = false
			}
			if err == nil && startLink.URI != "" && canCrawl(absoluteLink) && !seenLink && startLink.depth+1 < configuration.MaxDepth {
				seenMapMutex.Lock()
				seen[absoluteLink] = true
				seenMapMutex.Unlock()

				go func(link string, token chan struct{}) {
					foundLinks, depth, pageResults := indexPage(Crawler{link, depth}, token)
					results = append(results, pageResults)
					if foundLinks != nil {
						worklist <- linkList{foundLinks, depth}
					} else if finish {
						n=0
					}
				}(absoluteLink, tokens)
			} else {
				if !canCrawl(absoluteLink) {
					fmt.Println("Cannot Legally Crawl Link ", absoluteLink)
				} else if seen[absoluteLink] {
					fmt.Println("Already seen link", absoluteLink, " Skipping")
				}
			}
		}
	}
	return results
}
func indexPage(uri Crawler, token chan struct{}) ([]string, int, indexResponse){
	token <- struct{}{}
		fmt.Println("Indexing: ", uri.URI, "at depth", strconv.Itoa(uri.depth))
		resp, _ := getRequest(uri.URI)
	<-token

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	title := getTitleFromBody(body)
	words := getWordsFromBody(body)

	urlCache, totalWords := mapReduceWords(words)
	fmt.Println("Total Words Cached for Title", title, ":", strconv.Itoa(totalWords))
	updateCache(urlCache, title)

	links := getLinksFromBody(body)
	//If Max Depth is reached don't continue adding links to the queue
	if uri.depth >= configuration.MaxDepth {
		links = nil
	}
	return links, uri.depth+1, indexResponse{1, totalWords}


}

func getRequest(uri string) (*http.Response, error) {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", uri, nil)
	req.Header.Set("User-Agent", configuration.CrawlerAgent)

	res, err := client.Do(req)
	if err != nil{
		return nil, err
	}

	return res, nil
}

func canCrawl(URL string) bool{
	//Check robots.txt

	parsedUrl, err := url.Parse(URL)
	if err != nil {
		log.Fatal(err)
	}
	robotsURL := parsedUrl.Scheme+"://"+parsedUrl.Host+"/robots.txt"
	resp, err := http.Get(robotsURL)
	if err != nil {
		return false
	}

	data, err := robotstxt.FromResponse(resp)
	resp.Body.Close()
	if err != nil {
		log.Println("Error parsing robots.txt for URL", robotsURL, err.Error())
		return false
	}
	return data.TestAgent(URL, configuration.CrawlerAgent)
}

func formatURL(link string, base string) (string, error){
	baseURL, err := url.Parse(base)
	if err != nil {
		panic(err)
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		panic(err)
	}
	formattedURL := baseURL.ResolveReference(linkURL)
	if formattedURL.Scheme == "" || formattedURL.Host == "" {
		return "", errors.New("Cannot Format URL")
	}
	return formattedURL.String(), nil
}

func getTitleFromBody(body string) string {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	title := document.Find("title").Text()
	return title
}

func getLinksFromBody(body string) []string {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		panic(err)
	}
	var links []string
	document.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists{
			links = append(links, href)
		}
	})

	return links
}

func getWordsFromBody(body string) []string {
	var words []string
	domDoc := html.NewTokenizer(strings.NewReader(body))
	startToken := domDoc.Token()
	loopDom:
		for {
			tt:= domDoc.Next()
			switch {
			case tt == html.ErrorToken:
				break loopDom
			case tt == html.StartTagToken:
				startToken = domDoc.Token()
			case tt == html.TextToken:
				if startToken.Data == "script" {
					continue
				}
				textContent := strings.TrimSpace(html.UnescapeString(string(domDoc.Text())))
				if len(textContent) > 0 {
					wordStrings := strings.Fields(textContent)
					for _, word := range wordStrings {
						words = append(words, word)
					}
				}
			}
		}
	return words
}

func mapReduceWords(words []string) (map[string]int, int){
	var data = make(map[string]int)

	for _, word := range words {
		word = strings.ToLower(word)
		if match, _  := regexp.MatchString("^[a-z]+$", word); match {
			count := data[word]
			data[word] = count + 1
		}

	}

	return data, len(data)
}

func updateCache(data map[string]int, title string) map[string]map[string]int{
	for word, count := range data {
		indexCashMutex.RLock()
		if _, found := indexCache[word]; !found {
			indexCache[word] = make(map[string]int)
		}
		indexCashMutex.RUnlock()

		indexCashMutex.Lock()
		indexCache[word][title] = count
		indexCashMutex.Unlock()
	}


	return indexCache
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

func searchIndexForWord(word string) PairList {
	if titles, ok := indexCache[word]; ok {
		pl := make(PairList, len(titles))
		i := 0
		for k, v := range titles {
			pl[i] = Pair{k,v}
			i++
		}
		sort.Sort(sort.Reverse(pl))

		return pl
	}
	return nil
}

type Pair struct {
	Title string
	Count int
}
type PairList []Pair

func (p PairList) Len() int {return len(p)}
func (p PairList) Less(i, j int) bool { if p[i].Count == p[j].Count {
											return p[i].Title < p[j].Title
										} else {
											return p[i].Count < p[j].Count
										} }
func (p PairList) Swap(i,j int) { p[i], p[j] = p[j], p[i]}

func main() {
	extractConfig("config.json")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeHandler)
	router.HandleFunc("/index", indexPageHandler).Methods("POST")
	router.HandleFunc("/index", deleteIndexHandler).Methods("DELETE")
	router.HandleFunc("/search/{word}", searchIndexForWordHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("This is working")
	respondWithJSON(w, 200, "Yay. Thanks")
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
	if err:=json.NewEncoder(w).Encode(payload); err != nil {
		fmt.Println("Error writing Payload: ", payload)
	}
}