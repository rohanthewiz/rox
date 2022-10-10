package rox

import (
	"fmt"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rohanthewiz/rox/core/cors"
	"github.com/valyala/fasthttp"
)

func TestRox(t *testing.T) {
	r := initTestRox()
	fsvr := &fasthttp.Server{ // Prepare the test server once
		Handler: r.prepareServer(),
	}

	type expectedResp struct {
		StatusCodeMin int
		StatusCodeMax int
		Contains      string
	}

	type args struct {
		method, target string
		body           io.Reader
	}
	tests := []struct {
		name    string
		args    args
		resp    expectedResp
		wantErr bool
	}{
		{name: "Root",
			args: args{"GET", "/", nil},
			resp: expectedResp{200, 400, "Hello there! Rox here."},
		},
		{name: "Route with param",
			args: args{"GET", "/greet/sue", nil},
			resp: expectedResp{200, 400, "Hey sue"},
		},
		{name: "Route overlapping route with param",
			args: args{"GET", "/greet/city", nil},
			resp: expectedResp{200, 400, "Hey big city!"},
		},
		{name: "Route overlapping route with param 2",
			args: args{"GET", "/greet/city/street", nil},
			resp: expectedResp{200, 400, "Hey big city street!"},
		},
		{name: "Route with multiple params",
			args: args{"GET", "/student/john/class/Math", nil},
			resp: expectedResp{200, 400, "john takes Math"},
		},
		{name: "Static route CSS",
			args: args{"GET", "/css/sample.css", nil},
			resp: expectedResp{200, 400, "background-color"},
		},
		{name: "Static route Image",
			args: args{"GET", "/images/dove.jpg", nil},
			resp: expectedResp{200, 400, ""},
		},
		{name: "Unknown route",
			args: args{"GET", "/abcd/efg", nil},
			resp: expectedResp{200, 400, "Yo! It's not found"},
		},
	}

	for _, tt := range tests {
		fmt.Println()
		t.Run(tt.name, func(t *testing.T) {
			gotResp, err := TestRunner(r, fsvr, httptest.NewRequest(tt.args.method, tt.args.target, tt.args.body))
			if (err != nil) != tt.wantErr {
				t.Errorf("TestRunner() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if sc := gotResp.StatusCode(); sc < tt.resp.StatusCodeMin && sc >= tt.resp.StatusCodeMax {
				t.Errorf("!! Status code out of range. Expected %d <= sc < %d",
					tt.resp.StatusCodeMin, tt.resp.StatusCodeMax)
			}
			if bd := string(gotResp.Body()); !strings.Contains(bd, tt.resp.Contains) {
				t.Errorf("!! Got body: %s,\nShould contain: %s", bd, tt.resp.Contains)
			}
		})
	}
}

// initTestRox creates a Rox router for testing and initializes it with some routes
func initTestRox() *Rox {
	r := New(Options{
		Verbose: true,
		Port:    "3020",
	})

	var customNotFoundHdlr fasthttp.RequestHandler = func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Yo! It's not found")
	}
	r.Options.CustomNotFoundHandler = &customNotFoundHdlr

	// Logging middleware
	r.Use(
		func(ctx *fasthttp.RequestCtx) (ok bool) {
			fmt.Printf("MW:: Requested path: %s\n", ctx.Path())
			return true
		},
		fasthttp.StatusServiceUnavailable, // return a statusCode if not ok
	)

	// Auth middleware
	r.Use(
		func(ctx *fasthttp.RequestCtx) (ok bool) {
			authed := true // pretend we got a good response from our auth logic
			if !authed {
				return false
			}
			// log.Printf("MW:: You are authorized for: %s\n", ctx.Path())
			return true
		},
		fasthttp.StatusUnauthorized, // return a statusCode if not ok
	)

	// CORS middleware
	r.Use(cors.MidWareCors, fasthttp.StatusNotImplemented)

	// Add routes for static files
	r.AddStaticFilesRoute("/images/", "dist_test/images", 1)
	r.AddStaticFilesRoute("/css/", "dist_test/css", 1)
	// rx.AddStaticFilesRoute("/.well-known/acme-challenge/", "certs", 0) // great for letsEncrypt!

	r.Get("/", func(ctx *fasthttp.RequestCtx, params Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hello there! Rox here.")
	})
	r.Get("/greet/:name", func(ctx *fasthttp.RequestCtx, params Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey " + params.ByName("name") + "!")
	})
	r.Get("/greet/city", func(ctx *fasthttp.RequestCtx, params Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey big city!")
	})
	r.Get("/greet/city/street", func(ctx *fasthttp.RequestCtx, params Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString("Hey big city street!")
	})
	r.Get("/student/:name/class/:className", func(ctx *fasthttp.RequestCtx, params Params) {
		ctx.Response.Header.Add("Content-Type", "text/html")
		_, _ = ctx.WriteString(params.ByName("name") + " takes " + params.ByName("className"))
	})

	return r
}
