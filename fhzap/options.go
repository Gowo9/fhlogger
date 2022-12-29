package fhzap

import (
	"github.com/valyala/fasthttp"
)

type fhZapOptions struct {
	logMsg          string
	skipPathMap     map[string]struct{}
	isSkipFunc      IsSkipFunc
	preCtxDealFunc  PreCtxDealFunc
	postCtxDealFunc PostCtxDealFunc
}

func (fhZOpt fhZapOptions) inSkipPathMap(ctx *fasthttp.RequestCtx) bool {
	_, exist := fhZOpt.skipPathMap[string(ctx.Path())]
	return exist
}

var defaultFHZapOptions = fhZapOptions{
	logMsg:          "request record",
	skipPathMap:     nil,
	isSkipFunc:      nil,
	preCtxDealFunc:  DefaultPreCtxDealFunc,
	postCtxDealFunc: DefaultPostCtxDealFunc,
}

type Option interface {
	apply(*fhZapOptions)
}

type funcFHZapOption struct {
	f func(*fhZapOptions)
}

func (fdo *funcFHZapOption) apply(do *fhZapOptions) {
	fdo.f(do)
}

func newFHZapFuncOption(f func(*fhZapOptions)) *funcFHZapOption {
	return &funcFHZapOption{
		f: f,
	}
}

func WithLogMsg(msg string) Option {
	return newFHZapFuncOption(func(o *fhZapOptions) {
		o.logMsg = msg
	})
}

func WithSkipPaths(pathList []string) Option {
	return newFHZapFuncOption(func(o *fhZapOptions) {
		o.skipPathMap = make(map[string]struct{}, len(pathList))
		for _, p := range pathList {
			o.skipPathMap[p] = struct{}{}
		}
	})
}

func WithIsSkipFunc(fn IsSkipFunc) Option {
	return newFHZapFuncOption(func(o *fhZapOptions) {
		o.isSkipFunc = fn
	})
}

func WithPreCtxDealFunc(fn PreCtxDealFunc) Option {
	return newFHZapFuncOption(func(o *fhZapOptions) {
		o.preCtxDealFunc = fn
	})
}

func WithPostCtxDealFunc(fn PostCtxDealFunc) Option {
	return newFHZapFuncOption(func(o *fhZapOptions) {
		o.postCtxDealFunc = fn
	})
}
