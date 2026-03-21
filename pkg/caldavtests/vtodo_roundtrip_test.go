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
	"strings"
	"testing"
	"time"

	ics "github.com/arran4/golang-ical"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVTodoRoundTrip(t *testing.T) {
	// Helper: PUT a VTODO, GET it back, parse the VTODO
	putAndGet := func(t *testing.T, _, path string, vtodoBody string) *ics.VTodo {
		t.Helper()
		e := setupTestEnv(t)

		rec := caldavPUT(t, e, path, vtodoBody)
		require.True(t, rec.Code >= 200 && rec.Code < 300,
			"PUT failed with status %d. Body:\n%s", rec.Code, rec.Body.String())

		rec2 := caldavGET(t, e, path)
		require.Equal(t, 200, rec2.Code, "GET failed. Body:\n%s", rec2.Body.String())

		cal := parseICalFromResponse(t, rec2)
		return getVTodo(t, cal)
	}

	t.Run("SUMMARY round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.1.12 (rfc5545.txt line 5179)
		vtodo := NewVTodo("rt-summary", "My Task Summary").Build()
		result := putAndGet(t, "rt-summary", "/dav/projects/36/rt-summary.ics", vtodo)
		assert.Equal(t, "My Task Summary", getVTodoProperty(result, ics.ComponentPropertySummary))
	})

	t.Run("DESCRIPTION round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.1.5 (rfc5545.txt line 4688)
		vtodo := NewVTodo("rt-desc", "Desc Test").
			Description("This is a detailed description").
			Build()
		result := putAndGet(t, "rt-desc", "/dav/projects/36/rt-desc.ics", vtodo)
		desc := getVTodoProperty(result, ics.ComponentPropertyDescription)
		assert.Contains(t, desc, "This is a detailed description")
	})

	t.Run("DESCRIPTION with newlines round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-desc-nl", "Desc Newline Test").
			Description("Line 1\\nLine 2\\nLine 3").
			Build()
		result := putAndGet(t, "rt-desc-nl", "/dav/projects/36/rt-desc-nl.ics", vtodo)
		desc := getVTodoProperty(result, ics.ComponentPropertyDescription)
		// Should preserve the newline structure
		assert.True(t, strings.Contains(desc, "Line 1") && strings.Contains(desc, "Line 2"),
			"Description should preserve multi-line content. Got: %s", desc)
	})

	t.Run("DUE round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.2.3 (rfc5545.txt line 5353)
		due := time.Date(2024, 6, 15, 14, 30, 0, 0, time.UTC)
		vtodo := NewVTodo("rt-due", "Due Test").Due(due).Build()
		result := putAndGet(t, "rt-due", "/dav/projects/36/rt-due.ics", vtodo)

		dueStr := getVTodoProperty(result, ics.ComponentPropertyDue)
		assert.Contains(t, dueStr, "20240615",
			"DUE date should be preserved. Got: %s", dueStr)
	})

	t.Run("DTSTART round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.2.4 (rfc5545.txt line 5415)
		start := time.Date(2024, 5, 1, 9, 0, 0, 0, time.UTC)
		vtodo := NewVTodo("rt-dtstart", "Start Test").DtStart(start).Build()
		result := putAndGet(t, "rt-dtstart", "/dav/projects/36/rt-dtstart.ics", vtodo)

		startStr := getVTodoProperty(result, ics.ComponentPropertyDtStart)
		assert.Contains(t, startStr, "20240501",
			"DTSTART should be preserved. Got: %s", startStr)
	})

	t.Run("PRIORITY round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.1.9 (rfc5545.txt line 4956)
		// CalDAV priority 1 = Vikunja priority 5 (highest)
		// The round-trip may map through Vikunja's priority system
		// Vikunja maps: CalDAV 1→Vikunja 5→CalDAV 1
		vtodo := NewVTodo("rt-priority-1", "Priority 1 Test").Priority(1).Build()
		result := putAndGet(t, "rt-priority-1", "/dav/projects/36/rt-priority-1.ics", vtodo)
		assert.Equal(t, "1", getVTodoProperty(result, ics.ComponentPropertyPriority),
			"Priority 1 (highest) should round-trip")
	})

	t.Run("PRIORITY 5 round-trips as 5", func(t *testing.T) {
		// CalDAV 5 = Vikunja 2 (medium) → CalDAV 5
		vtodo := NewVTodo("rt-priority-5", "Priority 5 Test").Priority(5).Build()
		result := putAndGet(t, "rt-priority-5", "/dav/projects/36/rt-priority-5.ics", vtodo)
		assert.Equal(t, "5", getVTodoProperty(result, ics.ComponentPropertyPriority),
			"Priority 5 (medium) should round-trip")
	})

	t.Run("PRIORITY 9 round-trips as 9", func(t *testing.T) {
		// CalDAV 9 = Vikunja 1 (low) → CalDAV 9
		vtodo := NewVTodo("rt-priority-9", "Priority 9 Test").Priority(9).Build()
		result := putAndGet(t, "rt-priority-9", "/dav/projects/36/rt-priority-9.ics", vtodo)
		assert.Equal(t, "9", getVTodoProperty(result, ics.ComponentPropertyPriority),
			"Priority 9 (lowest) should round-trip")
	})

	t.Run("PRIORITY 0 (unset) round-trips", func(t *testing.T) {
		// Priority 0 means unset — should not appear in output
		vtodo := NewVTodo("rt-priority-0", "No Priority Test").Build()
		result := putAndGet(t, "rt-priority-0", "/dav/projects/36/rt-priority-0.ics", vtodo)
		priorityStr := getVTodoProperty(result, ics.ComponentPropertyPriority)
		assert.True(t, priorityStr == "" || priorityStr == "0",
			"Priority 0 (unset) should round-trip as empty or 0. Got: %s", priorityStr)
	})

	t.Run("COMPLETED and STATUS:COMPLETED round-trip", func(t *testing.T) {
		// RFC 5545 §3.8.2.1 (rfc5545.txt line 5240)
		completed := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
		vtodo := NewVTodo("rt-completed", "Completed Test").
			Completed(completed).
			Status("COMPLETED").
			Build()
		result := putAndGet(t, "rt-completed", "/dav/projects/36/rt-completed.ics", vtodo)

		compStr := getVTodoProperty(result, ics.ComponentPropertyCompleted)
		assert.Contains(t, compStr, "20240315",
			"COMPLETED date should be preserved. Got: %s", compStr)

		statusStr := getVTodoProperty(result, ics.ComponentPropertyStatus)
		assert.Equal(t, "COMPLETED", statusStr,
			"STATUS should be COMPLETED when task is done")
	})

	t.Run("CATEGORIES round-trip", func(t *testing.T) {
		// RFC 5545 §3.8.1.2 (rfc5545.txt line 4520)
		vtodo := NewVTodo("rt-categories", "Categories Test").
			Categories("work", "urgent", "bug").
			Build()
		result := putAndGet(t, "rt-categories", "/dav/projects/36/rt-categories.ics", vtodo)

		catProp := result.GetProperty(ics.ComponentPropertyCategories)
		require.NotNil(t, catProp, "CATEGORIES property should be present")

		catStr := catProp.Value
		assert.Contains(t, catStr, "work", "Should contain 'work' category")
		assert.Contains(t, catStr, "urgent", "Should contain 'urgent' category")
		assert.Contains(t, catStr, "bug", "Should contain 'bug' category")
	})

	t.Run("Single CATEGORY round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-cat-single", "Single Category").
			Categories("solo-label").
			Build()
		result := putAndGet(t, "rt-cat-single", "/dav/projects/36/rt-cat-single.ics", vtodo)

		catProp := result.GetProperty(ics.ComponentPropertyCategories)
		require.NotNil(t, catProp, "CATEGORIES property should be present")
		assert.Contains(t, catProp.Value, "solo-label")
	})

	t.Run("VALARM with absolute trigger round-trips", func(t *testing.T) {
		// RFC 5545 §3.8.6 (rfc5545.txt line 7352)
		alarmTime := time.Date(2024, 6, 15, 8, 0, 0, 0, time.UTC)
		vtodo := NewVTodo("rt-alarm-abs", "Alarm Absolute Test").
			Due(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)).
			AlarmAbsolute(alarmTime).
			Build()
		result := putAndGet(t, "rt-alarm-abs", "/dav/projects/36/rt-alarm-abs.ics", vtodo)

		// Check that the VTODO contains a VALARM
		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "BEGIN:VALARM", "Should contain a VALARM component")
		assert.Contains(t, body, "TRIGGER", "VALARM should have a TRIGGER")
	})

	t.Run("VALARM with relative-to-start trigger round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-alarm-rel-start", "Alarm Relative Start").
			DtStart(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			AlarmRelativeStart("-PT15M").
			Build()
		result := putAndGet(t, "rt-alarm-rel-start", "/dav/projects/36/rt-alarm-rel-start.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "BEGIN:VALARM", "Should contain a VALARM component")
		assert.Contains(t, body, "TRIGGER", "VALARM should have a TRIGGER")
	})

	t.Run("VALARM with relative-to-end trigger round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-alarm-rel-end", "Alarm Relative End").
			Due(time.Date(2024, 6, 15, 12, 0, 0, 0, time.UTC)).
			AlarmRelativeEnd("-PT30M").
			Build()
		result := putAndGet(t, "rt-alarm-rel-end", "/dav/projects/36/rt-alarm-rel-end.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "BEGIN:VALARM", "Should contain a VALARM component")
	})

	t.Run("COLOR via X-APPLE-CALENDAR-COLOR round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-color", "Color Test").
			Color("#ff0000FF").
			Build()
		result := putAndGet(t, "rt-color", "/dav/projects/36/rt-color.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		// Vikunja should preserve the color in at least one of the color properties
		colorFound := strings.Contains(body, "ff0000") ||
			strings.Contains(body, "FF0000") ||
			strings.Contains(body, "#ff0000")
		assert.True(t, colorFound,
			"Color should be preserved in some form. Got:\n%s", body)
	})
}

func TestVTodoRRuleRoundTrip(t *testing.T) {
	// RFC 5545 §3.8.5.3 (rfc5545.txt line 6794)

	putAndGet := func(t *testing.T, _, path string, vtodoBody string) *ics.VTodo {
		t.Helper()
		e := setupTestEnv(t)
		rec := caldavPUT(t, e, path, vtodoBody)
		require.True(t, rec.Code >= 200 && rec.Code < 300,
			"PUT failed with %d", rec.Code)
		rec2 := caldavGET(t, e, path)
		require.Equal(t, 200, rec2.Code)
		cal := parseICalFromResponse(t, rec2)
		return getVTodo(t, cal)
	}

	t.Run("RRULE FREQ=DAILY round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-rrule-daily", "Daily Repeat").
			Due(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			Rrule("FREQ=DAILY;INTERVAL=1").
			Build()
		result := putAndGet(t, "rt-rrule-daily", "/dav/projects/36/rt-rrule-daily.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "RRULE", "Should contain RRULE")
		assert.Contains(t, body, "DAILY", "Should contain DAILY frequency")
	})

	t.Run("RRULE FREQ=WEEKLY round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-rrule-weekly", "Weekly Repeat").
			Due(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			Rrule("FREQ=WEEKLY;INTERVAL=2").
			Build()
		result := putAndGet(t, "rt-rrule-weekly", "/dav/projects/36/rt-rrule-weekly.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "RRULE", "Should contain RRULE")
		assert.Contains(t, body, "WEEKLY", "Should contain WEEKLY frequency")
	})

	t.Run("RRULE FREQ=MONTHLY round-trips", func(t *testing.T) {
		vtodo := NewVTodo("rt-rrule-monthly", "Monthly Repeat").
			Due(time.Date(2024, 6, 15, 9, 0, 0, 0, time.UTC)).
			Rrule("FREQ=MONTHLY;BYMONTHDAY=15").
			Build()
		result := putAndGet(t, "rt-rrule-monthly", "/dav/projects/36/rt-rrule-monthly.ics", vtodo)

		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		assert.Contains(t, body, "RRULE", "Should contain RRULE")
		assert.Contains(t, body, "MONTHLY", "Should contain MONTHLY frequency")
	})
}

func TestVTodoPriorityMapping(t *testing.T) {
	// RFC 5545 §3.8.1.9 (rfc5545.txt line 4956):
	// "A value of 0 specifies an undefined priority. A value of 1
	//  is the highest priority. A value of 9 is the lowest priority."
	//
	// Vikunja mapping (pkg/caldav/priority.go):
	// CalDAV 1 → Vikunja 5 → CalDAV 1 (DO NOW)
	// CalDAV 2 → Vikunja 4 → CalDAV 2 (Urgent)
	// CalDAV 3 → Vikunja 3 → CalDAV 3 (High)
	// CalDAV 4 → Vikunja 3 → CalDAV 3 (maps to High, LOSSY)
	// CalDAV 5 → Vikunja 2 → CalDAV 5 (Medium)
	// CalDAV 6-8 → Vikunja 1 → CalDAV 9 (maps to Low, LOSSY)
	// CalDAV 9 → Vikunja 1 → CalDAV 9 (Low)

	putAndGetPriority := func(t *testing.T, uid string, inputPriority int) string {
		t.Helper()
		e := setupTestEnv(t)
		vtodo := NewVTodo(uid, "Priority "+uid).Priority(inputPriority).Build()
		path := "/dav/projects/36/" + uid + ".ics"
		rec := caldavPUT(t, e, path, vtodo)
		require.True(t, rec.Code >= 200 && rec.Code < 300)
		rec2 := caldavGET(t, e, path)
		require.Equal(t, 200, rec2.Code)
		cal := parseICalFromResponse(t, rec2)
		return getVTodoProperty(getVTodo(t, cal), ics.ComponentPropertyPriority)
	}

	// Lossless round-trips
	t.Run("Priority 1 (highest) round-trips losslessly", func(t *testing.T) {
		assert.Equal(t, "1", putAndGetPriority(t, "p1", 1))
	})
	t.Run("Priority 2 round-trips losslessly", func(t *testing.T) {
		assert.Equal(t, "2", putAndGetPriority(t, "p2", 2))
	})
	t.Run("Priority 3 round-trips losslessly", func(t *testing.T) {
		assert.Equal(t, "3", putAndGetPriority(t, "p3", 3))
	})
	t.Run("Priority 5 round-trips losslessly", func(t *testing.T) {
		assert.Equal(t, "5", putAndGetPriority(t, "p5", 5))
	})
	t.Run("Priority 9 (lowest) round-trips losslessly", func(t *testing.T) {
		assert.Equal(t, "9", putAndGetPriority(t, "p9", 9))
	})

	// Lossy mappings (document the behavior)
	t.Run("Priority 4 maps to 3 (lossy)", func(t *testing.T) {
		result := putAndGetPriority(t, "p4", 4)
		assert.Equal(t, "3", result,
			"CalDAV priority 4 maps to Vikunja 3 (High), which exports as CalDAV 3")
	})
	t.Run("Priority 6 maps to 9 (lossy)", func(t *testing.T) {
		result := putAndGetPriority(t, "p6", 6)
		assert.Equal(t, "9", result,
			"CalDAV priority 6 maps to Vikunja 1 (Low), which exports as CalDAV 9")
	})
	t.Run("Priority 7 maps to 9 (lossy)", func(t *testing.T) {
		result := putAndGetPriority(t, "p7", 7)
		assert.Equal(t, "9", result)
	})
	t.Run("Priority 8 maps to 9 (lossy)", func(t *testing.T) {
		result := putAndGetPriority(t, "p8", 8)
		assert.Equal(t, "9", result)
	})
}

func TestVTodoDurationRoundTrip(t *testing.T) {
	// RFC 5545 §3.8.2.5 (rfc5545.txt line 5495):
	// "In a VTODO calendar component the property may be used to
	//  specify a positive duration of time that the to-do is expected
	//  to take for its completion."

	putAndGet := func(t *testing.T, uid, vtodoBody string) *ics.VTodo {
		t.Helper()
		e := setupTestEnv(t)
		path := "/dav/projects/36/" + uid + ".ics"
		rec := caldavPUT(t, e, path, vtodoBody)
		require.True(t, rec.Code >= 200 && rec.Code < 300)
		rec2 := caldavGET(t, e, path)
		require.Equal(t, 200, rec2.Code)
		cal := parseICalFromResponse(t, rec2)
		return getVTodo(t, cal)
	}

	t.Run("DTSTART + DURATION computes end date", func(t *testing.T) {
		// When DTSTART and DURATION are specified, Vikunja should compute
		// EndDate = DTSTART + DURATION (pkg/caldav/parsing.go:412-414)
		vtodo := NewVTodo("rt-duration", "Duration Test").
			DtStart(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			Duration("PT2H").
			Build()
		result := putAndGet(t, "rt-duration", vtodo)

		// Vikunja stores DTSTART and EndDate (DTSTART+DURATION)
		// On export, it may output DTSTART and DTEND, or DTSTART and DURATION
		body := result.Serialize(&ics.SerializationConfiguration{MaxLength: 75, PropertyMaxLength: 75, NewLine: "\r\n"})
		hasEnd := strings.Contains(body, "DTEND") || strings.Contains(body, "DURATION")
		assert.True(t, hasEnd,
			"Should preserve end time information (via DTEND or DURATION). Got:\n%s", body)
	})

	t.Run("DTSTART + DUE are both preserved", func(t *testing.T) {
		vtodo := NewVTodo("rt-start-due", "Start and Due").
			DtStart(time.Date(2024, 6, 1, 9, 0, 0, 0, time.UTC)).
			Due(time.Date(2024, 6, 15, 17, 0, 0, 0, time.UTC)).
			Build()
		result := putAndGet(t, "rt-start-due", vtodo)

		startStr := getVTodoProperty(result, ics.ComponentPropertyDtStart)
		dueStr := getVTodoProperty(result, ics.ComponentPropertyDue)
		assert.Contains(t, startStr, "20240601", "DTSTART should be preserved")
		assert.Contains(t, dueStr, "20240615", "DUE should be preserved")
	})
}
