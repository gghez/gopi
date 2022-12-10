package html

import (
	"log"

	"github.com/anaskhan96/soup"
)

func Parse(url string) (*soup.Root, error) {
	log.Printf("query: %q", url)

	resp, err := soup.Get(url)
	if err != nil {
		return nil, err
	}

	doc := soup.HTMLParse(resp)
	return &doc, nil
}

func ParseAndFind(url string, args ...string) (*soup.Root, error) {
	doc, err := Parse(url)
	if err != nil {
		return nil, err
	}

	findResult := doc.Find(args...)

	return &findResult, nil
}
