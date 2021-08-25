package fragments

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// HtmlFragment is representation of HTML fragments.
type HtmlFragment struct {
	doc *goquery.Document
	sync.RWMutex
}

// NewHtmlFragment creates a new fragment of HTML.
func NewHtmlFragment(r io.Reader, root *html.Node) (*HtmlFragment, error) {
	h := new(HtmlFragment)

	ns, err := html.ParseFragment(r, root)
	if err != nil {
		return nil, err
	}

	for _, n := range ns {
		root.AppendChild(n)
	}
	h.doc = goquery.NewDocumentFromNode(root)

	return h, nil
}

// Document get the full document representation
// of the HTML fragment.
func (h *HtmlFragment) Fragment() *goquery.Document {
	return h.doc
}

// Fragments is returning the selection of fragments
// from an HTML page.
func (h *HtmlFragment) Fragments() (map[string]*Fragment, error) {
	h.RLock()
	defer h.RUnlock()

	scripts := h.doc.Find("head script[type=fragment]")
	fragments := h.doc.Find("fragment").AddSelection(scripts)

	ff := make(map[string]*Fragment)

	fragments.Each(func(i int, s *goquery.Selection) {
		f := FromSelection(s)

		if !f.deferred {
			ff[f.ID()] = f
		}
	})

	return ff, nil
}

// Html creates the HTML output of the created document.
func (h *HtmlFragment) Html() (string, error) {
	h.RLock()
	defer h.RUnlock()

	html, err := h.doc.Html()
	if err != nil {
		return "", err
	}

	return html, nil
}

// AppendHead ...
func (d *HtmlFragment) AppendHead(ns ...*html.Node) {
	head := d.doc.Find("head")
	head.AppendNodes(ns...)
}

// Fragment is a <fragment> in the <header> or <body>
// of a HTML page.
type Fragment struct {
	deferred bool
	fallback string
	method   string
	primary  bool
	src      string
	timeout  int64

	id  string
	ref string

	statusCode int
	head       []*html.Node

	f *HtmlFragment
	s *goquery.Selection
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

	id, ok := s.Attr("id")
	if !ok {
		id = uuid.New().String()
	}
	f.id = id

	ref, _ := s.Attr("ref")
	f.ref = ref

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

// Ref represents the reference to another fragment
func (f *Fragment) Ref() string {
	return f.ref
}

// ID represents a unique id for the fragment
func (f *Fragment) ID() string {
	return f.id
}

// HtmlFragment returns embedded fragments of HTML.
func (f *Fragment) HtmlFragment() *HtmlFragment {
	return f.f
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

	if err := uri.Parse(nil, []byte(src)); err != nil {
		return err
	}

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

	// if res.StatusCode() != http.StatusOK {
	// 	// TODO: wrap in custom error, to not replace
	// 	return fmt.Errorf("resolve: could not resolve fragment at %s", f.Src())
	// }

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

	root := &html.Node{
		Type:     html.ElementNode,
		DataAtom: atom.Body,
		Data:     "body",
	}

	doc, err := NewHtmlFragment(bytes.NewReader(body), root)
	if err != nil {
		return nil
	}
	f.f = doc

	return nil
}
