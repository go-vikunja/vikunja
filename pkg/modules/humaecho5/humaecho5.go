// Package humaecho5 is a Huma adapter for labstack/echo/v5.
//
// Adapted from github.com/danielgtaylor/huma/v2/adapters/humaecho (MIT)
// with the echo/v5 port proposed in https://github.com/danielgtaylor/huma/pull/959.
// Remove this package once the upstream PR lands and the official adapter
// supports echo/v5.
package humaecho5

import (
	"context"
	"crypto/tls"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/labstack/echo/v5"
)

// MultipartMaxMemory is the maximum memory to use when parsing multipart
// form data.
var MultipartMaxMemory int64 = 8 * 1024

// echoContextKey is the context key under which the underlying *echo.Context
// is stashed on the request's context.Context. Handlers that run inside a
// Huma-dispatched call can retrieve it via ctx.Value(EchoContextKey).
type echoContextKey struct{}

// EchoContextKey is the exported key for retrieving the underlying echo
// context from a Huma handler's context.Context.
var EchoContextKey = echoContextKey{}

// Unwrap extracts the underlying Echo context from a Huma context. Panics if
// called on a context from a different adapter.
func Unwrap(ctx huma.Context) *echo.Context {
	for {
		if c, ok := ctx.(interface{ Unwrap() huma.Context }); ok {
			ctx = c.Unwrap()
			continue
		}
		break
	}
	if c, ok := ctx.(*echoCtx); ok {
		return c.Unwrap()
	}
	panic("not a humaecho5 context")
}

type echoCtx struct {
	op     *huma.Operation
	orig   *echo.Context
	status int
}

var _ huma.Context = &echoCtx{}

func (c *echoCtx) Unwrap() *echo.Context     { return c.orig }
func (c *echoCtx) Operation() *huma.Operation { return c.op }

func (c *echoCtx) Context() context.Context {
	// Stash the underlying echo context so downstream helpers
	// (e.g. auth.GetAuthFromContext) can retrieve it.
	return context.WithValue((*c.orig).Request().Context(), EchoContextKey, c.orig)
}

func (c *echoCtx) Method() string     { return (*c.orig).Request().Method }
func (c *echoCtx) Host() string       { return (*c.orig).Request().Host }
func (c *echoCtx) RemoteAddr() string { return (*c.orig).Request().RemoteAddr }
func (c *echoCtx) URL() url.URL       { return *(*c.orig).Request().URL }

func (c *echoCtx) Param(name string) string  { return (*c.orig).Param(name) }
func (c *echoCtx) Query(name string) string  { return (*c.orig).QueryParam(name) }
func (c *echoCtx) Header(name string) string { return (*c.orig).Request().Header.Get(name) }

func (c *echoCtx) EachHeader(cb func(name, value string)) {
	for name, values := range (*c.orig).Request().Header {
		for _, value := range values {
			cb(name, value)
		}
	}
}

func (c *echoCtx) BodyReader() io.Reader { return (*c.orig).Request().Body }

func (c *echoCtx) GetMultipartForm() (*multipart.Form, error) {
	err := (*c.orig).Request().ParseMultipartForm(MultipartMaxMemory)
	return (*c.orig).Request().MultipartForm, err
}

func (c *echoCtx) SetReadDeadline(deadline time.Time) error {
	return huma.SetReadDeadline((*c.orig).Response(), deadline)
}

func (c *echoCtx) SetStatus(code int) {
	c.status = code
	(*c.orig).Response().WriteHeader(code)
}

func (c *echoCtx) Status() int { return c.status }

func (c *echoCtx) AppendHeader(name, value string) {
	(*c.orig).Response().Header().Add(name, value)
}

func (c *echoCtx) SetHeader(name, value string) {
	(*c.orig).Response().Header().Set(name, value)
}

func (c *echoCtx) BodyWriter() io.Writer { return (*c.orig).Response() }

func (c *echoCtx) TLS() *tls.ConnectionState { return (*c.orig).Request().TLS }

func (c *echoCtx) Version() huma.ProtoVersion {
	r := (*c.orig).Request()
	return huma.ProtoVersion{
		Proto:      r.Proto,
		ProtoMajor: r.ProtoMajor,
		ProtoMinor: r.ProtoMinor,
	}
}

type router interface {
	Add(method, path string, handler echo.HandlerFunc, middlewares ...echo.MiddlewareFunc) echo.RouteInfo
}

type echoAdapter struct {
	http.Handler
	router router
}

func (a *echoAdapter) Handle(op *huma.Operation, handler func(huma.Context)) {
	// Convert {param} to :param for Echo's router.
	path := op.Path
	path = strings.ReplaceAll(path, "{", ":")
	path = strings.ReplaceAll(path, "}", "")
	a.router.Add(op.Method, path, func(c *echo.Context) error {
		ctx := &echoCtx{op: op, orig: c}
		handler(ctx)
		return nil
	})
}

// New creates a new Huma API using the provided Echo router.
func New(r *echo.Echo, config huma.Config) huma.API {
	return huma.NewAPI(config, &echoAdapter{Handler: r, router: r})
}

// NewWithGroup creates a new Huma API using the provided Echo router and group.
func NewWithGroup(r *echo.Echo, g *echo.Group, config huma.Config) huma.API {
	return huma.NewAPI(config, &echoAdapter{Handler: r, router: g})
}
