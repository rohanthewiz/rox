// Copyright (c) 2019,CAO HONGJU. All rights reserved.
// Use of this source code is governed by a MIT-style

package rox

import (
	"context"
	"sync"
)

const (
	maxParams = 20
)

// Params holds the path parameters extracted from the HTTP request.
type Params struct {
	path    string
	indices [maxParams * 2]int16
	names   []string
}

// ByName returns the value of the first parameter
// that matched the given name.
// Otherwise, an empty string is returned.
func (p Params) ByName(name string) string {
	for i, v := range p.names {
		if v == name {
			return p.Value(i)
		}
	}
	return ""
}

// Count returns the number of parameters.
func (p Params) Count() int {
	return len(p.names)
}

// Name returns the parameter name of the given index.
func (p Params) Name(i int) string {
	return p.names[i]
}

// Value returns the parameter value of the given index.
func (p Params) Value(i int) string {
	i = i << 1
	return p.path[p.indices[i]:p.indices[i+1]]
}

// PathParams pulls the path parameters from a request context,
// or returns nil if none are present.
func PathParams(c context.Context) *Params {
	p, _ := c.Value(paramsKey).(*Params)
	return p
}

var (
	paramsKey     = key{}
	paramsCtxPool = sync.Pool{
		New: func() any {
			return new(paramsCtx)
		},
	}
)

type key struct{}

func newParamsCtx(parent context.Context) *paramsCtx {
	c := paramsCtxPool.Get().(*paramsCtx)
	c.Context = parent
	return c
}

type paramsCtx struct {
	context.Context
	params Params
}

func (c *paramsCtx) Value(key any) any {
	if paramsKey == key {
		return &c.params
	}
	return c.Context.Value(key)
}

func (c *paramsCtx) Close() {
	c.Context = nil
	c.params.names = nil
	paramsCtxPool.Put(c)
}
