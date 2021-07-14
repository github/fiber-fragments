package fragments

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
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

	body       string
	statusCode int
	head       []*html.Node

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

	f.head = make([]*html.Node, 0)

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
	return time.Duration(f.timeout) * time.Second
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

// Links returns the new nodes that go in the head via
// the LINK HTTP header entity.
func (f *Fragment) Links() []*html.Node {
	return f.head
}

// Resolve is resolving all needed data, setting headers
// and the status code.
func (f *Fragment) Resolve() ResolverFunc {
	return func(c *fiber.Ctx, cfg Config) error {
		err := f.do(c, cfg, f.src)
		if err == nil {
			return err
		}

		if err != fasthttp.ErrTimeout {
			return err
		}

		err = f.do(c, cfg, f.fallback)
		if err != nil {
			return err
		}

		return nil
	}
}

func (f *Fragment) do(c *fiber.Ctx, cfg Config, src string) error {
	req := fasthttp.AcquireRequest()
	res := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(res)

	c.Request().CopyTo(req)

	uri := fasthttp.AcquireURI()
	defer fasthttp.ReleaseURI(uri)

	uri.Parse(nil, []byte(src))

	if len(uri.Host()) == 0 {
		uri.SetHost(cfg.DefaultHost)
	}
	req.SetRequestURI(uri.String())
	req.Header.Del(fiber.HeaderConnection)

	t := f.Timeout()
	if err := client.DoTimeout(req, res, t); err != nil {
		return err
	}

	res = cfg.FilterResponse(res)
	f.statusCode = res.StatusCode()

	if res.StatusCode() != http.StatusOK {
		// TODO: wrap in custom error, to not replace
		return fmt.Errorf("resolve: could not resolve fragment at %s", f.Src())
	}

	res.Header.Del(fiber.HeaderConnection)

	contentEncoding := res.Header.Peek("Content-Encoding")
	body := res.Body()

	var err error
	if bytes.EqualFold(contentEncoding, []byte("gzip")) {
		body, err = res.BodyGunzip()
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}
	}

	h := Header(string(res.Header.Peek("link")))
	nodes := CreateNodes(h.Links())
	f.head = append(f.head, nodes...)

	f.s.ReplaceWithHtml(string(body))

	return nil
}
