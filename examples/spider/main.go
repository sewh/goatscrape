// This example starts with a few seed urls,
// and includes a parse function that always returns
// the same URL. This demonstrates the built in
// URL sanitisers offered by the framework.
package main

import (
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
		MaxPages:              5,
		MaxConcurrentRequests: 1,
		Parse: middleware.ExtractAllLinks,
	}
	example.AddPreRequestMiddleware(middleware.RandomiseUserAgent)

	example.Start()
}
