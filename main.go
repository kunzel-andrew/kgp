package main

import (
	"bytes"
	"crypto/tls"
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
type App struct {
	Router *mux.Router
}

var seen = make(map[string]bool)

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
	sitesIndexed int
	wordsIndexed int
}

func (a *App) indexPageHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		URL			string `json:"URL"`
	}
	var parsedBody body
	var response indexResponse
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Please include URL in Body of Request")
	}
	json.Unmarshal(reqBody, &parsedBody)

	response = indexPage(Crawler{parsedBody.URL, 0})
	respondWithJSON(w, http.StatusOK, response)

}

func indexPage(uri Crawler) indexResponse {
	//Set Up the Queue to do a Breadth First Search
	queue := make(chan Crawler)
	seen[uri.URI] = true
	sitesIndexed = 0
	wordsIndexed = 0

	go func() {
		queue <- uri
	}()

	for uri := range queue {
		enqueue(uri, queue)
	}

	return indexResponse{sitesIndexed, wordsIndexed}
}

func enqueue(uri Crawler, queue chan Crawler) {
	fmt.Println("Indexing", uri.URI, "At Depth", uri.depth)
	sitesIndexed += 1
	seen[uri.URI] = true
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	client := http.Client{Transport: transport}
	resp, err := client.Get(uri.URI)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	title := getTitleFromBody(body)
	words := getWordsFromBody(body)

	urlCache, totalWords := mapReduceWords(words)
	wordsIndexed += totalWords
	fmt.Println("Total Words Cached for Title", title, ":", strconv.Itoa(totalWords))
	updateCache(urlCache, title)

	links := getLinksFromBody(body)

	for _, link := range links {
		absoluteLink, err := formatURL(link, uri.URI)
		if err == nil && uri.URI != "" &&  canCrawl(absoluteLink) && !seen[absoluteLink] && uri.depth+1 < configuration.MaxDepth{
			next := Crawler{absoluteLink, uri.depth+1}
			seen[absoluteLink] = true
			fmt.Println("Added Link to crawl:", absoluteLink)
			go func() { queue <- next }()
		} else {
			if !canCrawl(absoluteLink) {
				fmt.Println("Cannot Legally Crawl Link ", absoluteLink)
			} else if seen[absoluteLink] {
				fmt.Println("Already seen link", absoluteLink, " Skipping")
			} else if uri.depth+1 >= configuration.MaxDepth {
				fmt.Println("Not crawling links on this page because max depth has been reached")
			}

		}
	}
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
		if _, found := indexCache[word]; !found {
			indexCache[word] = make(map[string]int)
		}
		indexCache[word][title] = count
	}
	return indexCache
}


func (a *App) deleteIndexHandler(w http.ResponseWriter, r *http.Request) {
	indexCache = make(map[string]map[string]int)

	respondWithJSON(w, http.StatusNoContent, "")
}

func (a *App) searchIndexForWordHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	word, _ := params["word"]
	response := searchIndexForWord(word)
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

func (a *App) Initialize() {
	extractConfig("config.json")

	a.Router = mux.NewRouter().StrictSlash(true)
	a.initializeRoutes()
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(":8000", a.Router))
}

func (a *App) initializeRoutes() {
	a.Router.HandleFunc("/index", a.indexPageHandler).Methods("POST")
	a.Router.HandleFunc("/index", a.deleteIndexHandler).Methods("DELETE")
	a.Router.HandleFunc("/search/{word:[a-zA-Z]+}", a.searchIndexForWordHandler).Methods("GET")

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
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}