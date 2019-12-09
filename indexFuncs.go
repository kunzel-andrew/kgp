package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/temoto/robotstxt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

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
	if resp.Body != nil {
		buf.ReadFrom(resp.Body)
	}
	body := buf.String()

	title, _ := getTitleFromBody(body)
	words, _ := getWordsFromBody(body)

	urlCache, totalWords := mapReduceWords(words)
	fmt.Println("Total Words Cached for Title", title, ":", strconv.Itoa(totalWords))
	updateCache(urlCache, title)

	links, _ := getLinksFromBody(body)
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
		return "", err
	}
	linkURL, err := url.Parse(link)
	if err != nil {
		return "", err
	}
	formattedURL := baseURL.ResolveReference(linkURL)
	if formattedURL.Scheme == "" || formattedURL.Host == "" {
		return "", errors.New("Cannot Format URL")
	}
	return formattedURL.String(), nil
}

func getTitleFromBody(body string) (string, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return "", err
	}
	title := document.Find("title").Text()
	return title, nil
}

func getLinksFromBody(body string) ([]string, error) {
	document, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	var links []string
	document.Find("a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists{
			links = append(links, href)
		}
	})

	return links, nil
}

func getWordsFromBody(body string) ([]string, error) {
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
	return words, nil
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