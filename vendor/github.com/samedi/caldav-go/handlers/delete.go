package handlers

import (
	"github.com/samedi/caldav-go/global"
	"net/http"
)

type deleteHandler struct {
	request  *http.Request
	response *Response
}

func (dh deleteHandler) Handle() *Response {
	precond := requestPreconditions{dh.request}

	// get the event from the storage
	resource, _, err := global.Storage.GetShallowResource(dh.request.URL.Path)
	if err != nil {
		return dh.response.SetError(err)
	}

	// TODO: Handle delete on collections
	if resource.IsCollection() {
		return dh.response.Set(http.StatusMethodNotAllowed, "")
	}

	// check ETag pre-condition
	resourceEtag, _ := resource.GetEtag()
	if !precond.IfMatch(resourceEtag) {
		return dh.response.Set(http.StatusPreconditionFailed, "")
	}

	// delete event after pre-condition passed
	err = global.Storage.DeleteResource(resource.Path)
	if err != nil {
		return dh.response.SetError(err)
	}

	return dh.response.Set(http.StatusNoContent, "")
}
