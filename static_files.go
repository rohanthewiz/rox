package rox

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

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
func (r *Rox) GetFSHandler(ctx *fasthttp.RequestCtx) (handler fasthttp.RequestHandler, ok bool) {
	path := ctx.Path()
	for _, astPath := range r.Options.assetPaths {
		if bytes.HasPrefix(path, astPath.Prefix) {
			return fasthttp.FSHandler(astPath.FileSystemRoot, astPath.StripSlashes), true
		}
	}
	return
}
