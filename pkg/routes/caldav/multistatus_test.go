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
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const msHead = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
	`<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:CS="http://calendarserver.org/ns/">`

func renderMultistatus(t *testing.T, ms multistatus) string {
	t.Helper()
	var sb strings.Builder
	require.NoError(t, ms.writeTo(&sb))
	out := sb.String()

	// Every output must be well-formed.
	d := xml.NewDecoder(strings.NewReader(out))
	for {
		_, err := d.Token()
		if err != nil {
			require.ErrorContains(t, err, "EOF", out)
			break
		}
	}
	return out
}

func TestMultistatusWriter(t *testing.T) {
	t.Run("escapes text and hrefs", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{responses: []msResponse{
			propResponse(`/dav/projects/1/a&b<c"d.ics`, []prop{
				{
					name:  davName("displayname"),
					value: textValue(`Tom & "Jerry" <3`),
				},
				{
					name:  davName("owner"),
					value: hrefsValue{`/dav/principals/a&b/`},
				},
			}, nil),
		}})

		assert.Equal(t, msHead+
			`<D:response><D:href>/dav/projects/1/a&amp;b&lt;c&#34;d.ics</D:href>`+
			`<D:propstat><D:prop>`+
			`<D:displayname>Tom &amp; &#34;Jerry&#34; &lt;3</D:displayname>`+
			`<D:owner><D:href>/dav/principals/a&amp;b/</D:href></D:owner>`+
			`</D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>`+
			`</D:response></D:multistatus>`, out)

		var parsed struct {
			Href string `xml:"response>href"`
			Name string `xml:"response>propstat>prop>displayname"`
		}
		require.NoError(t, xml.Unmarshal([]byte(out), &parsed))
		assert.Equal(t, `/dav/projects/1/a&b<c"d.ics`, parsed.Href)
		assert.Equal(t, `Tom & "Jerry" <3`, parsed.Name)
	})

	t.Run("propstats with the same status stay separate and in order", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{responses: []msResponse{
			propstatResponse{
				href: "/x",
				propstats: []propstat{
					{
						status: http.StatusForbidden,
						props:  []prop{{name: davName("b")}},
					},
					{
						status: http.StatusForbidden,
						props:  []prop{{name: davName("a")}},
					},
				},
			},
		}})

		assert.Equal(t, msHead+`<D:response><D:href>/x</D:href>`+
			`<D:propstat><D:prop><D:b/></D:prop><D:status>HTTP/1.1 403 Forbidden</D:status></D:propstat>`+
			`<D:propstat><D:prop><D:a/></D:prop><D:status>HTTP/1.1 403 Forbidden</D:status></D:propstat>`+
			`</D:response></D:multistatus>`, out)
	})

	t.Run("minimal drops 404 propstats", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{
			minimal: true,
			responses: []msResponse{
				propResponse("/found", []prop{{
					name:  davName("getetag"),
					value: textValue("e"),
				}}, []xml.Name{davName("missing")}),
				propResponse("/nothing", nil, []xml.Name{davName("missing")}),
			},
		})

		assert.Equal(t, msHead+
			`<D:response><D:href>/found</D:href>`+
			`<D:propstat><D:prop><D:getetag>e</D:getetag></D:prop><D:status>HTTP/1.1 200 OK</D:status></D:propstat>`+
			`</D:response>`+
			`<D:response><D:href>/nothing</D:href>`+
			`<D:propstat><D:prop/><D:status>HTTP/1.1 200 OK</D:status></D:propstat>`+
			`</D:response></D:multistatus>`, out)
	})

	t.Run("properties in other namespaces keep their namespace", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{responses: []msResponse{
			propResponse("/x", nil, []xml.Name{
				{
					Space: "urn:x-vikunja-unknown",
					Local: "thing",
				},
				{
					Space: `urn:"quoted"`,
					Local: "q",
				},
				{Local: "bare"},
				{
					Space: nsCalendarServer,
					Local: "getctag",
				},
			}),
		}})

		assert.Contains(t, out, `<thing xmlns="urn:x-vikunja-unknown"/>`)
		assert.Contains(t, out, `<q xmlns="urn:&#34;quoted&#34;"/>`)
		assert.Contains(t, out, `<bare xmlns=""/>`)
		assert.Contains(t, out, `<CS:getctag/>`)

		var parsed struct {
			Prop struct {
				Any []struct {
					XMLName xml.Name
				} `xml:",any"`
			} `xml:"response>propstat>prop"`
		}
		require.NoError(t, xml.Unmarshal([]byte(out), &parsed))
		props := parsed.Prop.Any
		require.Len(t, props, 4)
		assert.Equal(t, xml.Name{
			Space: "urn:x-vikunja-unknown",
			Local: "thing",
		}, props[0].XMLName)
		assert.Equal(t, xml.Name{
			Space: `urn:"quoted"`,
			Local: "q",
		}, props[1].XMLName)
		assert.Equal(t, xml.Name{Local: "bare"}, props[2].XMLName)
		assert.Equal(t, xml.Name{
			Space: nsCalendarServer,
			Local: "getctag",
		}, props[3].XMLName)
	})

	t.Run("child elements with attributes", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{responses: []msResponse{
			propResponse("/x", []prop{
				{
					name: calDAVName("supported-calendar-component-set"),
					value: elementsValue{
						{
							name: calDAVName("comp"),
							attrs: []attr{{
								name:  "name",
								value: "VTODO",
							}},
						},
						{
							name: calDAVName("comp"),
							attrs: []attr{{
								name:  "name",
								value: `a"&`,
							}},
						},
					},
				},
				{
					name: davName("resourcetype"),
					value: elementsValue{
						{name: davName("collection")},
						{name: calDAVName("calendar")},
					},
				},
				{name: davName("resourcetype")},
			}, nil),
		}})

		assert.Contains(t, out, `<C:supported-calendar-component-set><C:comp name="VTODO"/><C:comp name="a&#34;&amp;"/></C:supported-calendar-component-set>`)
		assert.Contains(t, out, `<D:resourcetype><D:collection/><C:calendar/></D:resourcetype>`)
		assert.Contains(t, out, `<D:resourcetype/>`)
	})

	t.Run("status-only response and sync-token", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{
			syncToken: `data:,"36-1&2"`,
			responses: []msResponse{
				statusResponse{
					href:   "/dav/projects/36/gone.ics",
					status: http.StatusNotFound,
				},
			},
		})

		assert.Equal(t, msHead+
			`<D:response><D:href>/dav/projects/36/gone.ics</D:href><D:status>HTTP/1.1 404 Not Found</D:status></D:response>`+
			`<D:sync-token>data:,&#34;36-1&amp;2&#34;</D:sync-token>`+
			`</D:multistatus>`, out)
	})

	t.Run("response without propstats gets an empty 200 propstat", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{responses: []msResponse{propResponse("/x", nil, nil)}})

		assert.Equal(t, msHead+`<D:response><D:href>/x</D:href>`+
			`<D:propstat><D:prop/><D:status>HTTP/1.1 200 OK</D:status></D:propstat>`+
			`</D:response></D:multistatus>`, out)
	})

	t.Run("minimal leaves status-only responses alone", func(t *testing.T) {
		out := renderMultistatus(t, multistatus{
			minimal: true,
			responses: []msResponse{
				statusResponse{
					href:   "/gone",
					status: http.StatusNotFound,
				},
			},
		})

		assert.Equal(t, msHead+`<D:response><D:href>/gone</D:href><D:status>HTTP/1.1 404 Not Found</D:status></D:response></D:multistatus>`, out)
	})

	t.Run("empty multistatus", func(t *testing.T) {
		assert.Equal(t, msHead+`</D:multistatus>`, renderMultistatus(t, multistatus{}))
	})
}

func TestWriteDAVError(t *testing.T) {
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("REPORT", "/", nil), rec)

	require.NoError(t, writeDAVError(c, http.StatusForbidden, davName("valid-sync-token")))

	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Equal(t, "application/xml; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, `<?xml version="1.0" encoding="UTF-8"?>`+"\n"+
		`<D:error xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:CS="http://calendarserver.org/ns/"><D:valid-sync-token/></D:error>`,
		rec.Body.String())
}

func TestWriteMultistatus(t *testing.T) {
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("PROPFIND", "/", nil), rec)

	require.NoError(t, writeMultistatus(c, multistatus{}))

	assert.Equal(t, http.StatusMultiStatus, rec.Code)
	assert.Equal(t, "application/xml; charset=utf-8", rec.Header().Get("Content-Type"))
	assert.Equal(t, msHead+`</D:multistatus>`, rec.Body.String())
	assert.Empty(t, rec.Header().Get("Preference-Applied"))
}

func TestWriteMultistatusMinimal(t *testing.T) {
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(httptest.NewRequest("PROPFIND", "/", nil), rec)

	require.NoError(t, writeMultistatus(c, multistatus{minimal: true}))

	assert.Equal(t, "return=minimal", rec.Header().Get("Preference-Applied"))
}
