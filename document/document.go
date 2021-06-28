package document

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"

	"github.com/github/fiber-fragments/fragment"
)

// Document ...
type Document interface {
	Fragments() ([]*fragment.Fragment, error)
	Document() *goquery.Document
	AppendHead(ns ...*html.Node)
}

type document struct {
	doc *goquery.Document
}

// NewDocument ...
func NewDocument(doc *goquery.Document) Document {
	d := new(document)
	d.doc = doc

	return d
}

// Document ...
func (d *document) Document() *goquery.Document {
	return d.doc
}

// Fragments ...
func (d *document) Fragments() ([]*fragment.Fragment, error) {
	scripts := d.doc.Find("head script[type=fragment]")
	fragments := d.doc.Find("fragment").AddSelection(scripts)

	ff := make([]*fragment.Fragment, 0, fragments.Length())

	fragments.Each(func(i int, s *goquery.Selection) {
		f := fragment.FromSelection(s)

		ff = append(ff, f)
	})

	return ff, nil
}

// AppendHead ...
func (d *document) AppendHead(ns ...*html.Node) {
	head := d.doc.Find("head")
	head.AppendNodes(ns...)
}
