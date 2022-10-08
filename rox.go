package rox

import (
	"fmt"
	"log"
	"regexp"

	"github.com/valyala/fasthttp"
)

const defaultPort = "8800"
const defaultTLSPort = "443"
const ipAny = "0.0.0.0"

// RoxHandler handles HTTP requests.
type RoxHandler func(ctx *fasthttp.RequestCtx, params Params)

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
// The syntax of the pattern reference rox.NewPattern.
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

// Match returns the handler to use and path params
// matched the given method and path.
//
// If there is no registered handler that applies to the given method and path,
// Match returns a nil handler and an empty path parameters.
func (r *Rox) Match(method string, path string) (h RoxHandler, params Params) {
	t := r.selectTree(method)
	if t != nil {
		h = t.match(path, &params)
	}
	return
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
	// ----

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
		fileHander, ok := r.GetFSHandler(ctx)
		if ok {
			fileHander(ctx)
			return
		}

		t := r.selectTree(string(ctx.Method()))
		if t != nil {
			var params Params
			path := string(ctx.Path())

			if h := t.staticMatch(path); h != nil {
				fmt.Println("We have a direct match for -", path)
				h(ctx, params)
				return
			}

			h := t.patternMatch(path, &params)
			if h != nil {
				fmt.Println("We have a pattern match for -", path)
				h(ctx, params)
				return
			}

			msg := "Unknown Route (404) for " + path
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
	r.get.init()
	r.post.init()
	r.delete.init()
	r.put.init()
	r.patch.init()
	r.head.init()
	r.connect.init()
	r.trace.init()
	r.options.init()
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
