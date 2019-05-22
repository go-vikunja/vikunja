package handlers

import (
	"net/http"
)

type requestPreconditions struct {
	request *http.Request
}

func (p *requestPreconditions) IfMatch(etag string) bool {
	etagMatch := p.request.Header["If-Match"]
	return len(etagMatch) == 0 || etagMatch[0] == "*" || etagMatch[0] == etag
}

func (p *requestPreconditions) IfMatchPresent() bool {
	return len(p.request.Header["If-Match"]) != 0
}

func (p *requestPreconditions) IfNoneMatch(value string) bool {
	valueMatch := p.request.Header["If-None-Match"]
	return len(valueMatch) == 1 && valueMatch[0] == value
}
