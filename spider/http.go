package spider

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

var cache = map[string]bool{}
var cacheLock sync.Mutex

const TIER_THRESHOLD = 2

// Http(s) crawler
type HttpSeed struct {
	Url string
}

func (seed *HttpSeed) Crawl(tier int, wg *sync.WaitGroup) (CrawlResult, error) {
	addToCache(seed.Url, true)
	log.Println("Crawling " + seed.Url + ", tier " + fmt.Sprint(tier))
	resp, err := http.Get(seed.Url)
	if err != nil {
		return CrawlResult{}, err
	}

	contentType := resp.Header.Get("content-type")

	if strings.HasPrefix(contentType, "text/html") {
		return CrawlHtml(resp, tier, wg)
	}

	// return an error
	return CrawlResult{}, nil
}

func CrawlHtml(resp *http.Response, tier int, wg *sync.WaitGroup) (CrawlResult, error) {
	doc, err := html.Parse(resp.Body)

	if err != nil {
		return CrawlResult{}, err
	}

	// Extract data from the document
	title := ""
	description := ""
	content := ""
	seeds := []HttpSeed{}

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "title" {
			title = n.FirstChild.Data
		} else if n.Type == html.ElementNode && n.Data == "a" {
			href := getAttr("href", n.Attr)
			if href != "" && tier < TIER_THRESHOLD {
				url, err := handleRelativeUrl(resp.Request.URL.String(), href)
				if err != nil {
					log.Fatalln(err)
				}
				if !strings.HasPrefix(url, "mailto:") {
					seed := HttpSeed{Url: url}
					seeds = append(seeds, seed)
				}
			}
		} else if n.Type == html.ElementNode && n.Data == "meta" {
			if getAttr("name", n.Attr) == "description" {
				description = getAttr("content", n.Attr)
			}
		} else if n.Type == html.TextNode {
			content += n.Data
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	deduppedSeeds := dedup(seeds)
	crawledSeeds := []CrawlResult{}

	wg.Add(len(deduppedSeeds))
	for _, seed := range deduppedSeeds {
		if !checkCache(seed.Url) {
			go func() {
				res, err := seed.Crawl(tier+1, wg)
				fmt.Println(res)
				if err != nil {
					log.Fatalln(err)
				}
				crawledSeeds = append(crawledSeeds, res)
				wg.Done()
				time.Sleep(100) // Pause for a millisecond
			}()
			time.Sleep(100) // Pause for a millisecond
		} else {
			fmt.Println("found cached url")
		}
	}

	wg.Wait()

	return CrawlResult{
		Title:       title,
		Description: description,
		Content:     content,
		Seeds:       crawledSeeds,
	}, nil
}

func handleRelativeUrl(base string, relative string) (string, error) {
	baseParsed, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	relativeParsed, err := url.Parse(relative)
	relativeParsed.Fragment = ""
	if err != nil {
		return "", err
	}

	if relativeParsed.IsAbs() {
		return relativeParsed.String(), nil
	}

	return baseParsed.ResolveReference(relativeParsed).String(), nil
}

func getAttr(name string, attr []html.Attribute) string {
	for _, attr := range attr {
		if attr.Key == name {
			return attr.Val
		}
	}
	return ""
}

func checkCache(url string) bool {
	cacheLock.Lock()
	defer cacheLock.Unlock()
	_, ok := cache[url]
	return ok
}

func addToCache(url string, res bool) {
	fmt.Println("adding to cache " + url)
	cacheLock.Lock()
	defer cacheLock.Unlock()
	cache[url] = res
}

func dedup(seeds []HttpSeed) []HttpSeed {
	occured := map[HttpSeed]bool{}
	result := []HttpSeed{}

	for _, val := range seeds {
		if occured[val] != true {
			occured[val] = true

			result = append(result, val)
		}
	}

	return result
}
