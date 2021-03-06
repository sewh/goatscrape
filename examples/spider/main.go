// This example loads five pages from XKCD, one page
// at a time. It extracts all the links for each page
// crawled using the plugins.ExtractAllLinks function.
// Since only the XKCD domain is allowed, only
// XKCD pages are crawled. It also uses the
// plugins.RandomiseUserAgent function.
package main

import (
	"regexp"

	"github.com/stevie-holdway/goatscrape"
	"github.com/stevie-holdway/goatscrape/plugins"
)

func main() {
	example := goatscrape.Spider{
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
		Getter:                plugins.BasicGetter,
		Parse:                 plugins.ExtractAllLinks,
		Links:                 &plugins.BasicLinkStore{},
		Verbose:               true,
	}
	example.AddPreRequestMiddleware(plugins.RandomiseUserAgent)

	example.Start()
}
