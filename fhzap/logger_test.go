package fhzap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"go.uber.org/zap"
)

var testURLList = []string{
	"/foo",
	"/no-log",
}

func defaultHandler() fasthttp.RequestHandler {
	r := router.New()
	r.GET("/foo", func(ctx *fasthttp.RequestCtx) {})
	r.GET("/no-log", func(ctx *fasthttp.RequestCtx) {})

	// delayMiddleware 故意延遲一小段時間，讓 log 中的 Latency 數值較明顯
	delayMiddleware := func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)
			time.Sleep(100 * time.Microsecond)
		}
	}

	return delayMiddleware(r.Handler)
}

func logEmulator(t *testing.T, fhZapInit func(logger *zap.Logger) *FHZap) (outRes []byte, err error) {
	originalStderr := os.Stderr
	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		err = fmt.Errorf("os.Pipe() failed %s", err)
		return
	}
	os.Stderr = pipeW

	// init zap logger

	zapLogger, err := zap.NewProduction()
	if err != nil {
		err = fmt.Errorf("zap.NewProduction() %s", err)
		return
	}
	fhZap := fhZapInit(zapLogger)

	// run fasthttp server

	ln := fasthttputil.NewInmemoryListener()
	rHandler := defaultHandler()
	s := fasthttp.Server{
		Handler: fhZap.Combined(rHandler),
	}
	go func() {
		if serveErr := s.Serve(ln); serveErr != nil {
			t.Error(serveErr)
		}
	}()

	// simulation client request

	client := http.Client{
		Transport: &http.Transport{
			Dial: func(_, _ string) (net.Conn, error) {
				return ln.Dial()
			},
		},
	}
	const urlPrefix = "http://no.use.host"
	for _, testURL := range testURLList {
		req, _ := http.NewRequestWithContext(context.Background(), "GET", urlPrefix+testURL, http.NoBody)
		req.Header.Set("user-agent", "Client-User-Agent")
		resp, _ := client.Do(req)
		resp.Body.Close()
		if err != nil {
			t.Error(err)
		}
	}

	// read result

	_ = zapLogger.Sync()
	pipeW.Close()
	outRes, err = io.ReadAll(pipeR)
	if err != nil {
		err = fmt.Errorf("io.ReadAll(r) %s", err)
		return
	}
	os.Stderr = originalStderr
	_ = pipeR.Close()

	return outRes, nil
}

// linePrepare prepare output line for check result
func linePrepare(t *testing.T, output []byte, expectLogNum int, check func(line []byte)) {
	lines := bytes.Split(output, []byte{'\n'})
	logNum := len(lines) - 1
	if logNum != expectLogNum {
		t.Errorf("log amount = %d, want %d", logNum, expectLogNum)
	}

	if check != nil {
		for _, line := range lines[:logNum] {
			check(line)
		}
	}
}

// TestDefault 單純測試預設設定下是否會生成 log 訊息
func TestDefault(t *testing.T) {
	output, err := logEmulator(t,
		func(logger *zap.Logger) *FHZap {
			fhZap := New(logger)
			return fhZap
		},
	)
	if err != nil {
		t.Error(err)
	}

	linePrepare(t, output, len(testURLList),
		func(line []byte) {
			type defaultCheck struct {
				Level   string
				TS      float64
				Caller  string
				Msg     string
				IP      string
				Method  string
				URI     string
				Agent   string
				Status  int
				Latency float64
			}

			var dc defaultCheck
			err = json.Unmarshal(line, &dc)
			if err == nil {
				dcElem := reflect.ValueOf(&dc).Elem()
				dcFieldNum := dcElem.NumField()
				for i := 0; i < dcFieldNum; i++ {
					if dcElem.Field(i).IsZero() {
						t.Errorf("%s has no value", dcElem.Type().Field(i).Name)
					}
				}
			} else {
				t.Log(string(line))
				t.Error(err)
			}
		},
	)
}

// TestWithLogMsg 測試 WithLogMsg 設定選項
func TestWithLogMsg(t *testing.T) {
	const testMsg = "test message text"

	output, err := logEmulator(t, func(logger *zap.Logger) *FHZap {
		fhZap := New(logger, WithLogMsg(testMsg))
		return fhZap
	})
	if err != nil {
		t.Error(err)
	}

	linePrepare(t, output, len(testURLList),
		func(line []byte) {
			type msgCheck struct {
				Msg string
			}

			var mc msgCheck
			err = json.Unmarshal(line, &mc)
			if err == nil {
				if mc.Msg != testMsg {
					t.Errorf("Msg = %s, want "+testMsg, mc.Msg)
				}
			} else {
				t.Error(err)
			}
		},
	)
}

// TestWithSkipPaths 測試 WithSkipPaths 設定選項
func TestWithSkipPaths(t *testing.T) {
	output, err := logEmulator(t, func(logger *zap.Logger) *FHZap {
		fhZap := New(logger, WithSkipPaths([]string{"/no-log"}))
		return fhZap
	})
	if err != nil {
		t.Error(err)
	}

	linePrepare(t, output, len(testURLList)-1,
		func(line []byte) {
			type uriCheck struct {
				URI string
			}

			var uc uriCheck
			err = json.Unmarshal(line, &uc)
			if err == nil {
				if uc.URI == "no-log" {
					t.Errorf("find URI('no-log') in logs")
				}
			} else {
				t.Error(err)
			}
		},
	)
}

// TestWithIsSkip 測試 WithIsSkipFunc 設定選項
func TestWithIsSkip(t *testing.T) {
	output, err := logEmulator(t, func(logger *zap.Logger) *FHZap {
		fhZap := New(logger, WithIsSkipFunc(
			func(rc *fasthttp.RequestCtx) bool {
				// always return true => no log will be write
				return true
			}))
		return fhZap
	})
	if err != nil {
		t.Error(err)
	}

	linePrepare(t, output, 0, nil)
}
