package fragments

import (
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
)

// Resolver ...
type Resolver struct {
	wg sync.WaitGroup
}

// NewResolver ...
func NewResolver() *Resolver {
	return &Resolver{}
}

// ResolverFunc ...
type ResolverFunc func(c *fiber.Ctx, cfg Config) error

// Resolve blocks until all fragments have been called.
func (r *Resolver) Resolve(c *fiber.Ctx, cfg Config, doc *Document) error {
	ff, err := doc.Fragments()
	if err != nil {
		return err
	}

	for _, f := range ff {
		r.run(c, cfg, f.ResolveSrc())
	}

	r.wg.Wait()

	return nil
}

func (r *Resolver) run(c *fiber.Ctx, cfg Config, fn ResolverFunc) {
	r.wg.Add(1)

	go func() {
		defer r.wg.Done()

		err := fn(c, cfg)
		if err != nil {
			fmt.Println(err)
		}
	}()
}
