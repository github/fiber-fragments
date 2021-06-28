package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"

	fragments "github.com/github/fiber-fragments"
)

func main() {
	// Create a new engine
	engine := html.New("./views", ".html")

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	app.Get("/index", fragments.Template(fragments.Config{}, "index", fiber.Map{}, "layouts/main"))

	app.Listen(":8080")
}
