package rox

import (
	"github.com/valyala/fasthttp"
)

// Handler handles HTTP requests.
type Handler func(ctx *fasthttp.RequestCtx, params Params)
