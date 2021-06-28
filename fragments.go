// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package fragments

import (
	"bufio"
	"bytes"
	"io"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/gofiber/fiber/v2"

	"github.com/github/fiber-fragments/document"
	"github.com/github/fiber-fragments/resolver"
)

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

	// Client is a HTTP client to make the requests
	Client *http.Client
}

// New ...
func New(config ...Config) fiber.Handler {
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

	// Return middleware handler
	return func(c *fiber.Ctx) error {
		// Filter request to skop middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			return c.Next()
		}

		var buf bytes.Buffer
		t := bufio.NewWriter(&buf)

		err := cfg.RenderFunc(c, t)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		d, err := goquery.NewDocumentFromReader(&buf)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		doc := document.NewDocument(d)

		r := resolver.New()
		err = r.WithContext(c.Context(), doc, c.Request())
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		html, err := doc.Document().Html()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).SendString(err.Error())
		}

		c.Write([]byte(html))

		return nil
	}
}
