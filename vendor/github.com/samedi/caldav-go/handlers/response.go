package handlers

import (
	"github.com/samedi/caldav-go/errs"
	"io"
	"net/http"
)

// Response represents the handled CalDAV response. Used this when one needs to proxy the generated
// response before being sent back to the client.
type Response struct {
	Status int
	Header http.Header
	Body   string
	Error  error
}

// NewResponse initializes a new response object.
func NewResponse() *Response {
	return &Response{
		Header: make(http.Header),
	}
}

// Set sets the the status and body of the response.
func (r *Response) Set(status int, body string) *Response {
	r.Status = status
	r.Body = body

	return r
}

// SetHeader adds a header to the response.
func (r *Response) SetHeader(key, value string) *Response {
	r.Header.Set(key, value)

	return r
}

// SetError sets the response as an error. It inflects the response status based on the provided error.
func (r *Response) SetError(err error) *Response {
	r.Error = err

	switch err {
	case errs.ResourceNotFoundError:
		r.Status = http.StatusNotFound
	case errs.UnauthorizedError:
		r.Status = http.StatusUnauthorized
	case errs.ForbiddenError:
		r.Status = http.StatusForbidden
	default:
		r.Status = http.StatusInternalServerError
	}

	return r
}

// Write writes the response back to the client using the provided `ResponseWriter`.
func (r *Response) Write(writer http.ResponseWriter) {
	if r.Error == errs.UnauthorizedError {
		r.SetHeader("WWW-Authenticate", `Basic realm="Restricted"`)
	}

	for key, values := range r.Header {
		for _, value := range values {
			writer.Header().Set(key, value)
		}
	}

	writer.WriteHeader(r.Status)
	io.WriteString(writer, r.Body)
}
