package fragments

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Fragment is a <fragment> in the <header> or <body>
// of a HTML page.
type Fragment struct {
	deferred bool
	fallback string
	method   string
	primary  bool
	src      string
	timeout  int64

	once sync.Once
	s    *goquery.Selection
}

// FromSelection creates a new fragment from a
// fragment selection in the DOM.
func FromSelection(s *goquery.Selection) *Fragment {
	f := new(Fragment)
	f.s = s

	src, _ := s.Attr("src")
	f.src = src

	fallback, _ := s.Attr("fallback")
	f.fallback = fallback

	method, _ := s.Attr("method")
	f.method = method

	timeout, ok := s.Attr("timeout")
	if !ok {
		timeout = "60"
	}
	t, _ := strconv.ParseInt(timeout, 10, 64)
	f.timeout = t

	deferred, ok := s.Attr("deferred")
	f.deferred = ok && strings.ToUpper(deferred) != "FALSE"

	primary, ok := s.Attr("primary")
	f.primary = ok && strings.ToUpper(primary) != "FALSE"

	return f
}

// Src is the URL for the fragment.
func (f *Fragment) Src() string {
	return f.src
}

// Fallback is the fallback URL for the fragment.
func (f *Fragment) Fallback() string {
	return f.fallback
}

// Timeout is the timeout for fetching the fragment.
func (f *Fragment) Timeout() time.Duration {
	return time.Duration(f.timeout)
}

// Method is the HTTP method to use for fetching the fragment.
func (f *Fragment) Method() string {
	return f.method
}

// Element is a pointer to the selected element in the DOM.
func (f *Fragment) Element() *goquery.Selection {
	return f.s
}

// Deferred is deferring the fetching to the browser.
func (f *Fragment) Deferred() bool {
	return f.deferred
}

// Primary denotes a fragment as responsible for setting
// the response code of the entire HTML page.
func (f *Fragment) Primary() bool {
	return f.primary
}
