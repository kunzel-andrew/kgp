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

func indexPage(uri Crawler) {
	//Set Up the Queue to do a Breadth First Search
	queue := make(chan Crawler)

	go func() {
		queue <- uri
	}()

	for uri := range queue {
		enqueue(uri, queue)
	}
}

func enqueue(uri Crawler, queue chan Crawler) {
	fmt.Println("Indexing", uri.URI, "At Depth", uri.depth)
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

	words:= getWordsFromBody(body)
	fmt.Println("Words from the site: ", words)
	links := getLinksFromBody(body)
	fmt.Println("Attempting to Crawl the following links:", links)
	for _, link := range links {
		absoluteLink, err := formatURL(link, uri.URI)
		if err == nil && uri.URI != "" &&  canCrawl(absoluteLink) && !seen[absoluteLink] && uri.depth+1 < configuration.MaxDepth{
			next := Crawler{absoluteLink, uri.depth+1}
			go func() { queue <- next }()
		} else {
			if !canCrawl(absoluteLink) {
				fmt.Println("Cannot Crawl Legally Crawl Link ", absoluteLink)
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

	resp, err := http.Get(parsedUrl.Scheme+"://"+parsedUrl.Host+"/robots.txt")
	if err != nil {
		return false
	}

	data, err := robotstxt.FromResponse(resp)
	resp.Body.Close()
	if err != nil {
		log.Println("Error parsing robots.txt", err.Error())
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

/*	var newEvent event

	events = append(events, newEvent)
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newEvent)*/


func main() {
	extractConfig("config.json")

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", homeLink)
	router.HandleFunc("/index", indexPageHandler).Methods("POST")
	//router.HandleFunc("/index", deleteIndex).Methods("DELETE")
	//router.HandleFunc("/index/{url}", getIndexForURL).Methods("GET")
	//router.HandleFunc("/search/{word}", searchIndexForWord).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func extractConfig(filename string) {
	err := gonfig.GetConf(filename, &configuration)
	if err != nil {
		panic(err)
	}
}