package ginger

import (
	"time"

	"github.com/gorilla/websocket"
)

type TypescriptOpt struct {
	Models       []any    `json:"models"`
	FunctionName string   `json:"function_name"`
	Paths        []string `json:"paths"`
	Forms        []string `json:"forms"`
	Body         string   `json:"body"`
	Response     string   `json:"response"`
}

type RateLimitOpt struct {
	Rate     int64         `json:"rate"`
	Duration time.Duration `json:"duration"`
}

type CacheOpt struct {
	Duration time.Duration `json:"duration"`
}

type ApiOpts struct {
	RateLimit  *RateLimitOpt  `json:"rate_limit"`
	Cache      *CacheOpt      `json:"cache"`
	Typescript *TypescriptOpt `json:"typescript"`
}

func GET[T any](path string, handler func(ctx *Context[T]), opts ...*ApiOpts) {
	var opt *ApiOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	Engine.ApiHandlers = append(Engine.ApiHandlers, ApiHandler{
		Handler: apiToHandler[T](handler),
		Method:  "GET",
		Path:    path,
		Opts:    opt,
	})
}

func POST[T any](path string, handler func(ctx *Context[T]), opts ...*ApiOpts) {
	var opt *ApiOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	Engine.ApiHandlers = append(Engine.ApiHandlers, ApiHandler{
		Handler: apiToHandler[T](handler),
		Method:  "POST",
		Path:    path,
		Opts:    opt,
	})
}

func PUT[T any](path string, handler func(ctx *Context[T]), opts ...*ApiOpts) {
	var opt *ApiOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	Engine.ApiHandlers = append(Engine.ApiHandlers, ApiHandler{
		Handler: apiToHandler[T](handler),
		Method:  "PUT",
		Path:    path,
		Opts:    opt,
	})
}

func DELETE[T any](path string, handler func(ctx *Context[T]), opts ...*ApiOpts) {
	var opt *ApiOpts
	if len(opts) > 0 {
		opt = opts[0]
	}

	Engine.ApiHandlers = append(Engine.ApiHandlers, ApiHandler{
		Handler: apiToHandler[T](handler),
		Method:  "DELETE",
		Path:    path,
		Opts:    opt,
	})
}

func WS[T any](path string, handler func(ctx *Context[T], ws *websocket.Conn)) {
	Engine.WsHandlers = append(Engine.WsHandlers, WsHandler{
		Handler: wsToHandler[T](handler),
		Path:    path,
	})
}

func Corn(pattern string, handler func(), initExec bool) {
	Engine.CronHandlers = append(Engine.CronHandlers, CornHandler{
		Handler:  handler,
		InitExec: initExec,
		Pattern:  pattern,
	})
}

func InitJob(handler func(), after bool) {
	Engine.InitJobs = append(Engine.InitJobs, InitJobHandler{
		Handler: handler,
		After:   after,
	})
}

func Middleware(handler func(ctx *Context[struct{}]), matchPaths []string, skipPaths []string) {
	Engine.Middlewares = append(Engine.Middlewares, MiddlewareHandler{
		Handler:    middlewareToHandler(handler),
		MatchPaths: matchPaths,
		SkipPaths:  skipPaths,
	})
}
