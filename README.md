RoxRouter
=========

RoxRouter is based off of the wonderful pattern matching algorithm of [apirouter](https://godoc.org/github.com/cnotch/apirouter)
married to [fasthttp](https://github.com/valyala/fasthttp). It should be a very fast minimalistic server, and should make a great starting point for your next project once we come out of alpha.

### Why Rox?
- You are looking for a light Router that doesn't try to do everything for you.
- You need minimal dependencies - we include only those of fasthttp!
- You need something robust - Testing built in
- You need flexible static and dynamic routing - Thanks to APIRouter and fasthttp
- You need something fast - Fasthttp is the fastest server for Go, period.

### Project Status
Alpha - not recommended for Production as yet

### Example
(See also rox_test.go)

```go
package main

import (
	"log"
	"github.com/valyala/fasthttp"
	"github.com/rohanthewiz/rox"
)

func main() {
	r := rox.New(rox.Options{
		Verbose: true,
		Port:    "3020",
		TLS: rox.TLSOpts {
			UseTLS:   false,
			CertFile: "/etc/letsencrypt/live/mysite.com/cert.pem",
			KeyFile:  "/etc/letsencrypt/live/mysite.com/privkey.pem",
		},
	})

	var customHdlr fasthttp.RequestHandler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Yo. It's not found")
	}
	r.Options.CustomNotFoundHandler = &customHdlr

	// Logging middleware
	r.Use(
		func(ctx *fasthttp.RequestCtx) (ok bool) {
			log.Printf("MW:: Requested path: %s\n", ctx.Path())
			return true
		},
		fasthttp.StatusServiceUnavailable, // 503
	)

	// Auth middleware
	r.Use(
		func(ctx *fasthttp.RequestCtx) (ok bool) {
			authed := true // pretend we got a good response from our auth check
			if !authed {
				return false
			}
			log.Printf("MW:: You are authorized for: %s\n", ctx.Path())
			return true
		},
		fasthttp.StatusUnauthorized,
	)

	// Add routes for static files
	r.AddStaticFilesRoute("/images/", "artifacts/images", 1)
	r.AddStaticFilesRoute("/css/", "artifacts/css", 1)
	// r.AddStaticFilesRoute("/.well-known/acme-challenge/", "certs", 0) // great for letsEncrypt!

	r.Get("/", func(ctx *fasthttp.RequestCtx, params rox.Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hello there! Rox here.")
	})
	r.Get("/greet/:name", func(ctx *fasthttp.RequestCtx, params rox.Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey " + params.ByName("name") + "!")
	})
	r.Get("/greet/city", func(ctx *fasthttp.RequestCtx, params rox.Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey big city!")
	})
	r.Get("/greet/city/street", func(ctx *fasthttp.RequestCtx, params rox.Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey big city street!")
	})

	r.Serve()
}
```
