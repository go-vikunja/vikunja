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

package caldavtests

import (
	"fmt"
	"strings"
	"time"
)

// VTodoBuilder constructs VCALENDAR/VTODO strings for test requests.
type VTodoBuilder struct {
	uid             string
	summary         string
	description     string
	priority        int
	due             time.Time
	dtstart         time.Time
	completed       time.Time
	status          string
	categories      []string
	relatedTo       []relatedToEntry
	alarms          []alarmEntry
	rrule           string
	color           string
	percentComplete int
	sequence        int
	duration        string
	dtstamp         time.Time
	created         time.Time
	lastMod         time.Time
	extraProps      []string
}

type relatedToEntry struct {
	reltype string // "PARENT", "CHILD", or ""
	uid     string
}

type alarmEntry struct {
	trigger     string
	action      string
	description string
}

// NewVTodo starts building a VTODO with required fields.
func NewVTodo(uid, summary string) *VTodoBuilder {
	return &VTodoBuilder{
		uid:     uid,
		summary: summary,
		dtstamp: time.Now().UTC(),
		created: time.Now().UTC(),
		lastMod: time.Now().UTC(),
	}
}

func (b *VTodoBuilder) Description(d string) *VTodoBuilder     { b.description = d; return b }
func (b *VTodoBuilder) Priority(p int) *VTodoBuilder           { b.priority = p; return b }
func (b *VTodoBuilder) Due(t time.Time) *VTodoBuilder          { b.due = t; return b }
func (b *VTodoBuilder) DtStart(t time.Time) *VTodoBuilder      { b.dtstart = t; return b }
func (b *VTodoBuilder) Completed(t time.Time) *VTodoBuilder    { b.completed = t; return b }
func (b *VTodoBuilder) Status(s string) *VTodoBuilder          { b.status = s; return b }
func (b *VTodoBuilder) Categories(c ...string) *VTodoBuilder   { b.categories = c; return b }
func (b *VTodoBuilder) Rrule(r string) *VTodoBuilder           { b.rrule = r; return b }
func (b *VTodoBuilder) Color(c string) *VTodoBuilder           { b.color = c; return b }
func (b *VTodoBuilder) Sequence(s int) *VTodoBuilder           { b.sequence = s; return b }
func (b *VTodoBuilder) Duration(d string) *VTodoBuilder        { b.duration = d; return b }
func (b *VTodoBuilder) DtStamp(t time.Time) *VTodoBuilder      { b.dtstamp = t; return b }
func (b *VTodoBuilder) Created(t time.Time) *VTodoBuilder      { b.created = t; return b }
func (b *VTodoBuilder) LastModified(t time.Time) *VTodoBuilder { b.lastMod = t; return b }
func (b *VTodoBuilder) PercentComplete(p int) *VTodoBuilder    { b.percentComplete = p; return b }
func (b *VTodoBuilder) ExtraProp(line string) *VTodoBuilder {
	b.extraProps = append(b.extraProps, line)
	return b
}

func (b *VTodoBuilder) RelatedToParent(uid string) *VTodoBuilder {
	b.relatedTo = append(b.relatedTo, relatedToEntry{reltype: "PARENT", uid: uid})
	return b
}

func (b *VTodoBuilder) RelatedToChild(uid string) *VTodoBuilder {
	b.relatedTo = append(b.relatedTo, relatedToEntry{reltype: "CHILD", uid: uid})
	return b
}

func (b *VTodoBuilder) AlarmAbsolute(triggerTime time.Time) *VTodoBuilder {
	b.alarms = append(b.alarms, alarmEntry{
		trigger:     "TRIGGER;VALUE=DATE-TIME:" + formatTime(triggerTime),
		action:      "DISPLAY",
		description: b.summary,
	})
	return b
}

func (b *VTodoBuilder) AlarmRelativeStart(duration string) *VTodoBuilder {
	b.alarms = append(b.alarms, alarmEntry{
		trigger:     "TRIGGER;RELATED=START:" + duration,
		action:      "DISPLAY",
		description: b.summary,
	})
	return b
}

func (b *VTodoBuilder) AlarmRelativeEnd(duration string) *VTodoBuilder {
	b.alarms = append(b.alarms, alarmEntry{
		trigger:     "TRIGGER;RELATED=END:" + duration,
		action:      "DISPLAY",
		description: b.summary,
	})
	return b
}

func formatTime(t time.Time) string {
	return t.UTC().Format("20060102T150405Z")
}

// Build returns the complete VCALENDAR string wrapping the VTODO.
func (b *VTodoBuilder) Build() string {
	var sb strings.Builder

	sb.WriteString("BEGIN:VCALENDAR\r\n")
	sb.WriteString("VERSION:2.0\r\n")
	sb.WriteString("PRODID:-//Test//Test//EN\r\n")
	sb.WriteString("BEGIN:VTODO\r\n")
	fmt.Fprintf(&sb, "UID:%s\r\n", b.uid)
	fmt.Fprintf(&sb, "DTSTAMP:%s\r\n", formatTime(b.dtstamp))
	fmt.Fprintf(&sb, "SUMMARY:%s\r\n", b.summary)
	fmt.Fprintf(&sb, "CREATED:%s\r\n", formatTime(b.created))
	fmt.Fprintf(&sb, "LAST-MODIFIED:%s\r\n", formatTime(b.lastMod))

	if b.description != "" {
		fmt.Fprintf(&sb, "DESCRIPTION:%s\r\n", b.description)
	}
	if b.priority > 0 {
		fmt.Fprintf(&sb, "PRIORITY:%d\r\n", b.priority)
	}
	if !b.due.IsZero() {
		fmt.Fprintf(&sb, "DUE:%s\r\n", formatTime(b.due))
	}
	if !b.dtstart.IsZero() {
		fmt.Fprintf(&sb, "DTSTART:%s\r\n", formatTime(b.dtstart))
	}
	if !b.completed.IsZero() {
		fmt.Fprintf(&sb, "COMPLETED:%s\r\n", formatTime(b.completed))
	}
	if b.status != "" {
		fmt.Fprintf(&sb, "STATUS:%s\r\n", b.status)
	}
	if len(b.categories) > 0 {
		fmt.Fprintf(&sb, "CATEGORIES:%s\r\n", strings.Join(b.categories, ","))
	}
	if b.rrule != "" {
		fmt.Fprintf(&sb, "RRULE:%s\r\n", b.rrule)
	}
	if b.color != "" {
		fmt.Fprintf(&sb, "X-APPLE-CALENDAR-COLOR:%s\r\n", b.color)
	}
	if b.percentComplete > 0 {
		fmt.Fprintf(&sb, "PERCENT-COMPLETE:%d\r\n", b.percentComplete)
	}
	if b.sequence > 0 {
		fmt.Fprintf(&sb, "SEQUENCE:%d\r\n", b.sequence)
	}
	if b.duration != "" {
		fmt.Fprintf(&sb, "DURATION:%s\r\n", b.duration)
	}
	for _, rel := range b.relatedTo {
		if rel.reltype != "" {
			fmt.Fprintf(&sb, "RELATED-TO;RELTYPE=%s:%s\r\n", rel.reltype, rel.uid)
		} else {
			fmt.Fprintf(&sb, "RELATED-TO:%s\r\n", rel.uid)
		}
	}
	for _, alarm := range b.alarms {
		sb.WriteString("BEGIN:VALARM\r\n")
		sb.WriteString(alarm.trigger + "\r\n")
		fmt.Fprintf(&sb, "ACTION:%s\r\n", alarm.action)
		fmt.Fprintf(&sb, "DESCRIPTION:%s\r\n", alarm.description)
		sb.WriteString("END:VALARM\r\n")
	}
	for _, prop := range b.extraProps {
		sb.WriteString(prop + "\r\n")
	}
	sb.WriteString("END:VTODO\r\n")
	sb.WriteString("END:VCALENDAR\r\n")

	return sb.String()
}
