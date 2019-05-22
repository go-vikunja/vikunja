package handlers

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

// This function reads the request body and restore its content, so that
// the request body can be read a second time.
func readRequestBody(request *http.Request) string {
	// Read the content
	body, _ := ioutil.ReadAll(request.Body)
	// Restore the io.ReadCloser to its original state
	request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	// Use the content
	return string(body)
}
