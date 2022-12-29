# fasthttp logger

Requests zap middleware for [fasthttp](https://github.com/valyala/fasthttp)

## Example

[full code](/example/fhzap/main.go)

```golang
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
```
