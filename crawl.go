// Package goscrape is a web crawling and scraping
// framework inspired by Scrapy, and built for my own entertainment.
// The heart of the package is the Spider structure.
package goscrape

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"
)

// ParseFunc defines a function that takes a HTTP response,
// and returns a string slice of further URLs to crawl.
type ParseFunc func(*http.Response) []string

// PreRequestFunc is a function that modifies an existing
// http.Request object before it is made to a web server. It
// can be used for, as an example, modifying the user agent
// header before each request.
type PreRequestFunc func(*http.Request)

// Spider defines a single scrape job. Clients should create a new
// Spider instance and customise it before running the Start() method.
type Spider struct {
	Name                  string
	StartingURLs          []string
	AllowedDomains        []string
	DisallowedPages       []regexp.Regexp
	MaxPages              int
	MaxConcurrentRequests int

	// The Parse function should emit a list of urls
	// that should be added to the crawl.
	Parse                ParseFunc
	PreRequestMiddleware []PreRequestFunc

	Verbose bool

	Client http.Client

	// Unexported fields
	lists struct {
		crawled []string
		toCrawl []string
		lock    sync.Mutex
	}

	hasAllowedDomains       bool
	hasMaxPages             bool
	hasPreRequestMiddleware bool
	hasParse                bool

	totalSpidered int

	wg sync.WaitGroup
}

// AddPreRequestMiddleware takes a veradic amount of PreRequestFunc
// arguments, and make sure each of the functions added are called
// on the http.Request object before a request is made.
func (s *Spider) AddPreRequestMiddleware(funcs ...PreRequestFunc) {
	s.hasPreRequestMiddleware = true
	for _, f := range funcs {
		s.PreRequestMiddleware = append(s.PreRequestMiddleware, f)
	}
}

// Start begins the job with the settings defined in the spider
// structure's configuration.
func (s *Spider) Start() (err error) {
	err = s.validateSettings()
	if err != nil {
		log.Fatal("[goscrape] " + err.Error())
	}

	s.loadStartingURLS()
	log.Println("[" + s.Name + "] Starting Spider")
	s.crawlLoop()

	return nil
}

func (s *Spider) processRequestMiddleware(req *http.Request) {
	if !s.hasPreRequestMiddleware {
		return
	}
	for _, m := range s.PreRequestMiddleware {
		m(req)
	}
}

func (s *Spider) validateSettings() error {
	if s.Name == "" {
		return errors.New("Crawls must have a name.")
	}

	if len(s.StartingURLs) == 0 {
		return errors.New("Crawl must have starting URLs.")
	}

	if s.Parse != nil {
		s.hasParse = true
	} else {
		s.hasParse = false
	}

	if len(s.PreRequestMiddleware) > 0 {
		s.hasPreRequestMiddleware = true
	} else {
		s.hasPreRequestMiddleware = false
	}

	if s.MaxPages <= 0 {
		s.hasMaxPages = false
	} else {
		s.hasMaxPages = true
	}

	if len(s.AllowedDomains) == 0 {
		s.hasAllowedDomains = false
	} else {
		s.hasAllowedDomains = true
	}

	if s.MaxConcurrentRequests <= 0 {
		s.MaxConcurrentRequests = 1
	}

	s.totalSpidered = 0

	return nil
}

func (s *Spider) crawlLoop() error {
	for len(s.lists.toCrawl) != 0 {
		var temp []string

		if s.hasMaxPages {
			if s.totalSpidered >= s.MaxPages {
				return nil
			}
		}
		// Load in the URLS to be gotten
		counter := 0
		for _, elem := range s.lists.toCrawl {
			if counter == s.MaxConcurrentRequests {
				break
			}
			if s.hasMaxPages {
				if (s.totalSpidered + counter) >= s.MaxPages {
					break
				}
			}
			counter++

			temp = append(temp, elem)
		}

		s.wg.Add(counter)

		// Crawl each page, and call the parse function
		for _, uri := range temp {
			// Abort if we have spidered too many pages
			if s.hasMaxPages {
				if s.totalSpidered >= s.MaxPages {
					continue
				}
			}
			// Move the current url to the crawled list
			s.lists.lock.Lock()
			s.lists.crawled = append(s.lists.crawled, uri)
			s.deleteFromToCrawl(uri)
			s.lists.lock.Unlock()

			// Begin the crawl part
			go s.getPage(uri)
			s.totalSpidered++
		}
		s.wg.Wait()
	}
	return nil
}

func (s *Spider) loadStartingURLS() {
	for _, url := range s.StartingURLs {
		s.lists.toCrawl = append(s.lists.toCrawl, url)
	}
}

func (s *Spider) deleteFromToCrawl(url string) {
	for i, e := range s.lists.toCrawl {
		if e == url {
			// Thank the good baby Jesus for SliceTricks...
			s.lists.toCrawl = append(s.lists.toCrawl[:i], s.lists.toCrawl[i+1:]...)
		}
	}
}

func (s *Spider) getPage(uri string) {
	err := s.verifyURL(uri)
	defer func() {
		s.wg.Done()
	}()
	if err != nil && s.Verbose {
		log.Println("[" + s.Name + "] " + err.Error())
		return
	}

	req, _ := http.NewRequest("GET", uri, nil)
	s.processRequestMiddleware(req)

	resp, err := s.Client.Do(req)
	log.Println("[" + s.Name + "] Spidered " + uri)
	if s.hasParse {
		links := s.Parse(resp)
		s.lists.lock.Lock()

		// Add the parsed links to the list, provided
		// it's a valid link
		for _, l := range links {
			err = s.verifyURL(l)
			err2 := s.isPageDisallowed(l)
			if err != nil || err2 != nil {
				continue
			}
			if !s.doesLinkExist(l) {
				s.lists.toCrawl = append(s.lists.toCrawl, l)
			}
		}
		s.lists.lock.Unlock()
	}

}

func (s *Spider) doesLinkExist(uri string) bool {
	for _, e := range s.lists.toCrawl {
		if e == uri {
			return true
		}
	}

	for _, e := range s.lists.crawled {
		if e == uri {
			return true
		}
	}

	return false
}

func (s *Spider) verifyURL(uri string) error {
	u, err := url.Parse(uri)
	if err != nil {
		return err
	}

	if !u.IsAbs() {
		return errors.New(uri + " not an absolute URL.")
	}

	shouldContinue := false
	for _, e := range s.AllowedDomains {
		if e == u.Host {
			shouldContinue = true
		}
	}
	if !shouldContinue {
		return errors.New(uri + " not listed as allowed in spider settings.")
	}

	return nil
}

func (s *Spider) isPageDisallowed(uri string) error {
	for _, r := range s.DisallowedPages {
		if len(r.FindAllString(uri, 1)) > 0 {
			return errors.New(uri + " is disallowed.")
		}
	}

	return nil
}
