# goscrape

goscrape is a web crawling/scraping framework for the Go language loosly inspired by Scrapy for Python. goscrape was written for a few reasons;

* To automate tasks where large amounts of HTTP content needs to be downloaded and processed;
* To allow developers to produce a single statically linked Go binary for crawling tasks;
* To define a spidering tasks in terms of configuring a struct, rather than writing code for every single task involved in crawling;
* ... but mainly because I was bored ;)

See the examples directory for some runnable code examples.

## Current Status

goscrape is currently very alpha. There's a lot that I still want to do with it, but for my very small scale tests it seems to work as expected. That said, be warned that it's probably very buggy still.

Any contributions greatfully received :)

## TODO List

* Unit tests (at some point)
* More sophisticated examples
* ~~~Add a cookie store~~~
    * Changed the http.Client that's used by the spider to public so theoretically you could just use the http.Client's API to do this.
* Add more URL validation (such as automatic expansions for relative URLS)
* Add some more ready-to-go middleware functions
* Add some ready-to-go parse functions (such as get all hrefs from a page, for example)
* Clean up the code (there's places where some code duplication has occured)
    * Currently need to do some more documentation to get it through go lint
* Add a clear redirect policy. Currently 301 redirects are automatically allowed, but this may not suit all scenarios. This will require syncing up the http.Client with the getAndVerifyHead function.
* Sprinkle a little more awesome to take it from a toy to deployable tool
