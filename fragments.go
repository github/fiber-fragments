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

// RenderFunc ...
type RenderFunc func(c *fiber.Ctx, out io.Writer) error

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool

	// SuccessHandler defines a function which is executed for a valid key.
	// Optional. Default: nil
	SuccessHandler fiber.Handler

	// ErrorHandler defines a function which is executed for an invalid key.
	// It may be used to define a custom error.
	// Optional. Default: 401 Invalid or expired key
	ErrorHandler fiber.ErrorHandler

	// RenderFunc ...
	RenderFunc RenderFunc
}

func Template(config Config, name string, bind interface{}, layouts ...string) fiber.Handler {
	// // Set default config
	// cfg := configDefault(config)

	return func(c *fiber.Ctx) error {
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
				return err
			}
		} else {
			// Render raw template using 'name' as filepath if no engine is set
			var tmpl *template.Template
			if _, err = readContent(buf, name); err != nil {
				return err
			}
			// Parse template
			if tmpl, err = template.New("").Parse(string(buf.Bytes())); err != nil {
				return err
			}
			buf.Reset()
			// Render template
			if err = tmpl.Execute(buf, bind); err != nil {
				return err
			}
		}

		r := bytes.NewReader(buf.Bytes())
		doc, err := NewDocument(r)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		return Do(c, doc, "localhost:3000")
	}
}

func Do(c *fiber.Ctx, doc *Document, addr string) error {
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

			c.Request().CopyTo(req)

			uri := fasthttp.AcquireURI()
			uri.SetHost(addr)
			uri.SetPath(f.Src())

			req.SetRequestURI(uri.String())

			req.Header.Del(fiber.HeaderConnection)
			if err := client.Do(req, res); err != nil {
				return err
			}

			if err := client.Do(req, res); err != nil {
				return err
			}

			if res.StatusCode() != http.StatusOK {
				return fmt.Errorf("resolve: could not resolve fragment at %s", f.Src())
			}

			res.Header.Del(fiber.HeaderConnection)
			body := res.Body()

			f.Element().ReplaceWithHtml(string(body))

			return nil
		})
	}

	// this is sync, we wait for everything to resolve
	if err := g.Wait(); err != nil {
		return err
	}

	// render the final output
	html, err := doc.Document().Html()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
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

	if cfg.SuccessHandler == nil {
		cfg.SuccessHandler = func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	if cfg.ErrorHandler == nil {
		cfg.ErrorHandler = func(c *fiber.Ctx, err error) error {
			return nil
		}
	}

	return cfg
}
