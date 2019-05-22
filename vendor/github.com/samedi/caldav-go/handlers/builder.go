package handlers

import (
	"net/http"
)

// HandlerInterface represents a CalDAV request handler. It has only one function `Handle`,
// which is used to handle the CalDAV request and returns the response.
type HandlerInterface interface {
	Handle() *Response
}

// NewHandler returns a new CalDAV request handler object based on the provided request.
// With the returned request handler, you can call `Handle()` to handle the request.
func NewHandler(request *http.Request) HandlerInterface {
	response := NewResponse()

	switch request.Method {
	case "GET":
		return getHandler{request, response, false}
	case "HEAD":
		return getHandler{request, response, true}
	case "PUT":
		return putHandler{request, response}
	case "DELETE":
		return deleteHandler{request, response}
	case "PROPFIND":
		return propfindHandler{request, response}
	case "OPTIONS":
		return optionsHandler{response}
	case "REPORT":
		return reportHandler{request, response}
	default:
		return notImplementedHandler{response}
	}
}
