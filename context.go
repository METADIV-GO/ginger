package ginger

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/METADIV-GO/ginger/pkg/file_type"
	"github.com/METADIV-GO/ginger/pkg/logger"
	gin_request "github.com/METADIV-GO/ginger/pkg/request"
	"github.com/METADIV-GO/gorm"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
)

type Context[T any] struct {
	Engine *engine
	GinCtx *gin.Context

	TraceId string

	Request  *T
	Response *Response

	// use internal
	startAt time.Time
	hasResp bool
	isFile  bool
	status  int
}

func NewContext[T any](ginCtx *gin.Context) *Context[T] {
	traceId, err := gonanoid.Generate("2346789abcdefghijkmnopqrtwxyzABCDEFGHJKLMNOPQRTUVWXYZ", 21)
	if err != nil {
		panic(err)
	}
	return &Context[T]{
		Engine:  Engine,
		GinCtx:  ginCtx,
		TraceId: traceId,
		Request: gin_request.GinRequest[T](ginCtx),
		hasResp: false,
		isFile:  false,
		startAt: time.Now(),
	}
}

/*
Page returns the pagination object from the request.
*/
func (c *Context[T]) Page() *gorm.Pagination {
	page := new(gorm.Pagination)
	c.GinCtx.Bind(page)
	return page
}

/*
Sort returns the sorting object from the request.
*/
func (c *Context[T]) Sort() *gorm.Sorting {
	sort := new(gorm.Sorting)
	c.GinCtx.Bind(sort)
	return sort
}

/*
Locale returns the locale from the request header.
*/
func (c *Context[T]) Locale() string {
	return c.GinCtx.GetHeader(HEADER_X_LOCALE)
}

/*
Authorization returns the authorization from the request header.
*/
func (c *Context[T]) Authorization() string {
	return c.GinCtx.GetHeader(HEADER_AUTHORIZATION)
}

/*
BearerToken returns the bearer token from the request header.
*/
func (c *Context[T]) BearerToken() string {
	auth := c.Authorization()
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		auth, "Bearer", ""), "BEARER", "bearer"), "bearer", ""), " ", "")
}

/*
TraceID returns the trace id from the request header.
*/
func (c *Context[T]) TraceID() string {
	return c.TraceId
}

/*
IP returns the client's IP address.
*/
func (c *Context[T]) IP() string {
	return c.GinCtx.ClientIP()
}

/*
Agent returns the client's user agent.
*/
func (c *Context[T]) Agent() string {
	return c.GinCtx.Request.UserAgent()
}

/*
LogErr logs the error message.
*/
func (c *Context[T]) LogErr(msg ...any) {
	msg = append([]any{fmt.Sprintf("trace_id: %s ip: %s agent: %s", c.TraceID(), c.IP(), c.Agent())}, msg...)
	logger.ERROR(msg...)
}

/*
LogInfo logs the info message.
*/
func (c *Context[T]) LogInfo(msg ...any) {
	msg = append([]any{fmt.Sprintf("trace_id: %s ip: %s agent: %s", c.TraceID(), c.IP(), c.Agent())}, msg...)
	logger.INFO(msg...)
}

/*
LogDebug logs the debug message.
*/
func (c *Context[T]) LogDebug(msg ...any) {
	msg = append([]any{fmt.Sprintf("trace_id: %s ip: %s agent: %s", c.TraceID(), c.IP(), c.Agent())}, msg...)
	logger.DEBUG(msg...)
}

/*
OK returns a successful response.
*/
func (c *Context[T]) OK(data any, page ...*gorm.Pagination) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	var p *gorm.Pagination
	if len(page) > 0 {
		p = page[0]
	}

	c.Response = &Response{
		Success:    true,
		TraceId:    c.TraceId,
		Time:       time.Now().Format(time.RFC3339),
		Duration:   time.Since(c.startAt).Milliseconds(),
		Pagination: p,
		Data:       data,
	}
	c.hasResp = true
	c.status = http.StatusOK
}

/*
OKFile returns a file response.
*/
func (c *Context[T]) OKFile(bytes []byte, filename ...string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	var name string
	if len(filename) == 0 || filename[0] == "" {
		name = "file"
	} else {
		name = filename[0]
	}

	c.GinCtx.Header("Content-Disposition", "filename="+name)
	c.GinCtx.Data(http.StatusOK, file_type.DetermineFileType(name), bytes)
	c.hasResp = true
	c.isFile = true
	c.status = http.StatusOK
}

/*
OKDownload is a helper function to respond with a 200 status code and file.
*/
func (c *Context[T]) OKDownload(bytes []byte, filename ...string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	var name string
	if len(filename) == 0 || filename[0] == "" {
		name = "file"
	} else {
		name = filename[0]
	}

	c.GinCtx.Header("Content-Disposition", "filename="+name)
	c.GinCtx.Data(http.StatusOK, "application/octet-stream", bytes)
	c.hasResp = true
	c.isFile = true
	c.status = http.StatusOK
}

/*
Err is a helper function to respond with an error status code (400).
*/
func (c *Context[T]) Err(message string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	c.Response = &Response{
		Success:    false,
		TraceId:    c.TraceId,
		Time:       time.Now().Format(time.RFC3339),
		Duration:   time.Since(c.startAt).Milliseconds(),
		ErrMessage: message,
	}
	c.hasResp = true
	c.status = http.StatusBadRequest
}

/*
Unauthorized is a helper function to respond with an unauthorized status code (401).
*/
func (c *Context[T]) Unauthorized(message string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	c.Response = &Response{
		Success:    false,
		TraceId:    c.TraceId,
		Time:       time.Now().Format(time.RFC3339),
		Duration:   time.Since(c.startAt).Milliseconds(),
		ErrMessage: message,
	}
	c.hasResp = true
	c.status = http.StatusUnauthorized
}

/*
Forbidden is a helper function to respond with a forbidden status code (403).
*/
func (c *Context[T]) Forbidden(message string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	c.Response = &Response{
		Success:    false,
		TraceId:    c.TraceId,
		Time:       time.Now().Format(time.RFC3339),
		Duration:   time.Since(c.startAt).Milliseconds(),
		ErrMessage: message,
	}
	c.hasResp = true
	c.status = http.StatusForbidden
}

/*
InternalServerError is a helper function to respond with an internal server error status code (500).
*/
func (c *Context[T]) InternalServerError(message string) {
	if c.hasResp {
		c.LogErr("double response")
		return
	}

	c.Response = &Response{
		Success:    false,
		TraceId:    c.TraceId,
		Time:       time.Now().Format(time.RFC3339),
		Duration:   time.Since(c.startAt).Milliseconds(),
		ErrMessage: message,
	}
	c.hasResp = true
	c.status = http.StatusInternalServerError
}
