// This example loads five pages from XKCD, one page
// at a time. It extracts all the links for each page
// crawled using the middleware.ExtractAllLinks function.
// Since only the XKCD domain is allowed, only
// XKCD pages are crawled. It also uses the
// middleware.RandomiseUserAgent function.
package main

import (
	"regexp"

	"github.com/stevie-holdway/goscrape"
	"github.com/stevie-holdway/goscrape/middleware"
)

func main() {
	example := goscrape.Spider{
		Name: "Example 1",
		StartingURLs: []string{
			"http://www.xkcd.com/",
		},
		AllowedDomains: []string{
			"www.xkcd.com",
		},
		DisallowedPages:       []regexp.Regexp{*regexp.MustCompile("http://www.xkcd.com/about")},
		MaxPages:              10,
		MaxConcurrentRequests: 1,
		Parse: middleware.ExtractAllLinks,
	}
	example.AddPreRequestMiddleware(middleware.RandomiseUserAgent)

	example.Start()
}
