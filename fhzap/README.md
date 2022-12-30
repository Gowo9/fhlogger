# fasthttp logger

Requests zap middleware for [fasthttp](https://github.com/valyala/fasthttp)

## Import

```golang
import "github.com/gowo9/fhlogger/fhzap"
```

## Example

[full code](/example/fhzap/main.go)

```golang
// init fhzap
fhZap := fhzap.New(zapLogger,
    // specify ignore path
    fhzap.WithSkipPaths([]string{"/no-log", "/favicon.ico"}),
)

// ...

// use fhZap.Combined function to embed other handler
finalHandler := fhZap.Combined(r.Handler)
if err = fasthttp.ListenAndServe("127.0.0.1:8080", finalHandler); err != nil {
    log.Fatalf("fasthttp.ListenAndServe failed, err=%s", err)
}
```
