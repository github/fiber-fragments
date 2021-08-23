package main

import (
	"time"

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
		c.Links("https://unpkg.com/tailwindcss@^2/dist/tailwind.min.css", "stylesheet")

		return c.Render("fragment1", fiber.Map{
			"Title": "Example 1",
		})
	})

	app.Get("/fragment2", func(c *fiber.Ctx) error {
		c.Links("https://unpkg.com/react-dom@17/umd/react-dom.development.js", "script", "")

		c.Response().SetStatusCode(403)

		return c.Render("fragment2", fiber.Map{
			"Title": "Example 2",
		})
	})

	app.Get("/fragment3", func(c *fiber.Ctx) error {
		timer1 := time.NewTimer(90 * time.Second)

		<-timer1.C // wait here for fallback

		return c.Render("fragment3", fiber.Map{
			"Title": "Example 3",
		})
	})

	app.Get("/fragment4", func(c *fiber.Ctx) error {
		return c.Render("fragment4", fiber.Map{
			"Title": "Example 4",
		})
	})

	app.Get("/fragment5", func(c *fiber.Ctx) error {
		return c.Render("fragment5", fiber.Map{
			"Title": "Example 5",
		})
	})

	app.Get("/fragment6", func(c *fiber.Ctx) error {
		return c.Render("fragment6", fiber.Map{
			"Title": "Example 6",
		})
	})

	app.Get("/fallback", func(c *fiber.Ctx) error {
		return c.Render("fallback", fiber.Map{
			"Title": "Fallback",
		})
	})

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}
