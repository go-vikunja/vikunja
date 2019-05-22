package handlers

import (
	"github.com/samedi/caldav-go/errs"
	"github.com/samedi/caldav-go/global"
	"net/http"
)

type putHandler struct {
	request  *http.Request
	response *Response
}

func (ph putHandler) Handle() *Response {
	requestBody := readRequestBody(ph.request)
	precond := requestPreconditions{ph.request}
	success := false

	// check if resource exists
	resourcePath := ph.request.URL.Path
	resource, found, err := global.Storage.GetShallowResource(resourcePath)
	if err != nil && err != errs.ResourceNotFoundError {
		return ph.response.SetError(err)
	}

	// PUT is allowed in 2 cases:
	//
	// 1. Item NOT FOUND and there is NO ETAG match header: CREATE a new item
	if !found && !precond.IfMatchPresent() {
		// create new event resource
		resource, err = global.Storage.CreateResource(resourcePath, requestBody)
		if err != nil {
			return ph.response.SetError(err)
		}

		success = true
	}

	if found {
		// TODO: Handle PUT on collections
		if resource.IsCollection() {
			return ph.response.Set(http.StatusPreconditionFailed, "")
		}

		// 2. Item exists, the resource etag is verified and there's no IF-NONE-MATCH=* header: UPDATE the item
		resourceEtag, _ := resource.GetEtag()
		if found && precond.IfMatch(resourceEtag) && !precond.IfNoneMatch("*") {
			// update resource
			resource, err = global.Storage.UpdateResource(resourcePath, requestBody)
			if err != nil {
				return ph.response.SetError(err)
			}

			success = true
		}
	}

	if !success {
		return ph.response.Set(http.StatusPreconditionFailed, "")
	}

	resourceEtag, _ := resource.GetEtag()
	return ph.response.SetHeader("ETag", resourceEtag).
		Set(http.StatusCreated, "")
}
