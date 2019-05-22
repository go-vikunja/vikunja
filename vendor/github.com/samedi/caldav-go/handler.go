package caldav

import (
	"net/http"

	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/handlers"
)

// RequestHandler handles the given CALDAV request and writes the reponse righ away. This function is to be
// used by passing it directly as the handle func to the `http` lib. Example: http.HandleFunc("/", caldav.RequestHandler).
func RequestHandler(writer http.ResponseWriter, request *http.Request) {
	response := HandleRequest(request)
	response.Write(writer)
}

// HandleRequest handles the given CALDAV request and returns the response. Useful when the caller
// wants to do something else with the response before writing it to the response stream.
func HandleRequest(request *http.Request) *handlers.Response {
	handler := handlers.NewHandler(request)
	return handler.Handle()
}

// HandleRequestWithStorage handles the request the same way as `HandleRequest` does, but before,
// it sets the given storage that will be used throughout the request handling flow.
func HandleRequestWithStorage(request *http.Request, stg data.Storage) *handlers.Response {
	SetupStorage(stg)
	return HandleRequest(request)
}
