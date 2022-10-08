package rox

import "github.com/valyala/fasthttp"

type MiddleWareFunc func(ctx *fasthttp.RequestCtx) (ok bool)

type MiddleWare struct {
	MidFunc  MiddleWareFunc
	FailCode int
}

// Use adds a middleware function before regular routes
// failCode is the http response code to return if we are immediately stopping the request
func (r *Rox) Use(m MiddleWareFunc, failCode int) {
	r.middlewares = append(r.middlewares, MiddleWare{MidFunc: m, FailCode: failCode})
}
