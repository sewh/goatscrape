// This example starts with a few seed urls,
// and includes a parse function that always returns
// the same URL. This demonstrates the built in
// URL sanitisers offered by the framework.
package main

import (
	"net/http"

	"github.com/stevie-holdway/goscrape"
	"github.com/stevie-holdway/goscrape/plugins"
)

func main() {
	example := goscrape.Spider{
		Name: "Example 1",
		StartingURLs: []string{
			"http://www.xkcd.com/",
			"http://www.xkcd.com/1/",
			"http://www.xkcd.com/2/",
		},
		AllowedDomains: []string{
			"www.xkcd.com",
		},
		MaxPages:              10,
		MaxConcurrentRequests: 5,
		Parse: func(req *http.Response) []string {
			m := make([]string, 0)
			m = append(m, "http://www.xkcd.com/3/")
			return m
		},
		Links:   &plugins.BasicLinkStore{},
		Verbose: true,
	}
	example.AddPreRequestMiddleware(plugins.RandomiseUserAgent)

	example.Start()
}
