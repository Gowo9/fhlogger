package main

import (
	"log"

	"github.com/fasthttp/router"
	"github.com/gowo9/fhlogger/fhzap"
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// init zap logger
	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	zapLogger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("zap.NewProduction failed, err=%s", err)
	}

	// init fhzap
	fhZap := fhzap.New(zapLogger,
		// specify ignore path
		fhzap.WithSkipPaths([]string{"/no-log", "/favicon.ico"}),
	)

	// init router
	r := router.New()
	r.GET("/foo", func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("foo")
	})
	r.GET("/bar", func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("bar")
	})
	r.GET("/no-log", func(ctx *fasthttp.RequestCtx) {
		_, _ = ctx.WriteString("no-log")
	})

	// combined middleware and listen
	finalHandler := fhZap.Combined(r.Handler)
	if err = fasthttp.ListenAndServe("127.0.0.1:8080", finalHandler); err != nil {
		log.Fatalf("fasthttp.ListenAndServe failed, err=%s", err)
	}
}
