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
	Port		int
}
var configuration Configuration
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

func homeLink(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome home!")
}

func indexPageHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		URL			string `json:"URL"`
	}
	var parsedBody body

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please provide a URL to index")
	}
	json.Unmarshal(reqBody, &parsedBody)

	indexPage(Crawler{parsedBody.URL, 0})
}

func indexPage(uri Crawler) (int, int){
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

	return sitesIndexed, wordsIndexed
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
	return data.TestAgent(URL, "Go-http-client/1.1")
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


func deleteIndexHandler(w http.ResponseWriter, r *http.Request) {
	indexCache = make(map[string]map[string]int)
}

func searchIndexForWordHandler(w http.ResponseWriter, r *http.Request) {
	type body struct {
		word			string `json:"URL"`
	}
	var parsedBody body

	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "Please provide a URL to index")
	}
	json.Unmarshal(reqBody, &parsedBody)

	fmt.Println(searchIndexForWord(parsedBody.word))
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
	key string
	value int
}
type PairList []Pair

func (p PairList) Len() int {return len(p)}
func (p PairList) Less(i, j int) bool { if p[i].value == p[j].value {
											return p[i].key < p[j].key
										} else {
											return p[i].value < p[j].value
										} }
func (p PairList) Swap(i,j int) { p[i], p[j] = p[j], p[i]}

func main() {
	extractConfig("config.json")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
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