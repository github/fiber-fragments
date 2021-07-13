# Fragments

![Github Action](https://github.com/github/fiber-fragments/workflows/Test%20%26%20Build/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/github/fiber-fragments)](https://goreportcard.com/report/github.com/github/fiber-fragments)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Taylor Swift](https://img.shields.io/badge/secured%20by-taylor%20swift-brightgreen.svg)](https://twitter.com/SwiftOnSecurity)
[![Volkswagen](https://auchenberg.github.io/volkswagen/volkswargen_ci.svg?v=1)](https://github.com/auchenberg/volkswagen)

Fragments middleware for [Fiber](https://github.com/gofiber/fiber) enables building microservices for the frontend.

A `<fragment>` symbolizes a part of a template that is served by a microservices. The middleware concurrently fetches those parts from the service and replaces it in the template. It supports `GET` and `POST` [HTTP methods](https://developer.mozilla.org/de/docs/Web/HTTP/Methods) to fetcht the content. Related resources like CSS or JavaScript are injected via the [HTTP `LINK` entity header field](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Link). A `<fragment>` can occure in the [`body` element](https://developer.mozilla.org/de/docs/Web/HTML/Element/body) or the [`header` element](https://developer.mozilla.org/de/docs/Web/HTML/Element/header). See [Example](#example) to learn more about using fragments.

[Tailor](https://github.com/zalando/tailor) by Zalando is prior art for this middleware.

## Fragement(s)

A `fragment` will be hybrid-polymorphic (if this is a thing). On the server it is parsed and evaluate by the middleware. 🦄 In the browser it will be a web component that received data from the middleware (this is still work in progress ⚠️).

### Server

* `src` The source to fetch for replacement in the DOM
* `method` can be of `GET` (default) or `POST`.
* `primary` denotes a fragment that sets the response code of the page
* `id` is an optional unique identifier (optional)
* `timeout` timeout of a fragement to receive in milliseconds (default is `300`)
* `deferred` is deferring the fetch to the browser
* `fallback` is the fallback source in case of timeout/error on the current fragment

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

## Benchmark(s)

This is run on a MacBook Pro 16 inch locally. It is the `example` run.

* Parsing a local template with extrapolation with the fragments
* Parsing the fragments
* Doing fragments
* Inlining results and adding `Link` header resources to the output

```bash
echo "GET http://127.0.0.1:8080/index" | vegeta attack -duration=5s -rate 2000 | tee results.bin | vegeta report
  vegeta report -type=json results.bin > metrics.json
  cat results.bin | vegeta plot > plot.html
  cat results.bin | vegeta report -type="hist[0,100ms,200ms,300ms]"

Requests      [total, rate, throughput]         10000, 2000.26, 2000.15
Duration      [total, attack, wait]             5s, 4.999s, 285.172µs
Latencies     [min, mean, 50, 90, 95, 99, max]  183.725µs, 251.517µs, 226.993µs, 310.698µs, 394.601µs, 563.022µs, 1.347ms
Bytes In      [total, mean]                     6240000, 624.00
Bytes Out     [total, mean]                     0, 0.00
Success       [ratio]                           100.00%
Status Codes  [code:count]                      200:10000
Error Set:
Bucket           #      %        Histogram
[0s,     100ms]  10000  100.00%  ###########################################################################
[100ms,  200ms]  0      0.00%
[200ms,  300ms]  0      0.00%
[300ms,  +Inf]   0      0.00%
```
