package fragments

import (
	"io"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/net/html"
)

// Document ...
type Document struct {
	doc        *goquery.Document
	html       *HtmlFragment
	statusCode int

	sync.RWMutex
}

// NewDocument ...
func NewDocument(r io.Reader, root *html.Node) (*Document, error) {
	d := new(Document)
	// set the default status code
	d.statusCode = fiber.StatusOK

	html, err := NewHtmlFragment(r, root)
	if err != nil {
		return nil, err
	}
	d.html = html

	return d, nil
}

// Document ...
func (d *Document) Document() *goquery.Document {
	return d.doc
}

// Html is returning the final HTML output.
func (d *Document) Html() (string, error) {
	d.RLock()
	defer d.RUnlock()

	html, err := d.html.Html()
	if err != nil {
		return "", err
	}

	return html, nil
}

// Fragments is returning the selection of fragments
// from an HTML page.
func (d *Document) Fragments() ([]*Fragment, error) {
	d.RLock()
	defer d.RUnlock()

	scripts := d.doc.Find("head script[type=fragment]")
	fragments := d.doc.Find("fragment").AddSelection(scripts)

	ff := make([]*Fragment, 0, fragments.Length())

	fragments.Each(func(i int, s *goquery.Selection) {
		f := FromSelection(s)

		if !f.deferred {
			ff = append(ff, f)
		}
	})

	return ff, nil
}

// Fragments is returning the selection of fragments
// from an HTML page.
func (d *Document) HtmlFragment() *HtmlFragment {
	d.RLock()
	defer d.RUnlock()

	return d.html
}

// SetStatusCode is setting the HTTP status code for the document.
func (d *Document) SetStatusCode(status int) {
	d.Lock()
	defer d.Unlock() // could do this atomic

	d.statusCode = status
}

// StatusCode is getting the HTTP status code for the document.
func (d *Document) StatusCode() int {
	d.RLock()
	defer d.RUnlock()

	return d.statusCode
}

// AppendHead ...
// func (d *Document) AppendHead(ns ...*html.Node) {
// 	head := d.doc.Find("head")
// 	head.AppendNodes(ns...)
// }
