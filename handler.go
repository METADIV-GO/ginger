package ginger

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ApiHandler struct {
	Handler gin.HandlerFunc `json:"-"`
	Method  string          `json:"method"`
	Path    string          `json:"path"`
	Opts    *ApiOpts        `json:"opts"`
}

type WsHandler struct {
	Handler gin.HandlerFunc `json:"-"`
	Path    string          `json:"path"`
}

type CornHandler struct {
	Handler  func() `json:"-"`
	InitExec bool   `json:"init_exec"`
	Pattern  string `json:"pattern"`
}

type InitJobHandler struct {
	Handler func() `json:"-"`
	After   bool   `json:"after"`
}

type MiddlewareHandler struct {
	Handler    gin.HandlerFunc `json:"-"`
	MatchPaths []string        `json:"match_paths"`
	SkipPaths  []string        `json:"exclude"`
}

func apiToHandler[T any](f func(ctx *Context[T])) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := NewContext[T](ctx)
		f(c)

		// if file is served, no need to respond
		if c.hasResp && c.isFile {
			return
		}

		// unexpected, service did not respond
		if !c.hasResp || c.Response == nil {
			ctx.JSON(500, gin.H{
				"message": "service did not respond",
			})
			return
		}

		ctx.JSON(c.status, c.Response)
	}
}

func wsToHandler[T any](f func(ctx *Context[T], ws *websocket.Conn)) gin.HandlerFunc {
	wsUpGrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	return func(c *gin.Context) {
		ws, err := wsUpGrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer ws.Close()

		ctx := NewContext[T](c)
		f(ctx, ws)
	}
}

func middlewareToHandler(f func(ctx *Context[struct{}])) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := NewContext[struct{}](ctx)
		f(c)
	}
}
