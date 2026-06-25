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

package audit

import (
	"encoding/json"

	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"

	"github.com/ThreeDotsLabs/watermill/message"
)

type auditListener struct {
	handle func(msg *message.Message) error
}

func (l *auditListener) Handle(msg *message.Message) error {
	return l.handle(msg)
}

func (l *auditListener) Name() string {
	return "audit"
}

// RegisterEventForAudit opts an event into audit logging. The event→Entry
// mapping is passed at registration, so opting in and defining the mapping
// are one unit and can't drift apart. Returning a nil Entry skips the event.
func RegisterEventForAudit[T any, PT interface {
	*T
	events.Event
}](toEntry func(PT) *Entry) {
	name := PT(new(T)).Name()
	events.RegisterListener(name, &auditListener{handle: func(msg *message.Message) error {
		if !license.IsFeatureEnabled(license.FeatureAuditLogs) {
			return nil // license is runtime-mutable — checked per event, not at registration
		}
		e := PT(new(T)) // fresh instance per message — handlers run concurrently
		if err := json.Unmarshal(msg.Payload, e); err != nil {
			return err
		}
		entry := toEntry(e)
		if entry == nil {
			return nil
		}
		enrichFromMetadata(entry, msg.Metadata)
		return WriteAuditEvent(entry)
	}})
}

func enrichFromMetadata(entry *Entry, meta message.Metadata) {
	entry.Source.IP = meta.Get(events.MetadataKeyIP)
	entry.Source.UserAgent = meta.Get(events.MetadataKeyUserAgent)
	entry.RequestID = meta.Get(events.MetadataKeyRequestID)
	if entry.Source.Type == "" {
		if entry.Source.IP != "" || entry.Source.UserAgent != "" {
			entry.Source.Type = SourceHTTP
		} else {
			entry.Source.Type = SourceSystem
		}
	}
}
