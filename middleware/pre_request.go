package middleware

import (
	"math/rand"
	"net/http"
)

// RandomiseUserAgent fullfils the goscrape.PreRequestFunc type. It takes
// a http.Request and applies a random user agent to the 'User-Agent' header.
func RandomiseUserAgent(req *http.Request) {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 6.3; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/37.0.2049.0 Safari/537.36",                                                                                                                                     // Chrome 37
		"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/36.0.1985.67 Safari/537.36",                                                                                                                                                // Chrome 36
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/35.0.1916.47 Safari/537.36",                                                                                                                              // Chrome 35
		"Mozilla/5.0 (Windows NT 5.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/34.0.1847.116 Safari/537.36 Mozilla/5.0 (iPad; U; CPU OS 3_2 like Mac OS X; en-us) AppleWebKit/531.21.10 (KHTML, like Gecko) Version/4.0.4 Mobile/7B334b Safari/531.21.10", // Chrome 34
		"Mozilla/5.0 (Windows NT 5.1; rv:31.0) Gecko/20100101 Firefox/31.0",                                                                                                                                                                                    // Firefox 31
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:25.0) Gecko/20100101 Firefox/29.0",                                                                                                                                                                        // Firefox 29
		"Mozilla/5.0 (X11; Linux x86_64; rv:28.0) Gecko/20100101 Firefox/28.0",                                                                                                                                                                                 // Firefox 28
		"Mozilla/5.0 (Windows NT 6.2; Win64; x64; rv:27.0) Gecko/20121011 Firefox/27.0",                                                                                                                                                                        // Firefox 27
		"Mozilla/5.0 (X11; Linux x86_64; rv:17.0) Gecko/20121202 Firefox/17.0 Iceweasel/17.0.1",                                                                                                                                                                // Iceweasel 17
		"Mozilla/5.0 (compatible; MSIE 10.0; Windows NT 6.1; WOW64; Trident/6.0)",                                                                                                                                                                              // IE 10
		"Mozilla/5.0 (Windows; U; MSIE 9.0; Windows NT 9.0; en-US))",                                                                                                                                                                                           // IE 9
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_6_8) AppleWebKit/537.13+ (KHTML, like Gecko) Version/5.1.7 Safari/534.57.2",                                                                                                                                 // Safari 5
	}

	userAgent := userAgents[rand.Intn(len(userAgents))]

	req.Header.Add("User-Agent", userAgent)
}
