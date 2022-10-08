RoxRouter
=========

RoxRouter is based off of the pattern matching algorithm of [apirouter](https://godoc.org/github.com/cnotch/apirouter)
married to [fasthttp](https://github.com/valyala/fasthttp). It should be very fast minimal server and should make a great staring point for you next project once we come out of alpha.

### Note
This is expirmental software and as such *should not* be used in Production as yet.

### Project Status
Alpha

### Example

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
		TLS: rox.RxTLS{
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
