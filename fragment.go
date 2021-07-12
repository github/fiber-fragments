package fragments

import (
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Fragment ...
type Fragment struct {
	src      string
	timeout  int64
	method   string
	fallback string
	primary  bool
	deferred bool

	once sync.Once
	s    *goquery.Selection
}

// FromSelection ...
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

// Src ...
func (f *Fragment) Src() string {
	return f.src
}

// Fallback ...
func (f *Fragment) Fallback() string {
	return f.fallback
}

// Timeout ...
func (f *Fragment) Timeout() time.Duration {
	return time.Duration(f.timeout)
}

// Method ...
func (f *Fragment) Method() string {
	return f.method
}

// Element ...
func (f *Fragment) Element() *goquery.Selection {
	return f.s
}

// Deferred ...
func (f *Fragment) Deferred() bool {
	return f.deferred
}

// Primary ...
func (f *Fragment) Primary() bool {
	return f.primary
}
