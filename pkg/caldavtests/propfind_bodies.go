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

// PROPFIND request bodies used by CalDAV clients.

// PropfindCurrentUserPrincipal requests the current-user-principal property.
// RFC 5397 §3
const PropfindCurrentUserPrincipal = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:current-user-principal/>
  </D:prop>
</D:propfind>`

// PropfindCalendarHomeSet requests the calendar-home-set property.
// RFC 4791 §6.2.1
const PropfindCalendarHomeSet = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>
    <C:calendar-home-set/>
  </D:prop>
</D:propfind>`

// PropfindCalendarCollectionProperties requests common calendar collection properties.
// RFC 4791 §5.2
const PropfindCalendarCollectionProperties = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:CS="http://calendarserver.org/ns/" xmlns:IC="http://apple.com/ns/ical/">
  <D:prop>
    <D:displayname/>
    <D:resourcetype/>
    <D:getetag/>
    <CS:getctag/>
    <C:supported-calendar-component-set/>
    <C:calendar-description/>
  </D:prop>
</D:propfind>`

// PropfindResourceProperties requests properties of a calendar resource (task).
const PropfindResourceProperties = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">
  <D:prop>
    <D:getetag/>
    <C:calendar-data/>
  </D:prop>
</D:propfind>`

// PropfindAllProps requests all properties (allprop).
// RFC 4918 §9.1
const PropfindAllProps = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
  <D:allprop/>
</D:propfind>`

// PropfindCurrentUserPrivilegeSet requests the current-user-privilege-set property.
// RFC 3744 §5.4
const PropfindCurrentUserPrivilegeSet = `<?xml version="1.0" encoding="utf-8" ?>
<D:propfind xmlns:D="DAV:">
  <D:prop>
    <D:current-user-privilege-set/>
  </D:prop>
</D:propfind>`

// ReportCalendarQuery is a calendar-query REPORT requesting all VTODOs.
// RFC 4791 §7.8
const ReportCalendarQuery = `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-query xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:D="DAV:">
  <D:prop>
    <D:getetag/>
    <C:calendar-data/>
  </D:prop>
  <C:filter>
    <C:comp-filter name="VCALENDAR">
      <C:comp-filter name="VTODO"/>
    </C:comp-filter>
  </C:filter>
</C:calendar-query>`

// ReportCalendarMultiget builds a calendar-multiget REPORT for specific hrefs.
// RFC 4791 §7.9
func ReportCalendarMultiget(hrefs ...string) string {
	var hrefXML string
	for _, href := range hrefs {
		hrefXML += "  <D:href>" + href + "</D:href>\n"
	}
	return `<?xml version="1.0" encoding="utf-8" ?>
<C:calendar-multiget xmlns:C="urn:ietf:params:xml:ns:caldav" xmlns:D="DAV:">
  <D:prop>
    <D:getetag/>
    <C:calendar-data/>
  </D:prop>
` + hrefXML + `</C:calendar-multiget>`
}
