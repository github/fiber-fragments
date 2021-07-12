// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package fragments

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"golang.org/x/sync/errgroup"
)

var client = fasthttp.Client{
	NoDefaultUserAgentHeader: true,
	DisablePathNormalizing:   true,
}

// Config ...
type Config struct {
	// Filter defines a function to skip the middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// ErrorHandler defines a function which is executed
	// It may be used to define a custom error.
	// Optional. Default: 401 Invalid or expired key
	ErrorHandler fiber.ErrorHandler

	// DefaultHost defines the host to use,
	// if no host is set on a fragment.
	// Optional. Default: localhost:3000
	DefaultHost string
}

// Template ...
func Template(config Config, name string, bind interface{}, layouts ...string) fiber.Handler {
	// Set default config
	cfg := configDefault(config)

	return func(c *fiber.Ctx) error {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		var err error
		var buf *bytes.Buffer = new(bytes.Buffer)

		if c.App().Config().Views != nil {
			// Render template based on global layout if exists
			if len(layouts) == 0 && c.App().Config().ViewsLayout != "" {
				layouts = []string{
					c.App().Config().ViewsLayout,
				}
			}
			// Render template from Views
			if err := c.App().Config().Views.Render(buf, name, bind, layouts...); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		} else {
			// Render raw template using 'name' as filepath if no engine is set
			var tmpl *template.Template
			if _, err = readContent(buf, name); err != nil {
				return cfg.ErrorHandler(c, err)
			}
			// Parse template
			if tmpl, err = template.New("").Parse(string(buf.Bytes())); err != nil {
				return cfg.ErrorHandler(c, err)
			}
			buf.Reset()
			// Render template
			if err = tmpl.Execute(buf, bind); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		}

		r := bytes.NewReader(buf.Bytes())
		doc, err := NewDocument(r)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		return Do(c, cfg, doc)
	}
}

// Do represents the core functionality of the middleware.
// It resolves the fragments from a parsed template.
func Do(c *fiber.Ctx, cfg Config, doc *Document) error {
	g, _ := errgroup.WithContext(c.Context())

	ff, err := doc.Fragments()
	if err != nil {
		return err
	}

	for _, f := range ff {
		f := f

		g.Go(func() error {
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
			if err := client.Do(req, res); err != nil {
				return err
			}

			t := f.Timeout()
			if err := client.DoTimeout(req, res, t); err != nil {
				return err
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

	// this is sync, we wait for everything to resolve
	if err := g.Wait(); err != nil {
		return cfg.ErrorHandler(c, err)
	}

	// render the final output
	html, err := doc.Document().Html()
	if err != nil {
		return cfg.ErrorHandler(c, err)
	}

	c.Response().Header.SetContentType(fiber.MIMETextHTMLCharsetUTF8)
	c.Response().SetBody([]byte(html))

	return nil
}

// readContent opens a named file and read content from it
func readContent(rf io.ReaderFrom, name string) (n int64, err error) {
	// Read file
	f, err := os.Open(filepath.Clean(name))
	if err != nil {
		return 0, err
	}
	defer func() {
		err = f.Close()
	}()
	return rf.ReadFrom(f)
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).SendString("cannot create response")
		}
	}

	if cfg.DefaultHost == "" {
		cfg.DefaultHost = "localhost:3000"
	}

	return cfg
}
