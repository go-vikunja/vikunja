package handlers

import (
	"net/http"
)

const (
	HD_DEPTH              = "Depth"
	HD_DEPTH_DEEP         = "1"
	HD_PREFER             = "Prefer"
	HD_PREFER_MINIMAL     = "return=minimal"
	HD_PREFERENCE_APPLIED = "Preference-Applied"
)

type headers struct {
	http.Header
}

func (h headers) IsDeep() bool {
	depth := h.Get(HD_DEPTH)
	return (depth == HD_DEPTH_DEEP)
}

func (h headers) IsMinimal() bool {
	prefer := h.Get(HD_PREFER)
	return (prefer == HD_PREFER_MINIMAL)
}
