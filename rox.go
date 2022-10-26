package rox

import (
	"bytes"
	"fmt"
	"log"
	"regexp"

	"github.com/valyala/fasthttp"
)

const defaultPort = "8800"
const defaultTLSPort = "443"
const ipAny = "0.0.0.0"

// Rox is our core router
type Rox struct {
	Options         Options
	middlewares     []MiddleWare
	get             tree
	post            tree
	delete          tree
	put             tree
	patch           tree
	head            tree
	connect         tree
	trace           tree
	options         tree
	newPattern      func(string, *[]*regexp.Regexp) (Pattern, error)
	notFoundHandler fasthttp.RequestHandler
}

type Options struct {
	Verbose               bool
	Port                  string
	TLS                   TLSOpts
	assetPaths            []AssetPath
	CustomMasterHandler   *fasthttp.RequestHandler
	CustomNotFoundHandler *fasthttp.RequestHandler
}

type TLSOpts struct {
	UseTLS   bool
	CertFile string
	KeyFile  string
	CertData []byte
	KeyData  []byte
}

// New returns a new Rox which is initialized with
// the given options and default pattern style.
//
// For the syntax of the pattern reference rox.NewPattern.
func New(opts ...Options) *Rox {
	r := &Rox{
		// notFoundHandler: http.NotFoundHandler(),
		newPattern: NewPattern,
	}
	if len(opts) > 0 {
		r.Options = opts[0]
	}
	return r
}

// Serve dispatches the request to the first handler
// which matches to req.Method and req.Path.
func (r *Rox) Serve() {
	mainReqHandler := r.prepareServer()

	if r.Options.Verbose {
		fmt.Println("Rox listening on port:", r.Options.Port)
	}

	if r.Options.TLS.UseTLS && r.Options.TLS.CertFile != "" {
		log.Fatal(fasthttp.ListenAndServeTLS(ipAny+":"+r.Options.Port, r.Options.TLS.CertFile,
			r.Options.TLS.KeyFile, mainReqHandler))
	} else if r.Options.TLS.UseTLS && len(r.Options.TLS.CertData) > 0 {
		log.Fatal(fasthttp.ListenAndServeTLSEmbed(ipAny+":"+r.Options.Port, r.Options.TLS.CertData,
			r.Options.TLS.KeyData, mainReqHandler))
	} else {
		log.Fatal(fasthttp.ListenAndServe(":"+r.Options.Port, mainReqHandler))
	}
}

// prepareServer prepares the routes and main handlers
func (r *Rox) prepareServer() fasthttp.RequestHandler {
	if r.Options.Verbose {
		fmt.Println("Preparing routes...")
	}
	r.initTrees()

	if r.Options.Port == "" {
		if r.Options.TLS.UseTLS {
			r.Options.Port = defaultTLSPort
		} else {
			r.Options.Port = defaultPort
		}
	}

	// Get NotFound handler
	var notFoundHandler fasthttp.RequestHandler
	if r.Options.CustomNotFoundHandler != nil {
		notFoundHandler = *r.Options.CustomNotFoundHandler
	} else {
		notFoundHandler = func(c *fasthttp.RequestCtx) {
			c.SetStatusCode(fasthttp.StatusNotFound)
		}
	}
	r.notFoundHandler = notFoundHandler

	// Get Master handler
	var mainReqHandler fasthttp.RequestHandler
	if r.Options.CustomMasterHandler != nil {
		mainReqHandler = *r.Options.CustomMasterHandler
	} else {
		mainReqHandler = initStdMasterHandler(r)
	}
	return mainReqHandler
}

func initStdMasterHandler(r *Rox) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Middlewares - they modify ctx or fail with the provided code
		for _, mw := range r.middlewares {
			if ok := mw.MidFunc(ctx); !ok {
				ctx.SetStatusCode(mw.FailCode)
				return
			}
		}

		// See if we match a file handler
		fileHander, ok := r.getFSHandler(ctx)
		if ok {
			fileHander(ctx)
			return
		}

		t := r.selectTree(string(ctx.Method()))
		if t != nil {
			var params Params
			path := string(ctx.Path())

			if h := t.StaticMatch(path); h != nil {
				fmt.Println("Direct match:", path)
				h(ctx, params)
				return
			}

			h, patt := t.PatternMatch(path, &params)
			if r.Options.Verbose && h != nil && patt != "" {
				fmt.Println("Pattern match:", path, "->", patt)
				h(ctx, params)
				return
			}

			msg := "Unknown Route (404) " + path
			log.Println(msg)
			r.notFoundHandler(ctx)

			ctx.SetStatusCode(fasthttp.StatusNotFound)

		} else {
			const msg = "Unknown HTTP method"
			log.Println(msg)
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			_, _ = ctx.WriteString(msg)
		}
	}
}

func (r *Rox) initTrees() {
	r.get.Init()
	r.post.Init()
	r.delete.Init()
	r.put.Init()
	r.patch.Init()
	r.head.Init()
	r.connect.Init()
	r.trace.Init()
	r.options.Init()
}

// selectTree returns the tree by the given HTTP method.
func (r *Rox) selectTree(method string) *tree {
	switch method {
	case fasthttp.MethodGet:
		return &r.get
	case fasthttp.MethodPost:
		return &r.post
	case fasthttp.MethodDelete:
		return &r.delete
	case fasthttp.MethodPut:
		return &r.put
	case fasthttp.MethodPatch:
		return &r.patch
	case fasthttp.MethodHead:
		return &r.head
	case fasthttp.MethodConnect:
		return &r.connect
	case fasthttp.MethodTrace:
		return &r.trace
	case fasthttp.MethodOptions:
		return &r.options
	default:
		return nil
	}
}

type AssetPath struct {
	Prefix         []byte // url prefix
	FileSystemRoot string // file locations
	StripSlashes   int    // how many slash words to strip from the url prefix
}

// Add a route to static files
// Prefix is a starting portion of the URL delimited by slashes
// fsRoot is the path to the top-level folder to serve files from
// StripSlashes is the number of slash delimited tokens to remove from the URL
// before appending it to the fsRoot to form the full file path
// Example: rx.AddStaticFilesRoute("/images/", "artifacts/images", 1)
func (r *Rox) AddStaticFilesRoute(prefix, fsRoot string, slashesToStrip int) {
	ap := AssetPath{Prefix: []byte(prefix), FileSystemRoot: fsRoot, StripSlashes: slashesToStrip}
	r.Options.assetPaths = append(r.Options.assetPaths, ap)
}

// See if we match a file handler - First match is the one we use
func (r *Rox) getFSHandler(ctx *fasthttp.RequestCtx) (handler fasthttp.RequestHandler, ok bool) {
	path := ctx.Path()
	for _, astPath := range r.Options.assetPaths {
		if bytes.HasPrefix(path, astPath.Prefix) {
			return fasthttp.FSHandler(astPath.FileSystemRoot, astPath.StripSlashes), true
		}
	}
	return
}
