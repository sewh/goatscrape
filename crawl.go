// Package goatscrape is a web crawling and scraping
// framework inspired by Scrapy, and built for my own entertainment.
// The heart of the package is the Spider structure.
package goatscrape

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

// LinkStore defines the interface that any object that looks
// to store, and manage, the toCrawl and the crawled lists should
// implement. All the methods in this interface are expected to be thread safe.
type LinkStore interface {
	// GetLinks should return a string slice of links to crawl, in the
	// amount defined in the amount paramter.
	GetLinks(amount int) []string
	// AddToCrawl should add the link parameter to the to crawl list.
	AddToCrawl(link string)
	// MoveToCrawled should delete the link in the to crawl list and
	// place it in the crawled list.
	MoveToCrawled(link string)
	// MoreToCrawl returns a boolean value if there are still links
	// in the to crawl list.
	MoreToCrawl() bool
}

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

	Links LinkStore

	hasAllowedDomains       bool
	hasMaxPages             bool
	hasPreRequestMiddleware bool
	hasParse                bool
	hasDisallowed           bool

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
		log.Fatal("[goatscrape] " + err.Error())
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

	if s.Links == nil {
		return errors.New("Spider must have a link store.")
	}

	if len(s.DisallowedPages) > 0 {
		s.hasDisallowed = true
	} else {
		s.hasDisallowed = false
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
	for s.Links.MoreToCrawl() {

		// Exit if we have spidered the maximum amount of pages.
		if s.hasMaxPages {
			if s.totalSpidered >= s.MaxPages {
				return nil
			}
		}

		// Load in the URLS to be gotten into a temporary buffer, never exceeding
		// the amount of
		amountToGet := s.MaxConcurrentRequests
		if s.hasMaxPages && (s.totalSpidered+amountToGet) >= s.MaxPages {
			// Only get the delta from MaxPages take totalSpidered
			amountToGet = s.MaxPages - s.totalSpidered
		}
		temp := s.Links.GetLinks(amountToGet)

		s.wg.Add(len(temp)) // Add the amount of links to the wait group.

		// Crawl each page, and call the parse function
		for _, uri := range temp {
			s.Links.MoveToCrawled(uri)
			go s.getPage(uri)
			s.totalSpidered++
		}
		s.wg.Wait() // Wait for all the pages to be downloaded
	}

	log.Println("[" + s.Name + "] has completed.")
	return nil
}

func (s *Spider) loadStartingURLS() {
	for _, link := range s.StartingURLs {
		s.Links.AddToCrawl(link)
	}
}

func (s *Spider) getPage(uri string) {
	// Make sure the page is okay to have a GET request issued.
	err := s.verifyURL(uri)
	err2 := s.getAndValidateHead(uri)
	defer func() {
		s.wg.Done() // Make sure we mark this is done at the end of the function.
	}()
	if err != nil {
		if s.Verbose {
			log.Println("[" + s.Name + "] " + err.Error())
		}
		return
	} else if err2 != nil && s.Verbose {
		if s.Verbose {
			log.Println("[" + s.Name + "] " + err2.Error())
		}
		return
	}

	// The page is fine, we can now crawl it.
	req, _ := http.NewRequest("GET", uri, nil)
	s.processRequestMiddleware(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return
	}

	log.Println("[" + s.Name + "] Spidered " + uri)
	// Call the user defined parse function if it exists and add all links
	// generated from it to the to crawl list
	if s.hasParse {
		links := s.Parse(resp)

		// Add the parsed links to the list, provided
		// it's a valid link
		for _, l := range links {
			err1 := s.verifyURL(l)
			err2 := s.isPageDisallowed(l)
			if err1 != nil || err2 != nil {
				continue
			}

			s.Links.AddToCrawl(l)
		}
	}

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
	if !s.hasDisallowed {
		return nil
	}
	for _, r := range s.DisallowedPages {
		if len(r.FindAllString(uri, 1)) > 0 {
			return errors.New(uri + " is disallowed.")
		}
	}

	return nil
}

func (s *Spider) getAndValidateHead(uri string) error {
	req, _ := http.NewRequest("HEAD", uri, nil)
	s.processRequestMiddleware(req)

	resp, err := s.Client.Do(req)
	if err != nil {
		return err
	}

	// Is resource okay?
	if !(resp.StatusCode/100 == 2 || resp.StatusCode/100 == 3) {
		return errors.New(uri + " returned non-okay status code " + resp.Status)
	}

	// Is a HTML page?
	if !strings.Contains(resp.Header.Get("Content-Type"), "html") {
		return errors.New(uri + " not a HTML page.")
	}

	return nil
}
