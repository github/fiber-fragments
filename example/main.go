package main

import (
	"github.com/gofiber/fiber/v2"

	fragments "github.com/github/fiber-fragments"
)

func main() {
	app := fiber.New()

	fragments := fragments.New(fragments.Config{})

	app.Get("/index", fragments)

	app.Listen(":8080")
}
