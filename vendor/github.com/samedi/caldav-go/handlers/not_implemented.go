package handlers

import (
	"net/http"
)

type notImplementedHandler struct {
	response *Response
}

func (h notImplementedHandler) Handle() *Response {
	return h.response.Set(http.StatusNotImplemented, "")
}
