package fhzap

import (
	"time"

	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type IsSkipFunc func(*fasthttp.RequestCtx) bool
type PreCtxDealFunc func(*fasthttp.RequestCtx) []zapcore.Field
type PostCtxDealFunc func(*fasthttp.RequestCtx, *[]zapcore.Field, time.Duration)

func DefaultIsSkipFunc(*fasthttp.RequestCtx) bool {
	return false
}

func DefaultPreCtxDealFunc(ctx *fasthttp.RequestCtx) []zapcore.Field {
	const defaultFieldsNum = 6

	zapFields := make([]zapcore.Field, 0, defaultFieldsNum)
	zapFields = append(zapFields,
		zap.String("ip", ctx.RemoteIP().String()),
		zap.ByteString("method", ctx.Request.Header.Method()),
		zap.String("uri", string(ctx.RequestURI())),
	)

	uaCopy := string(ctx.UserAgent())
	if uaCopy != "" {
		zapFields = append(zapFields, zap.String("agent", uaCopy))
	}

	return zapFields
}

func DefaultPostCtxDealFunc(ctx *fasthttp.RequestCtx, existField *[]zapcore.Field, lattency time.Duration) {
	*existField = append(*existField,
		zap.Int("status", ctx.Response.StatusCode()),
		zap.Duration("latency", lattency))
}

// FHZapLogger fasthttp zap logger
type FHZap struct {
	opts fhZapOptions

	logger *zap.Logger
}

func New(logger *zap.Logger, optArgs ...Option) *FHZap {
	if logger == nil {
		panic("logger can not be nil")
	}
	opts := defaultFHZapOptions
	for _, o := range optArgs {
		o.apply(&opts)
	}

	fhZap := &FHZap{
		opts:   opts,
		logger: logger,
	}
	fhZap.init()

	return fhZap
}

func (fhz *FHZap) init() {
	if fhz.opts.isSkipFunc == nil {
		if fhz.opts.skipPathMap != nil && len(fhz.opts.skipPathMap) > 0 {
			fhz.opts.isSkipFunc = fhz.opts.inSkipPathMap
		} else {
			fhz.opts.isSkipFunc = DefaultIsSkipFunc
		}
	}
}

func (fhz *FHZap) Combined(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		var sT, eT time.Time
		var logField []zapcore.Field

		isSkip := fhz.opts.isSkipFunc(ctx)
		if !isSkip {
			logField = fhz.opts.preCtxDealFunc(ctx)
			sT = time.Now()
		}

		next(ctx)

		if !isSkip {
			eT = time.Now()
			fhz.opts.postCtxDealFunc(ctx, &logField, eT.Sub(sT))
			fhz.logger.Info(fhz.opts.logMsg, logField...)
		}
	}
}
