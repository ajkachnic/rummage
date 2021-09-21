package main

import (
	"fmt"
	"log"
	"rummage/spider"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	// Crawl MDN
	seed := spider.HttpSeed{Url: "https://developer.mozilla.org/en-US/docs/Web"}
	result, err := seed.Crawl(0, &wg)
	if err != nil {
		log.Fatal(err)
	}

	logResult(result, "")

}

func logResult(result spider.CrawlResult, prefix string) {
	fmt.Println(prefix + result.Title)
	for _, res := range result.Seeds {
		if res.Title != "" {
			logResult(res, prefix+"  ")
		}
	}
}
