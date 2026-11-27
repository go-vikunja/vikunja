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

package caldav

import (
	"encoding/xml"
	"fmt"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/models"

	ics "github.com/arran4/golang-ical"
)

const (
	maxFilterElements = 200
	// RFC 4791 §9.9: time-range bounds are UTC date-times.
	timeRangeFormat = "20060102T150405Z"
)

// RFC 4791 §9.7.2: only these properties take a time-range.
var dateTimeProperties = map[string]bool{
	"DTSTART": true, "DTEND": true, "DUE": true, "COMPLETED": true,
	"CREATED": true, "DTSTAMP": true, "LAST-MODIFIED": true, "RECURRENCE-ID": true,
}

// Top-level components cannot nest (RFC 4791 §7.8, CALDAV:valid-filter).
var topLevelComponents = map[string]bool{
	"VCALENDAR": true, "VEVENT": true, "VTODO": true, "VJOURNAL": true, "VFREEBUSY": true, "VTIMEZONE": true,
}

func invalidFilter(format string, args ...any) error {
	return preconditionError{precondition: calDAVName("valid-filter"), reason: fmt.Sprintf(format, args...)}
}

func unsupportedFilter(format string, args ...any) error {
	return preconditionError{precondition: calDAVName("supported-filter"), reason: fmt.Sprintf(format, args...)}
}

// calendarFilter is a CALDAV:filter (RFC 4791 §9.7) that passed validation: whatever it
// holds can be evaluated.
type calendarFilter struct {
	// Always named VCALENDAR.
	calendar compFilter
}

type compFilter struct {
	// Upper case, like every name below.
	name string
	test compTest
}

type compTest interface{ isCompTest() }

// isNotDefined is CALDAV:is-not-defined.
type isNotDefined struct{}

// compMatch requires the component and all of its tests; without tests it only has to exist.
type compMatch struct {
	timeRange *timeRange
	props     []propFilter
	comps     []compFilter
}

type propFilter struct {
	name string
	test propTest
}

type propTest interface{ isPropTest() }

// propMatch requires the property, its value to match value (if set) and all params.
type propMatch struct {
	value  valueTest
	params []paramFilter
}

// valueTest is a time-range or a text-match.
type valueTest interface{ matchesValue(v string) bool }

type paramFilter struct {
	name string
	test paramTest
}

type paramTest interface{ isParamTest() }

// paramMatch requires the parameter, and a value matching text if set.
type paramMatch struct{ text *textMatch }

type textMatch struct {
	value    string
	foldCase bool
	negate   bool
}

// timeRange is half-open, [start, end); a missing bound is the earliest or latest time.
type timeRange struct{ start, end time.Time }

func (isNotDefined) isCompTest()  {}
func (compMatch) isCompTest()     {}
func (isNotDefined) isPropTest()  {}
func (propMatch) isPropTest()     {}
func (isNotDefined) isParamTest() {}
func (paramMatch) isParamTest()   {}

// xmlElement is a parsed XML element the filter is read from.
type xmlElement struct {
	name     xml.Name
	attrs    []xml.Attr
	text     strings.Builder
	children []*xmlElement
}

func (e *xmlElement) attr(local string) (string, bool) {
	for _, a := range e.attrs {
		if a.Name.Space == "" && a.Name.Local == local {
			return a.Value, true
		}
	}
	return "", false
}

// Elements in other namespaces are extensions and ignored (RFC 4918 §17).
func (e *xmlElement) calDAVChildren() []*xmlElement {
	var children []*xmlElement
	for _, child := range e.children {
		if child.name.Space == nsCalDAV {
			children = append(children, child)
		}
	}
	return children
}

func (e *xmlElement) nameAttr() (string, error) {
	name, _ := e.attr("name")
	if name == "" {
		return "", invalidFilter("<%s> without a name", e.name.Local)
	}
	return strings.ToUpper(name), nil
}

func parseFilter(el *xmlElement) (calendarFilter, error) {
	children := el.calDAVChildren()
	if len(children) != 1 || children[0].name.Local != "comp-filter" {
		return calendarFilter{}, invalidFilter("a filter holds exactly one comp-filter")
	}
	calendar, err := parseCompFilter(children[0], 0)
	if err != nil {
		return calendarFilter{}, err
	}
	return calendarFilter{calendar: calendar}, nil
}

// level 0 is the calendar, 1 a component in it (VTODO), 2 one nested in that (VALARM).
func parseCompFilter(el *xmlElement, level int) (compFilter, error) {
	name, err := el.nameAttr()
	if err != nil {
		return compFilter{}, err
	}
	switch {
	case level == 0 && name != "VCALENDAR":
		return compFilter{}, invalidFilter("the outermost comp-filter must be VCALENDAR, not %s", name)
	case level == 1 && name == "VCALENDAR", level == 2 && topLevelComponents[name]:
		return compFilter{}, invalidFilter("%s cannot be nested", name)
	}

	var match compMatch
	notDefined := false
	for _, child := range el.calDAVChildren() {
		switch child.name.Local {
		case "is-not-defined":
			notDefined = true
		case "time-range":
			switch {
			case level == 0:
				return compFilter{}, invalidFilter("VCALENDAR takes no time-range")
			case level == 2:
				return compFilter{}, unsupportedFilter("time-range on %s", name)
			case match.timeRange != nil:
				return compFilter{}, invalidFilter("more than one time-range in %s", name)
			}
			tr, err := parseTimeRange(child)
			if err != nil {
				return compFilter{}, err
			}
			match.timeRange = &tr
		case "prop-filter":
			prop, err := parsePropFilter(child)
			if err != nil {
				return compFilter{}, err
			}
			match.props = append(match.props, prop)
		case "comp-filter":
			if level == 2 {
				return compFilter{}, invalidFilter("%s holds no components", name)
			}
			comp, err := parseCompFilter(child, level+1)
			if err != nil {
				return compFilter{}, err
			}
			match.comps = append(match.comps, comp)
		default:
			return compFilter{}, invalidFilter("<%s> in a comp-filter", child.name.Local)
		}
	}

	if !notDefined {
		return compFilter{name: name, test: match}, nil
	}
	if match.timeRange != nil || len(match.props) > 0 || len(match.comps) > 0 {
		return compFilter{}, invalidFilter("is-not-defined in %s with other tests", name)
	}
	return compFilter{name: name, test: isNotDefined{}}, nil
}

func parsePropFilter(el *xmlElement) (propFilter, error) {
	name, err := el.nameAttr()
	if err != nil {
		return propFilter{}, err
	}

	var match propMatch
	notDefined := false
	for _, child := range el.calDAVChildren() {
		switch child.name.Local {
		case "is-not-defined":
			notDefined = true
		case "time-range", "text-match":
			if match.value != nil {
				return propFilter{}, invalidFilter("more than one time-range or text-match in %s", name)
			}
			switch {
			case child.name.Local == "text-match":
				match.value, err = parseTextMatch(child)
			case !dateTimeProperties[name]:
				return propFilter{}, invalidFilter("%s takes no time-range", name)
			default:
				match.value, err = parseTimeRange(child)
			}
			if err != nil {
				return propFilter{}, err
			}
		case "param-filter":
			param, err := parseParamFilter(child)
			if err != nil {
				return propFilter{}, err
			}
			match.params = append(match.params, param)
		default:
			return propFilter{}, invalidFilter("<%s> in a prop-filter", child.name.Local)
		}
	}

	if !notDefined {
		return propFilter{name: name, test: match}, nil
	}
	if match.value != nil || len(match.params) > 0 {
		return propFilter{}, invalidFilter("is-not-defined in %s with other tests", name)
	}
	return propFilter{name: name, test: isNotDefined{}}, nil
}

func parseParamFilter(el *xmlElement) (paramFilter, error) {
	name, err := el.nameAttr()
	if err != nil {
		return paramFilter{}, err
	}

	var match paramMatch
	notDefined := false
	for _, child := range el.calDAVChildren() {
		switch {
		case child.name.Local == "is-not-defined":
			notDefined = true
		case child.name.Local == "text-match" && match.text == nil:
			text, err := parseTextMatch(child)
			if err != nil {
				return paramFilter{}, err
			}
			match.text = &text
		default:
			return paramFilter{}, invalidFilter("<%s> in a param-filter", child.name.Local)
		}
	}

	if !notDefined {
		return paramFilter{name: name, test: match}, nil
	}
	if match.text != nil {
		return paramFilter{}, invalidFilter("is-not-defined in %s with a text-match", name)
	}
	return paramFilter{name: name, test: isNotDefined{}}, nil
}

func parseTextMatch(el *xmlElement) (textMatch, error) {
	match := textMatch{
		// caldav-go trimmed it too: pretty-printed bodies indent the text.
		value:    strings.TrimSpace(el.text.String()),
		foldCase: true,
	}

	// RFC 4791 §7.5.1: i;ascii-casemap is the default, and these two are required.
	switch collation, _ := el.attr("collation"); collation {
	case "", "i;ascii-casemap":
	case "i;octet":
		match.foldCase = false
	default:
		return textMatch{}, preconditionError{precondition: calDAVName("supported-collation"), reason: collation}
	}

	switch negate, _ := el.attr("negate-condition"); negate {
	case "", "no":
	case "yes":
		match.negate = true
	default:
		return textMatch{}, invalidFilter("negate-condition %q", negate)
	}

	if len(el.calDAVChildren()) > 0 {
		return textMatch{}, invalidFilter("text-match holds elements")
	}
	return match, nil
}

var latest = time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)

func parseTimeRange(el *xmlElement) (timeRange, error) {
	tr := timeRange{end: latest}
	start, hasStart := el.attr("start")
	end, hasEnd := el.attr("end")
	if !hasStart && !hasEnd {
		return timeRange{}, invalidFilter("time-range without start or end")
	}
	var err error
	if hasStart {
		if tr.start, err = time.Parse(timeRangeFormat, start); err != nil {
			return timeRange{}, invalidFilter("time-range start %q", start)
		}
	}
	if hasEnd {
		if tr.end, err = time.Parse(timeRangeFormat, end); err != nil {
			return timeRange{}, invalidFilter("time-range end %q", end)
		}
	}
	if !tr.end.After(tr.start) {
		return timeRange{}, invalidFilter("time-range ends before it starts")
	}
	return tr, nil
}

// matchesValue is RFC 4791 §9.9 for a DATE-TIME property.
func (r timeRange) matchesValue(v string) bool {
	t, err := time.Parse(timeRangeFormat, v)
	if err != nil {
		return false
	}
	return !t.Before(r.start) && t.Before(r.end)
}

func (m textMatch) matchesValue(v string) bool {
	needle := m.value
	if m.foldCase {
		v, needle = asciiUpper(v), asciiUpper(needle)
	}
	return strings.Contains(v, needle) != m.negate
}

// i;ascii-casemap (RFC 4790 §9.2) folds ASCII letters only.
func asciiUpper(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' {
			return r - 'a' + 'A'
		}
		return r
	}, s)
}

// overlapsTodo is the VTODO table of RFC 4791 §9.9, read from the fields the VTODO is
// rendered from: DTSTART is the start date, DUE the due date, COMPLETED the done date of a
// done task and CREATED the creation date, each only if rendered. The table's DURATION is
// the end date, if it lies after the start and there is no due date; Vikunja stores the
// next occurrence of a repeating task only, so that is the instance that counts.
func (r timeRange) overlapsTodo(t *models.Task) bool {
	rendered := func(at time.Time) bool { return at.Unix() > 0 }
	hasStart := rendered(t.StartDate)
	hasDue := rendered(t.DueDate)
	hasDuration := hasStart && !hasDue && t.EndDate.After(t.StartDate)
	hasCompleted := t.Done && rendered(t.DoneAt)
	hasCreated := rendered(t.Created)

	start, due, end := t.StartDate, t.DueDate, t.EndDate
	completed, created := t.DoneAt, t.Created
	// a <= b
	atMost := func(a, b time.Time) bool { return !a.After(b) }

	switch {
	case hasDuration:
		return atMost(r.start, end) && (r.end.After(start) || atMost(end, r.end))
	case hasStart && hasDue:
		return (r.start.Before(due) || atMost(r.start, start)) && (r.end.After(start) || atMost(due, r.end))
	case hasStart:
		return atMost(r.start, start) && r.end.After(start)
	case hasDue:
		return r.start.Before(due) && atMost(due, r.end)
	case hasCompleted && hasCreated:
		return (atMost(r.start, created) || atMost(r.start, completed)) && (atMost(created, r.end) || atMost(completed, r.end))
	case hasCompleted:
		return atMost(r.start, completed) && atMost(completed, r.end)
	case hasCreated:
		return r.end.After(created)
	}
	return true
}

// queryCandidate is a task a calendar-query looks at. Its rendering is only parsed if the
// filter looks at properties.
type queryCandidate struct {
	resource *taskResource
	parsed   bool
	// nil if the rendering does not parse.
	calendar *icalComponent
}

// icalComponent is a rendered component, as far as a filter can look into it.
type icalComponent struct {
	name  string
	props []ics.BaseProperty
	subs  []*icalComponent
}

func (q *queryCandidate) rendered() *icalComponent {
	if q.parsed {
		return q.calendar
	}
	q.parsed = true
	cal, err := ics.ParseCalendar(strings.NewReader(q.resource.content()))
	if err != nil {
		return nil
	}
	q.calendar = &icalComponent{name: "VCALENDAR"}
	for _, p := range cal.CalendarProperties {
		q.calendar.props = append(q.calendar.props, p.BaseProperty)
	}
	for _, c := range cal.Components {
		q.calendar.subs = append(q.calendar.subs, toICalComponent(c))
	}
	return q.calendar
}

func toICalComponent(c ics.Component) *icalComponent {
	comp := &icalComponent{}
	switch c := c.(type) {
	case *ics.VTodo:
		comp.name = "VTODO"
	case *ics.VAlarm:
		comp.name = "VALARM"
	case *ics.GeneralComponent:
		comp.name = strings.ToUpper(c.Token)
	}
	for _, p := range c.UnknownPropertiesIANAProperties() {
		comp.props = append(comp.props, p.BaseProperty)
	}
	for _, sub := range c.SubComponents() {
		comp.subs = append(comp.subs, toICalComponent(sub))
	}
	return comp
}

func (f calendarFilter) matches(q *queryCandidate) bool {
	match, defined := f.calendar.test.(compMatch)
	if !defined {
		return false
	}
	if len(match.props) > 0 {
		cal := q.rendered()
		if cal == nil || !allProps(match.props, cal.props) {
			return false
		}
	}
	for _, comp := range match.comps {
		if !q.matchesComponent(comp) {
			return false
		}
	}
	return true
}

// A calendar object holds exactly one component: the task's VTODO.
func (q *queryCandidate) matchesComponent(f compFilter) bool {
	match, defined := f.test.(compMatch)
	if f.name != "VTODO" {
		return !defined
	}
	if !defined {
		return false
	}
	if match.timeRange != nil && !match.timeRange.overlapsTodo(q.resource.task) {
		return false
	}
	if len(match.props) == 0 && len(match.comps) == 0 {
		return true
	}

	cal := q.rendered()
	if cal == nil {
		return false
	}
	for _, todo := range cal.subs {
		if todo.name == "VTODO" {
			return allProps(match.props, todo.props) && allSubComponents(match.comps, todo.subs)
		}
	}
	return false
}

// RFC 4791 §9.7.1: a nested comp-filter matches if any instance of the component does.
func allSubComponents(filters []compFilter, subs []*icalComponent) bool {
	for _, f := range filters {
		match, defined := f.test.(compMatch)
		found := false
		for _, sub := range subs {
			if sub.name == f.name && (!defined || allProps(match.props, sub.props)) {
				found = true
				break
			}
		}
		if found != defined {
			return false
		}
	}
	return true
}

func allProps(filters []propFilter, props []ics.BaseProperty) bool {
	for _, f := range filters {
		match, defined := f.test.(propMatch)
		found := false
		for _, p := range props {
			if strings.EqualFold(p.IANAToken, f.name) && (!defined || match.matches(p)) {
				found = true
				break
			}
		}
		if found != defined {
			return false
		}
	}
	return true
}

func (m propMatch) matches(p ics.BaseProperty) bool {
	if m.value != nil && !m.value.matchesValue(p.Value) {
		return false
	}
	for _, f := range m.params {
		match, defined := f.test.(paramMatch)
		found := false
		for name, values := range p.ICalParameters {
			if !strings.EqualFold(name, f.name) {
				continue
			}
			for _, v := range values {
				if !defined || match.text == nil || match.text.matchesValue(v) {
					found = true
				}
			}
		}
		if found != defined {
			return false
		}
	}
	return true
}
