package fragments

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

// Resolver ...
type Resolver struct {
	cancel func()
	wg     sync.WaitGroup
}

// Resolve blocks until all fragments have been called.
func (r *Resolver) Resolve(c *fiber.Ctx, doc *Document, cfg *Config) error {
	ff, err := doc.Fragments()
	if err != nil {
		r.cancel()

		return err
	}

	for _, f := range ff {
		r.run(func() error {
			req := fasthttp.AcquireRequest()
			res := fasthttp.AcquireResponse()

			defer fasthttp.ReleaseRequest(req)
			defer fasthttp.ReleaseResponse(res)

			c.Request().CopyTo(req)

			uri := fasthttp.AcquireURI()
			defer fasthttp.ReleaseURI(uri)

			uri.Parse(nil, []byte(f.src))

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

			if f.primary {
				doc.SetStatusCode(res.StatusCode())
			}

			if res.StatusCode() != http.StatusOK {
				// TODO: wrap in custom error
				return fmt.Errorf("resolve: could not resolve fragment at %s", f.Src())
			}

			res.Header.Del(fiber.HeaderConnection)

			contentEncoding := res.Header.Peek("Content-Encoding")
			body := res.Body()
			if bytes.EqualFold(contentEncoding, []byte("gzip")) {
				body, err = res.BodyGunzip()
				if err != nil {
					return cfg.ErrorHandler(c, err)
				}
			}

			h := Header(string(res.Header.Peek("link")))
			nodes := CreateNodes(h.Links())
			doc.AppendHead(nodes...)

			f.Element().ReplaceWithHtml(string(body))

			return nil
		})
	}

	r.wg.Wait()

	return nil
}

func (r *Resolver) run(fn func() error) {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		fn()
	}()
}
