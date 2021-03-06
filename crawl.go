// Package goatscrape is a web crawling and scraping framework. Its aim is to create a robust,
// powerful, crawling framework out of the box that packages a lot of default behaviour into plugins.
// It has the following advantages:
//
// 	- It is easy to use with the default plugins, but can be extended by those who need extra power or control;
// 	- It is performant, natively using concurrency and allowing spider tasks to be compiled to a single binary;
// 	- It can be used in multiple different use cases. It can be tailored to fit a range of tasks from basic screen
// 	  scraping to a bespoke tool that pulls tasks of a work queue and publishes its findings to a database.
//
// It was originally written by Steven Holdway and is released under the MIT License for ease of static linking.
package goatscrape

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"regexp"
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

// RequestFunc should take a pre-constructed http.Request object and
// return a http.Response object or an error. This is used for custom
// getters within the spider object.
type RequestFunc func(*http.Request) (*http.Response, error)

// Spider defines a single scrape job. Clients should create a new
// Spider instance and customise it before running the Start() method.
type Spider struct {
	// Name is the unique scan name. It is currently use for logging purposes.
	Name string
	// StartingURLs is a string slice of all the URLs that will be loaded into
	// the spider first. These should be used to seed the scanner.
	StartingURLs []string
	// AllowedDomains is a string slice with all the allowed domains. An empty
	// slice will cause the spider to assume that there are no domains that are not allowed.
	AllowedDomains []string
	// DisallowedPages is a slice of regular expressions. Each expression is evaluated on all links
	// returned from the Parse() function. If the expression matches then the link is not added to the
	// to crawl list.
	DisallowedPages []regexp.Regexp
	// MaxPages is the maximum amount of pages to crawl before the scanner returns. A setting of zero or less
	// causes the spider to assume there are no maximum pages.
	MaxPages int
	// MaxConcurrentRequests is the maximum amount of requests to run in parallel.
	MaxConcurrentRequests int

	// The Parse function should emit a list of urls that should be added to the crawl.
	Parse ParseFunc
	// PreRequestMiddleware is a slice of functions that implement PreRequestFunc. Each of these functions
	// is called on the http.Request object before it is execute by the http.Client.
	PreRequestMiddleware []PreRequestFunc
	// The function that gets a web page. Should take a http.Request and return a http.Response
	Getter RequestFunc

	// Verbose will cause more diagnostic information to be outputted if it's set to true.
	Verbose bool
	// Quiet will repress all output to stdout or stderr
	Quiet bool

	// Links is the LinkStore object that is used by this spider. LinkStores are responsible for storing,
	// and managing the crawled and the to crawl lists used by the spider during its operation.
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
	if !s.Quiet {
		log.Println("[" + s.Name + "] Starting Spider")
	}
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

	if s.Getter == nil {
		return errors.New("Spider must have a getter.")
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

	if !s.Quiet {
		log.Println("[" + s.Name + "] has completed.")
	}
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
	defer func() {
		s.wg.Done() // Make sure we mark this is done at the end of the function.
	}()
	if err != nil {
		if !s.Quiet && s.Verbose {
			log.Println("[" + s.Name + "] " + err.Error())
		}
		return
	}

	// The page is fine, we can now crawl it.
	req, _ := http.NewRequest("GET", uri, nil)
	s.processRequestMiddleware(req)

	resp, err := s.Getter(req)
	if err != nil {
		return
	}

	if !s.Quiet {
		log.Println("[" + s.Name + "] Spidered " + uri)
	}
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
