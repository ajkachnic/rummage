// The central interface to the crawler.
package spider

import "sync"

type CrawlResult struct {
	Title       string
	Content     string
	Description string
	Seeds       []CrawlResult
	Tier        int
	Type        string
}

// A Seed is a potential resource to be crawled
// It is most likely a URL, but could also be a different format (ex. Gemini)
type Seed interface {
	Crawl(tier int, wg *sync.WaitGroup) (CrawlResult, error)
}
