// Package plugins is a set of default pieces of functionality to use with the
// stevie-holdway/goatscrape package.
package plugins

import (
	"errors"
	"net/http"
	"strings"
)

var client http.Client

func BasicGetter(req *http.Request) (*http.Response, error) {
	// First make a head request to verify if the page is a html page
	reqCopy := *req

	reqCopy.Method = "HEAD"

	headResp, err := client.Do(&reqCopy)
	if err != nil {
		return nil, err
	}

	if !strings.Contains(headResp.Header.Get("Content-Type"), "html") {
		return nil, errors.New("Page does not have a content type of html.")
	}

	// Now we have verified we have a html page, we can actually issue a get request.
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
