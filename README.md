# Fragments

Fragments middleware for [Fiber](https://github.com/gofiber/fiber) allows to build microservice frontends.

[Tailor](https://github.com/zalando/tailor) by Zalando is prior art.

## Fragement(s)

A `fragment` is hybrid-polymorphic (if this is a thing). On the server it is parsed and evaluate by the middleware. In the browser it is a web component that receives data.

### Server

* `src` The source to fetch for replacement in the DOM
* `method` can be of `GET` (default) or `POST`.
* `id` is an optional unique identifier (optional)
* `timeout` timeout of a fragement to receive in milliseconds (default is `300`)
* `deferred` is deferring the fetch to the browser
* `fallback` is deferring the fetch to the browser if failed (default)

## Example

Import the middleware package this is part of the Fiber web framework

```go
package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html"

	"github.com/github/fiber-fragments"
)
```

After you initiate your Fiber app, you can plugin in the fragments middleware. The middleware draws the templates for the fragments to load from the template engine. Thus it supports using all [template](https://github.com/gofiber/template) engines supported by the Fiber team.

```go
	// Create a new engine
	engine := html.New("./views", ".html")

	// Pass the engine to the Views
	app := fiber.New(fiber.Config{
		Views: engine,
	})

  // Associates the route with a specific template with fragments to render
	app.Get("/index", fragments.Template(fragments.Config{}, "index", fiber.Map{}, "layouts/main"))

  // this would listen to port 8080
	app.Listen(":8080")
}
```

```html
<html>
<head>
    <script type="fragment" src="assets"></script>
</head>
<body>
    <h1>Example</h1>
    <fragment src="fragment1.html"></fragment>
</body>
</html>

```

```bash
echo "GET http://localhost:8080/index" | vegeta attack -duration=5s -rate 1000 | tee results.bin | vegeta report
  vegeta report -type=json results.bin > metrics.json
  cat results.bin | vegeta plot > plot.html
  cat results.bin | vegeta report -type="hist[0,100ms,200ms,300ms]"

Requests      [total, rate, throughput]         5000, 1000.22, 995.30
Duration      [total, attack, wait]             5.024s, 4.999s, 24.696ms
Latencies     [min, mean, 50, 90, 95, 99, max]  24.338ms, 29.048ms, 25.207ms, 28.564ms, 33.894ms, 136.354ms, 148.563ms
Bytes In      [total, mean]                     2860000, 572.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:5000
Error Set:
Bucket           #     %       Histogram
[0s,     100ms]  4864  97.28%  ########################################################################
[100ms,  200ms]  136   2.72%   ##
[200ms,  300ms]  0     0.00%
[300ms,  +Inf]   0     0.00%

```
