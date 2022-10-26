// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style

package rox

import (
	"fmt"
	"net/http"
	"strings"
)

// Api registers an api.
// 	- method:  supported HTTP methods,
// 	- pattern: url path matched pattern,
// 	- handler: http request handler.
func (r *Rox) Api(method string, pattern string, handler Handler) {
	if handler == nil {
		panic("router: nil handler")
	}

	if !strings.HasPrefix(pattern, "/") {
		panic(fmt.Errorf("router: pattern no leading / - %q", pattern))
	}

	t := r.selectTree(method)
	if t == nil {
		panic(fmt.Errorf("router: unknown http method - %q", method))
	}
	p := MustPattern(r.newPattern(pattern, &t.Regs))
	t.Add(p, handler)
}

// Get is a shortcut for Api(http.MethodGet, pattern, handler)
func (r *Rox) Get(pattern string, handler Handler) {
	r.Api(http.MethodGet, pattern, handler)
}

// Post is a shortcut for Api(http.MethodPost, pattern, handler)
func (r *Rox) Post(pattern string, handler Handler) {
	r.Api(http.MethodPost, pattern, handler)
}

// GetPost set Get and Post methods for pattern, handler)
func (r *Rox) GetPost(pattern string, handler Handler) {
	r.Api(http.MethodGet, pattern, handler)
	r.Api(http.MethodPost, pattern, handler)
}

// Put is a shortcut for Api(http.MethodPut, pattern, handler)
func (r *Rox) Put(pattern string, handler Handler) {
	r.Api(http.MethodPut, pattern, handler)
}

// Delete is a shortcut for Api(http.MethodDelete, pattern, handler)
func (r *Rox) Delete(pattern string, handler Handler) {
	r.Api(http.MethodDelete, pattern, handler)
}

// Head is a shortcut for Api(http.MethodHead, pattern, handler)
func (r *Rox) Head(pattern string, handler Handler) {
	r.Api(http.MethodHead, pattern, handler)
}

// MethodOptions is a shortcut for Api(http.MethodOptions, pattern, handler)
// Sorry for the asymmetry, but we will use Options for actual router options
func (r *Rox) MethodOptions(pattern string, handler Handler) {
	r.Api(http.MethodOptions, pattern, handler)
}

// Patch is a shortcut for Api(http.MethodPatch, pattern, handler)
func (r *Rox) Patch(pattern string, handler Handler) {
	r.Api(http.MethodPatch, pattern, handler)
}
