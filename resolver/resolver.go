package resolver

import (
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"
	o "github.com/andersnormal/pkg/opts"
	"github.com/github/fiber-fragments/document"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
)

// Resolver ...
type Resolver interface {
	WithContext(context.Context, document.Document, *fasthttp.Request) error
}

// ResolverFunc ...
type ResolverFunc = func(*goquery.Document) func() error

// New ...
func New() Resolver {
	return &resolver{}
}

type resolver struct {
	opts o.Opts
}

// WithContext ...
func (r *resolver) WithContext(ctx context.Context, doc document.Document, req *fasthttp.Request) error {
	ff, err := doc.Fragments()
	if err != nil {
		return err
	}

	g, gctx := errgroup.WithContext(ctx)

	for _, f := range ff {
		f := f

		g.Go(func() error {
			tctx, cancel := context.WithTimeout(gctx, time.Second*f.Timeout())
			defer cancel()

			url, err := url.Parse(f.Src())
			if err != nil {
				return err
			}
			url.Scheme = "http"
			url.Host = "127.0.0.1:8080"

			rr, err := http.NewRequestWithContext(tctx, f.Method(), url.String(), nil)
			if err != nil {
				return err
			}

			if clientIP, _, err := net.SplitHostPort(rr.RemoteAddr); err == nil {
				appendHostToXForwardHeader(req.Header, clientIP)
			}

			copyHeader(req.Header, rr.Header)

			c, err := r.opts.Get(Client)
			if err != nil {
				return err
			}
			client := c.(*http.Client)

			res, err := client.Do(rr)
			if err != nil {
				return err
			}

			var reader io.ReadCloser
			switch res.Header.Get("Content-Encoding") {
			case "gzip":
				reader, err = gzip.NewReader(res.Body)
				defer reader.Close()
			default:
				reader = res.Body
			}

			body, err := ioutil.ReadAll(reader)
			if err != nil {
				return err
			}

			if _, err := io.Copy(io.Discard, res.Body); err != nil {
				return err
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				return nil
			}

			l := header.Header(res.Header.Get("Link"))
			nodes := header.CreateNodes(l.Links())
			doc.AppendHead(nodes...)

			// do not replace when not resolved
			f.Element().ReplaceWithHtml(string(body))

			return nil
		})
	}

	// this is sync for now
	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
