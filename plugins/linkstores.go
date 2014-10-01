package plugins

import "sync"

type BasicLinkStore struct {
	crawled []string
	toCrawl []string
	sync.Mutex
}

func (b *BasicLinkStore) GetLinks(amount int) []string {
	var links []string
	b.Lock()

	counter := 0
	for _, l := range b.toCrawl {
		if counter > amount {
			break
		}

		links = append(links, l)
		counter++
	}

	b.Unlock()

	return links
}

func (b *BasicLinkStore) AddToCrawl(link string) {
	b.Lock()

	// Does the link already exist in either lists?
	for _, l := range b.toCrawl {
		if link == l {
			b.Unlock()
			return
		}
	}
	for _, l := range b.crawled {
		if link == l {
			b.Unlock()
			return
		}
	}

	b.toCrawl = append(b.toCrawl, link)

	b.Unlock()
}

func (b *BasicLinkStore) MoveToCrawled(link string) {
	b.Lock()

	// Delete the item in the toCrawl list
	canContinue := false
	for i, l := range b.toCrawl {
		if l == link {
			b.toCrawl = append(b.toCrawl[:i], b.toCrawl[i+1:]...)
			canContinue = true
		}
	}

	if canContinue {
		b.crawled = append(b.crawled, link)
	}

	b.Unlock()
}

func (b *BasicLinkStore) MoreToCrawl() bool {
	b.Lock()

	if len(b.toCrawl) > 0 {
		b.Unlock()
		return true
	}

	b.Unlock()

	return false
}
