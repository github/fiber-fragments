package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"
)

func main() {
	// Create a new engine
	engine := html.New(".", ".html")

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/fragment1", func(c *fiber.Ctx) error {
		c.Links("https://unpkg.com/react-dom@17/umd/react-dom.development.js", "script")

		return c.Render("fragment1", fiber.Map{
			"Title": "Example 1",
		})
	})

	app.Get("/fragment2", func(c *fiber.Ctx) error {
		c.Links("https://unpkg.com/react-dom@17/umd/react-dom.development.js", "script")

		return c.Render("fragment2", fiber.Map{
			"Title": "Example 2",
		})
	})

	app.Listen(":3000")
}
