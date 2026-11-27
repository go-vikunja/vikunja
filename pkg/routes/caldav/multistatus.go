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
	"bufio"
	"encoding/xml"
	"io"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"
)

const (
	nsDAV            = "DAV:"
	nsCalDAV         = "urn:ietf:params:xml:ns:caldav"
	nsCalendarServer = "http://calendarserver.org/ns/"
)

// Some clients only understand these exact prefixes.
var nsPrefixes = map[string]string{
	nsDAV:            "D",
	nsCalDAV:         "C",
	nsCalendarServer: "CS",
}

const rootNamespaces = ` xmlns:D="` + nsDAV + `" xmlns:C="` + nsCalDAV + `" xmlns:CS="` + nsCalendarServer + `"`

func davName(local string) xml.Name {
	return xml.Name{
		Space: nsDAV,
		Local: local,
	}
}

func calDAVName(local string) xml.Name {
	return xml.Name{
		Space: nsCalDAV,
		Local: local,
	}
}

type propValue interface {
	writeValue(w *xmlWriter)
}

type textValue string

type hrefsValue []string

type elementsValue []element

type element struct {
	name  xml.Name
	attrs []attr
}

type attr struct {
	name, value string
}

func (v textValue) writeValue(w *xmlWriter) { w.text(string(v)) }

func (v hrefsValue) writeValue(w *xmlWriter) {
	for _, href := range v {
		w.textElement(davName("href"), href)
	}
}

func (v elementsValue) writeValue(w *xmlWriter) {
	for _, el := range v {
		w.emptyElement(el.name, el.attrs)
	}
}

type prop struct {
	name xml.Name
	// nil renders an empty element.
	value propValue
}

type propstat struct {
	status int
	props  []prop
}

type msResponse interface {
	writeResponse(w *xmlWriter, minimal bool)
}

type statusResponse struct {
	href   string
	status int
}

type propstatResponse struct {
	href      string
	propstats []propstat
}

func propResponse(href string, found []prop, missing []xml.Name) propstatResponse {
	r := propstatResponse{href: href}
	if len(found) > 0 {
		r.propstats = append(r.propstats, propstat{
			status: http.StatusOK,
			props:  found,
		})
	}
	if len(missing) > 0 {
		props := make([]prop, len(missing))
		for i, name := range missing {
			props[i] = prop{name: name}
		}
		r.propstats = append(r.propstats, propstat{
			status: http.StatusNotFound,
			props:  props,
		})
	}
	return r
}

type multistatus struct {
	responses []msResponse
	syncToken string
	// Prefer: return=minimal, see draft-murchison-webdav-prefer.
	minimal bool
}

func (r statusResponse) writeResponse(w *xmlWriter, _ bool) {
	w.open(davName("response"))
	w.textElement(davName("href"), r.href)
	w.status(r.status)
	w.close(davName("response"))
}

func (r propstatResponse) writeResponse(w *xmlWriter, minimal bool) {
	propstats := slices.Clone(r.propstats)
	if minimal {
		propstats = slices.DeleteFunc(propstats, func(ps propstat) bool { return ps.status == http.StatusNotFound })
	}
	// A response needs a propstat or a status (RFC 4918 §14.24).
	if len(propstats) == 0 {
		propstats = []propstat{{status: http.StatusOK}}
	}

	w.open(davName("response"))
	w.textElement(davName("href"), r.href)
	for _, ps := range propstats {
		w.open(davName("propstat"))
		if len(ps.props) == 0 {
			w.emptyElement(davName("prop"), nil)
		} else {
			w.open(davName("prop"))
			for _, p := range ps.props {
				w.prop(p)
			}
			w.close(davName("prop"))
		}
		w.status(ps.status)
		w.close(davName("propstat"))
	}
	w.close(davName("response"))
}

func (ms multistatus) writeTo(out io.Writer) error {
	w := newXMLWriter(out)
	w.declaration()
	w.openRoot(davName("multistatus"))
	for _, r := range ms.responses {
		r.writeResponse(w, ms.minimal)
	}
	if ms.syncToken != "" {
		w.textElement(davName("sync-token"), ms.syncToken)
	}
	w.close(davName("multistatus"))
	return w.flush()
}

// prefersMinimal reports whether return=minimal is among the request's preferences
// (RFC 7240 §2: a comma-separated list, possibly over several header lines).
func prefersMinimal(h http.Header) bool {
	for _, line := range h.Values("Prefer") {
		for _, pref := range strings.Split(line, ",") {
			pref, _, _ = strings.Cut(pref, ";")
			name, value, _ := strings.Cut(pref, "=")
			if strings.EqualFold(strings.TrimSpace(name), "return") &&
				strings.EqualFold(strings.Trim(strings.TrimSpace(value), `"`), "minimal") {
				return true
			}
		}
	}
	return false
}

func writeMultistatus(c *echo.Context, ms multistatus) error {
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	if ms.minimal {
		c.Response().Header().Set("Preference-Applied", "return=minimal")
	}
	c.Response().WriteHeader(http.StatusMultiStatus)
	return ms.writeTo(c.Response())
}

func writeDAVError(c *echo.Context, status int, precondition xml.Name) error {
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(status)
	w := newXMLWriter(c.Response())
	w.declaration()
	w.openRoot(davName("error"))
	w.emptyElement(precondition, nil)
	w.close(davName("error"))
	return w.flush()
}

// xmlWriter streams straight into the response; bufio keeps the first write error.
type xmlWriter struct {
	w *bufio.Writer
}

func newXMLWriter(out io.Writer) *xmlWriter {
	return &xmlWriter{w: bufio.NewWriter(out)}
}

func (w *xmlWriter) flush() error { return w.w.Flush() }

func (w *xmlWriter) declaration() {
	_, _ = w.w.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
}

func (w *xmlWriter) openRoot(name xml.Name) {
	w.startTag(name)
	_, _ = w.w.WriteString(rootNamespaces)
	_ = w.w.WriteByte('>')
}

func (w *xmlWriter) open(name xml.Name) {
	w.startTag(name)
	_ = w.w.WriteByte('>')
}

func (w *xmlWriter) close(name xml.Name) {
	_, _ = w.w.WriteString("</")
	w.qualifiedName(name)
	_ = w.w.WriteByte('>')
}

func (w *xmlWriter) emptyElement(name xml.Name, attrs []attr) {
	w.startTag(name)
	for _, a := range attrs {
		_ = w.w.WriteByte(' ')
		_, _ = w.w.WriteString(a.name)
		_, _ = w.w.WriteString(`="`)
		w.text(a.value)
		_ = w.w.WriteByte('"')
	}
	_, _ = w.w.WriteString("/>")
}

func (w *xmlWriter) textElement(name xml.Name, text string) {
	w.open(name)
	w.text(text)
	w.close(name)
}

func (w *xmlWriter) prop(p prop) {
	if p.value == nil {
		w.emptyElement(p.name, nil)
		return
	}
	w.open(p.name)
	p.value.writeValue(w)
	w.close(p.name)
}

func (w *xmlWriter) status(code int) {
	w.textElement(davName("status"), "HTTP/1.1 "+strconv.Itoa(code)+" "+http.StatusText(code))
}

// Unprefixed elements always declare their namespace, so nesting never inherits the wrong one.
func (w *xmlWriter) startTag(name xml.Name) {
	_ = w.w.WriteByte('<')
	w.qualifiedName(name)
	if _, known := nsPrefixes[name.Space]; !known {
		_, _ = w.w.WriteString(` xmlns="`)
		w.text(name.Space)
		_ = w.w.WriteByte('"')
	}
}

func (w *xmlWriter) qualifiedName(name xml.Name) {
	if prefix, known := nsPrefixes[name.Space]; known {
		_, _ = w.w.WriteString(prefix)
		_ = w.w.WriteByte(':')
	}
	_, _ = w.w.WriteString(name.Local)
}

func (w *xmlWriter) text(s string) {
	_ = xml.EscapeText(w.w, []byte(s))
}
