package rox

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/rohanthewiz/rox/test_helpers"
	"github.com/valyala/fasthttp"
)

// TestRunner will test an endpoint on the passed in rox router
// First param: A rox router with added routes;
// Second param: An already prepared fasthttp.Server or nil (normally) in which case the server will be prepared in this function
// Third param: the request containing the endpoint to test
// See rox_test.go for example usage
func TestRunner(r *Rox, s *fasthttp.Server, req *http.Request) (resp *fasthttp.Response, err error) {
	// Prepare Request
	if req.Header.Get(HeaderContentType) == "" {
		req.Header.Add(HeaderContentType, ContentTypeText)
	}

	if req.Body != http.NoBody && req.Header.Get(HeaderContentLength) == "" {
		req.Header.Add(HeaderContentLength, strconv.FormatInt(req.ContentLength, 10))
	}

	reqRaw, err := httputil.DumpRequest(req, true)
	if err != nil {
		return resp, errors.New("Error obtaining raw HTTP request from req - " + err.Error())
	}

	cw := &test_helpers.ConnWrap{}
	cw.R.Write(reqRaw)

	// Prepare Server, or we'll just work with the one passed in
	if s == nil {
		s = &fasthttp.Server{
			Handler: r.prepareServer(),
		}
	}

	if err := s.ServeConn(cw); err != nil {
		return resp, errors.New("Unexpected error from serveConn - " + err.Error())
	}

	body, err := ioutil.ReadAll(&cw.W)
	if err != nil {
		return resp, errors.New("Error when reading response body - " + err.Error())
	}

	// Prepare response
	resp = &fasthttp.Response{}
	resp.SetBodyRaw(body)
	return
}
