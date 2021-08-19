// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package fragments

import (
	"bytes"
	"html/template"
	"io"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
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

	// FilterResponse defines a function to filter the responses
	// from the fragment sources.
	FilterResponse func(*fasthttp.Response) *fasthttp.Response

	// FilterRequest defines a function to filter the request
	// to the fragment sources.
	FilterRequest func(*fasthttp.Request) *fasthttp.Request

	// ErrorHandler defines a function which is executed
	// It may be used to define a custom error.
	// Optional. Default: 401 Invalid or expired key
	ErrorHandler fiber.ErrorHandler

	// FilterHead defines a function to filter the new
	// nodes in the <head> of the document passed by the LINK header entity.
	FilterHead func([]*html.Node) []*html.Node

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
			if tmpl, err = template.New("").Parse(buf.String()); err != nil {
				return cfg.ErrorHandler(c, err)
			}
			buf.Reset()
			// Render template
			if err = tmpl.Execute(buf, bind); err != nil {
				return cfg.ErrorHandler(c, err)
			}
		}

		r := bytes.NewReader(buf.Bytes())

		root := &html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Html,
			Data:     "html",
		}

		doc, err := NewDocument(r, root)
		if err != nil {
			return cfg.ErrorHandler(c, err)
		}

		return Do(c, cfg, doc)
	}
}

// Do represents the core functionality of the middleware.
// It resolves the fragments from a parsed template.
func Do(c *fiber.Ctx, cfg Config, doc *Document) error {
	r := NewResolver()
	statusCode, head, err := r.Resolve(c, cfg, doc)
	if err != nil {
		return err
	}

	// append all head nodes
	doc.AppendHead(cfg.FilterHead(head)...)

	// get final output
	html, err := doc.Html()
	if err != nil {
		return cfg.ErrorHandler(c, err)
	}

	c.Response().SetStatusCode(statusCode)
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

	if cfg.FilterResponse == nil {
		cfg.FilterResponse = func(res *fasthttp.Response) *fasthttp.Response {
			return res
		}
	}

	if cfg.FilterRequest == nil {
		cfg.FilterRequest = func(req *fasthttp.Request) *fasthttp.Request {
			return req
		}
	}

	if cfg.DefaultHost == "" {
		cfg.DefaultHost = "localhost:3000"
	}

	if cfg.FilterHead == nil {
		cfg.FilterHead = func(nodes []*html.Node) []*html.Node {
			return nodes
		}
	}

	return cfg
}
