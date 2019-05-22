package handlers

import (
	"net/http"
)

type optionsHandler struct {
	response *Response
}

// Returns the allowed methods and the DAV features implemented by the current server.
// For more information about the values and format read RFC4918 Sections 10.1 and 18.
func (oh optionsHandler) Handle() *Response {
	// Set the DAV compliance header:
	// 1: Server supports all the requirements specified in RFC2518
	// 3: Server supports all the revisions specified in RFC4918
	// calendar-access: Server supports all the extensions specified in RFC4791
	oh.response.SetHeader("DAV", "1, 3, calendar-access").
		SetHeader("Allow", "GET, HEAD, PUT, DELETE, OPTIONS, PROPFIND, REPORT").
		Set(http.StatusOK, "")

	return oh.response
}
