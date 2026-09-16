// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package events

import "context"

// RequestMeta carries information about the originating HTTP request. It is
// stashed on the request context by a middleware and copied onto message
// metadata at publish time, so listeners (e.g. audit) can attribute an event
// to a request without every dispatch site changing its signature.
type RequestMeta struct {
	IP        string
	UserAgent string
	RequestID string
}

// Message metadata keys holding request information.
const (
	MetadataKeyIP        = "request_ip"
	MetadataKeyUserAgent = "request_user_agent"
	MetadataKeyRequestID = "request_id"
)

type requestMetaKeyType struct{}

var requestMetaKey requestMetaKeyType

// WithRequestMeta returns a context carrying the given request metadata.
func WithRequestMeta(ctx context.Context, meta *RequestMeta) context.Context {
	return context.WithValue(ctx, requestMetaKey, meta)
}

// RequestMetaFromContext returns the request metadata stored on the context,
// or nil if there is none.
func RequestMetaFromContext(ctx context.Context) *RequestMeta {
	if ctx == nil {
		return nil
	}
	meta, _ := ctx.Value(requestMetaKey).(*RequestMeta)
	return meta
}
