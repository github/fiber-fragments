package fragments

import (
	"io"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

type Document struct {
	doc *goquery.Document
}

// NewDocument ...
func NewDocument(r io.Reader) (*Document, error) {
	d := new(Document)

	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}

	d.doc = doc

	return d, nil
}

// Document ...
func (d *Document) Document() *goquery.Document {
	return d.doc
}

// Fragments ...
func (d *Document) Fragments() ([]*Fragment, error) {
	scripts := d.doc.Find("head script[type=fragment]")
	fragments := d.doc.Find("fragment").AddSelection(scripts)

	ff := make([]*Fragment, 0, fragments.Length())

	fragments.Each(func(i int, s *goquery.Selection) {
		f := FromSelection(s)

		ff = append(ff, f)
	})

	return ff, nil
}

// AppendHead ...
func (d *Document) AppendHead(ns ...*html.Node) {
	head := d.doc.Find("head")
	head.AppendNodes(ns...)
}
