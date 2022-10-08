package cors

import "github.com/valyala/fasthttp"

var (
	corsAllowHeaders     = "authorization, Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"
	corsAllowMethods     = "HEAD,GET,POST,PUT,DELETE,OPTIONS"
	corsAllowOrigin      = "*"
	corsAllowCredentials = "true"
)

func MidWareCors(ctx *fasthttp.RequestCtx) (ok bool) {
	ctx.Response.Header.Set("Access-Control-Allow-Credentials", corsAllowCredentials)
	ctx.Response.Header.Set("Access-Control-Allow-Headers", corsAllowHeaders)
	ctx.Response.Header.Set("Access-Control-Allow-Methods", corsAllowMethods)
	ctx.Response.Header.Set("Access-Control-Allow-Origin", corsAllowOrigin)
	return true
}
