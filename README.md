# goatscrape

goatscrape is a web crawling/scraping framework for the Go language loosly inspired by Scrapy for Python. It favours composibility, and has the majority of its functionality seperated into plugins; making it easy to compose behaviour from default plugins or write your own. goatscrape was written for a few reasons;

* To automate tasks where large amounts of HTTP content needs to be downloaded and processed;
* To allow developers to produce a single statically linked Go binary for crawling tasks;
* To define a spidering tasks in terms of configuring a struct, rather than writing code for every single task involved in crawling;
* ... but mainly because I was bored ;)

See the examples directory for some runnable code examples.

goatscrape was originally called 'goscrape' but it was altered when there were a few other projects with that name. Despite popular belief, it only scrapes goats if there is some kind of goat oriented website to crawl.

## Current Status

goatscrape is currently very alpha. There's a lot that I still want to do with it, but for my very small scale tests it seems to work as expected. That said, be warned that it's probably very buggy still.

Any contributions greatfully received :)

## TODO List

* A walkthrough guide and tutorial with some examples. Will do this when the API looks like it is pretty stable.
* Unit tests (at some point)
* More sophisticated code examples
* ~~Add a cookie store~~
    * ~~Changed the http.Client that's used by the spider to public so theoretically you could just use the http.Client's API to do this.~~
        * Added getter plugins. Will add a getter plugin that takes a user supplied http.Client in the future.
* Add some more ready-to-go middleware functions
* Add some ready-to-go parse functions (such as get all hrefs from a page, for example)
    * Already added one of these, maybe a few others will be added when I think of them
* Add a clear redirect policy. Currently 301 redirects are automatically allowed, but this may not suit all scenarios. This will require syncing up the http.Client with the getAndVerifyHead function.
* Sprinkle a little more awesome to take it from a toy to deployable tool
