package fragments

import (
	"strconv"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Fragment ...
type Fragment struct {
	src     string
	timeout int64
	method  string

	once sync.Once
	s    *goquery.Selection
}

// FromSelection ...
func FromSelection(s *goquery.Selection) *Fragment {
	f := new(Fragment)

	f.s = s

	src, _ := s.Attr("src")
	f.src = src

	method, _ := s.Attr("method")
	f.method = method

	timeout, ok := s.Attr("timeout")
	if !ok {
		timeout = "60"
	}
	t, _ := strconv.ParseInt(timeout, 10, 64)
	f.timeout = t

	return f
}

// Src ...
func (f *Fragment) Src() string {
	return f.src
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
