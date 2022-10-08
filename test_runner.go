package rox

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strconv"

	"github.com/rohanthewiz/rerr" // for now
	"github.com/rohanthewiz/rox/core/constants"
	"github.com/rohanthewiz/rox/test_helpers"
	"github.com/valyala/fasthttp"
)

// TestRunner will test an endpoint on the passed in rox router
// A rox router with added routes is the first arg;
// the request containing the endpoint to test is the second arg
// See rox_test.go for example usage
func TestRunner(r *Rox, req *http.Request) (resp *fasthttp.Response, err error) {
	if req.Header.Get(constants.HeaderContentType) == "" {
		req.Header.Add(constants.HeaderContentType, constants.ContentTypeText)
	}

	if req.Body != http.NoBody && req.Header.Get(constants.HeaderContentLength) == "" {
		req.Header.Add(constants.HeaderContentLength, strconv.FormatInt(req.ContentLength, 10))
	}

	reqRaw, err := httputil.DumpRequest(req, true)
	if err != nil {
		return resp, rerr.Wrap(err, "Error obtaining raw HTTP request from req",
			"request", fmt.Sprintf("%v", req))
	}

	s := &fasthttp.Server{
		Handler: r.prepareServer(),
	}

	cw := &test_helpers.ConnWrap{}
	cw.R.Write(reqRaw)

	if err := s.ServeConn(cw); err != nil {
		return resp, rerr.Wrap(err, "Unexpected error from serveConn",
			"request", fmt.Sprintf("%v", req))
	}

	body, err := ioutil.ReadAll(&cw.W)
	if err != nil {
		return resp, rerr.Wrap(err, "Unexpected error from ReadAll",
			"request", fmt.Sprintf("%v", req))
	}
	// fmt.Println("***body -->", string(body))

	// Prepare response
	resp = &fasthttp.Response{}
	resp.SetBodyRaw(body)
	return
}
