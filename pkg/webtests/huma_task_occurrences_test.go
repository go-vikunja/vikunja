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

package webtests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The occurrence query (tasks-occurrences) and the preview (tasks-occurrences-preview) through
// real HTTP requests on /api/v2: the query against completion, access, count bounds,
// the preview's rule forms, refusal parity with task create, time zones and API tokens.
const (
	occPreviewPath = "/api/v2/tasks/occurrences/preview"
	occCreatePath  = "/api/v2/projects/1/tasks"
	occAnchor      = "2027-03-01T09:00:00Z" // a Monday
)

type occAnswer struct {
	Occurrences []time.Time `json:"occurrences"`
	Reason      string      `json:"reason"`
}

func occPath(taskID int64, query string) string {
	path := fmt.Sprintf("/api/v2/tasks/%d/occurrences", taskID)
	if query != "" {
		path += "?" + query
	}
	return path
}

// occDecode asserts the status and decodes an occurrence answer. occurrences is always a list,
// never null or absent.
func occDecode(t *testing.T, rec *httptest.ResponseRecorder, want int) occAnswer {
	t.Helper()
	require.Equal(t, want, rec.Code, rec.Body.String())
	raw := map[string]json.RawMessage{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw), rec.Body.String())
	require.Contains(t, raw, "occurrences", rec.Body.String())
	assert.NotEqual(t, "null", string(raw["occurrences"]), "occurrences is [] when there are none")
	answer := occAnswer{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &answer))
	return answer
}

// occCreate creates a task in project 1 through v2 and returns its id.
func occCreate(t *testing.T, e *echo.Echo, token string, body map[string]any) int64 {
	t.Helper()
	if _, ok := body["title"]; !ok {
		body["title"] = "recurring"
	}
	created := ruleTextDo(t, e, http.MethodPost, occCreatePath, marshal(t, body), token, "", http.StatusCreated)
	return int64(created["id"].(float64))
}

func occPreview(t *testing.T, e *echo.Echo, token string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	return humaRequest(t, e, http.MethodPost, occPreviewPath, marshal(t, body), token, "")
}

// occProblem decodes an error body's status, Vikunja code and detail.
func occProblem(t *testing.T, rec *httptest.ResponseRecorder) (status, code int, detail string) {
	t.Helper()
	var problem struct {
		Code   int    `json:"code"`
		Detail string `json:"detail"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &problem), rec.Body.String())
	return rec.Code, problem.Code, problem.Detail
}

func occDueDate(t *testing.T, e *echo.Echo, token string, taskID int64) time.Time {
	t.Helper()
	body := ruleTextRead(t, e, token, fmt.Sprintf("/api/v2/tasks/%d", taskID))
	due, err := time.Parse(time.RFC3339, body["due_date"].(string))
	require.NoError(t, err)
	return due
}

func assertSameInstants(t *testing.T, want, got []time.Time) {
	t.Helper()
	require.Len(t, got, len(want))
	for i := range want {
		assert.True(t, want[i].Equal(got[i]), "occurrence %d: want %s, got %s", i, want[i], got[i])
	}
}

// TestTaskOccurrences_MatchCompletion over HTTP: the listed dates are the due dates
// three completions through PATCH store, in order, COUNT decrement included.
func TestTaskOccurrences_MatchCompletion(t *testing.T) {
	e, token := ruleTextSetup(t, "")
	id := occCreate(t, e, token, map[string]any{"due_date": occAnchor, "repeats_rrule": "FREQ=WEEKLY;BYDAY=MO,TH;COUNT=5"})

	answer := occDecode(t, humaRequest(t, e, http.MethodGet, occPath(id, "count=3"), "", token, ""), http.StatusOK)
	require.Len(t, answer.Occurrences, 3)
	assert.Empty(t, answer.Reason)
	for i := 1; i < len(answer.Occurrences); i++ {
		assert.True(t, answer.Occurrences[i-1].Before(answer.Occurrences[i]), "ascending: %v", answer.Occurrences)
	}

	completed := make([]time.Time, 0, 3)
	for range 3 {
		ruleTextDo(t, e, http.MethodPatch, fmt.Sprintf("/api/v2/tasks/%d", id), `{"done":true}`, token, ruleTextMerge, http.StatusOK)
		completed = append(completed, occDueDate(t, e, token, id))
	}
	assertSameInstants(t, answer.Occurrences, completed)
	assert.True(t, completed[0].After(time.Date(2027, 3, 1, 9, 0, 0, 0, time.UTC)), "the first occurrence follows the anchor")
	assert.Contains(t, storedRule(t, id), "COUNT=2", "the query left the task alone; only the completions moved it")
}

// TestTaskOccurrences_Reasons: a task without a rule and a done task list [] with their reason; a
// rule whose COUNT runs out lists what is left.
func TestTaskOccurrences_Reasons(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	noRule := occCreate(t, e, token, map[string]any{"due_date": occAnchor})
	rec := humaRequest(t, e, http.MethodGet, occPath(noRule, ""), "", token, "")
	assert.Contains(t, rec.Body.String(), `"occurrences":[]`)
	assert.Equal(t, occAnswer{Occurrences: []time.Time{}, Reason: "no_rule"}, occDecode(t, rec, http.StatusOK))

	done := occCreate(t, e, token, map[string]any{"due_date": occAnchor, "repeats_rrule": "FREQ=DAILY", "done": true})
	assert.Equal(t, occAnswer{Occurrences: []time.Time{}, Reason: "task_done"},
		occDecode(t, humaRequest(t, e, http.MethodGet, occPath(done, ""), "", token, ""), http.StatusOK))

	counted := occCreate(t, e, token, map[string]any{"due_date": occAnchor, "repeats_rrule": "FREQ=DAILY;COUNT=3"})
	answer := occDecode(t, humaRequest(t, e, http.MethodGet, occPath(counted, "count=5"), "", token, ""), http.StatusOK)
	assert.Len(t, answer.Occurrences, 2, "COUNT=3 includes the current date")
	assert.Equal(t, "count_exhausted", answer.Reason)
}

// TestTaskOccurrences_Count: count defaults to 10, accepts 1 and 100 and refuses 0 and 101 with 422.
func TestTaskOccurrences_Count(t *testing.T) {
	e, token := ruleTextSetup(t, "")
	id := occCreate(t, e, token, map[string]any{"due_date": occAnchor, "repeats_rrule": "FREQ=DAILY"})

	assert.Len(t, occDecode(t, humaRequest(t, e, http.MethodGet, occPath(id, ""), "", token, ""), http.StatusOK).Occurrences, models.DefaultOccurrencesCount)
	for _, n := range []int{1, models.MaxOccurrencesCount} {
		answer := occDecode(t, humaRequest(t, e, http.MethodGet, occPath(id, fmt.Sprintf("count=%d", n)), "", token, ""), http.StatusOK)
		assert.Len(t, answer.Occurrences, n)
	}
	for _, n := range []int{0, models.MaxOccurrencesCount + 1} {
		rec := humaRequest(t, e, http.MethodGet, occPath(id, fmt.Sprintf("count=%d", n)), "", token, "")
		assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "count=%d: %s", n, rec.Body.String())
	}
}

// TestTaskOccurrences_Access: a caller without read access gets exactly the 403 a task read gives,
// a missing task exactly the read's 404; the owner reads the same task.
func TestTaskOccurrences_Access(t *testing.T) {
	e, err := setupTestEnv()
	require.NoError(t, err)
	user1 := humaTokenFor(t, &testuser1)
	owner := humaTokenFor(t, &testuser13) // task #34 lives in project 20, private to user13

	occDecode(t, humaRequest(t, e, http.MethodGet, occPath(34, ""), "", owner, ""), http.StatusOK)

	for name, c := range map[string]struct {
		id     int64
		status int
	}{
		"no read access": {34, http.StatusForbidden},
		"missing task":   {99999, http.StatusNotFound},
	} {
		t.Run(name, func(t *testing.T) {
			read := humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/tasks/%d", c.id), "", user1, "")
			require.Equal(t, c.status, read.Code, read.Body.String())
			query := humaRequest(t, e, http.MethodGet, occPath(c.id, ""), "", user1, "")
			assert.Equal(t, c.status, query.Code, query.Body.String())
			assert.JSONEq(t, read.Body.String(), query.Body.String(), "the same body as tasks-read")
		})
	}
}

// TestTaskOccurrences_RoutePrecedence: the static preview path is not shadowed by /tasks/{task},
// and GET /tasks/occurrences, which no route serves, does not answer 200.
func TestTaskOccurrences_RoutePrecedence(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	answer := occDecode(t, occPreview(t, e, token, map[string]any{"repeats_rrule": "FREQ=DAILY", "due_date": occAnchor, "count": 2}), http.StatusOK)
	assert.Len(t, answer.Occurrences, 2, "the preview handler answered")

	rec := humaRequest(t, e, http.MethodGet, "/api/v2/tasks/occurrences", "", token, "")
	assert.NotEqual(t, http.StatusOK, rec.Code, rec.Body.String())
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, "it reaches tasks-read with a non-numeric id: %s", rec.Body.String())

	// Paired: the real query on a task id still resolves.
	occDecode(t, humaRequest(t, e, http.MethodGet, occPath(1, ""), "", token, ""), http.StatusOK)
}

// TestOccurrencePreview_Forms: each rule form previews; both, neither and a missing due_date are 422.
func TestOccurrencePreview_Forms(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	structured := occDecode(t, occPreview(t, e, token, map[string]any{"repeat": map[string]any{"freq": "daily", "interval": 2}, "due_date": occAnchor, "count": 3}), http.StatusOK)
	text := occDecode(t, occPreview(t, e, token, map[string]any{"repeats_rrule": "FREQ=DAILY;INTERVAL=2", "due_date": occAnchor, "count": 3}), http.StatusOK)
	assertSameInstants(t, []time.Time{
		time.Date(2027, 3, 3, 9, 0, 0, 0, time.UTC),
		time.Date(2027, 3, 5, 9, 0, 0, 0, time.UTC),
		time.Date(2027, 3, 7, 9, 0, 0, 0, time.UTC),
	}, structured.Occurrences)
	assertSameInstants(t, structured.Occurrences, text.Occurrences)

	defaulted := occDecode(t, occPreview(t, e, token, map[string]any{"repeats_rrule": "FREQ=DAILY", "due_date": occAnchor}), http.StatusOK)
	assert.Len(t, defaulted.Occurrences, models.DefaultOccurrencesCount)

	for name, body := range map[string]map[string]any{
		"both forms":       {"repeat": map[string]any{"freq": "daily"}, "repeats_rrule": "FREQ=DAILY", "due_date": occAnchor},
		"neither form":     {"due_date": occAnchor},
		"missing due_date": {"repeats_rrule": "FREQ=DAILY"},
		"count 0":          {"repeats_rrule": "FREQ=DAILY", "due_date": occAnchor, "count": 0},
		"count 101":        {"repeats_rrule": "FREQ=DAILY", "due_date": occAnchor, "count": models.MaxOccurrencesCount + 1},
	} {
		t.Run(name+" is 422", func(t *testing.T) {
			rec := occPreview(t, e, token, body)
			assert.Equal(t, http.StatusUnprocessableEntity, rec.Code, rec.Body.String())
		})
	}
	rec := occPreview(t, e, token, map[string]any{"repeat": map[string]any{"freq": "daily"}, "repeats_rrule": "FREQ=DAILY", "due_date": occAnchor})
	assert.Contains(t, rec.Body.String(), `"location":"body.repeat"`)
	assert.Contains(t, rec.Body.String(), `"location":"body.repeats_rrule"`)

	// An empty text is a sent form, and the validator refuses it (on a task write it clears).
	empty := occPreview(t, e, token, map[string]any{"repeats_rrule": "", "due_date": occAnchor})
	status, code, _ := occProblem(t, empty)
	assert.Equal(t, http.StatusBadRequest, status, empty.Body.String())
	assert.Equal(t, models.ErrCodeInvalidData, code)
}

// TestOccurrencePreview_RefusalParity: a rule task create refuses is refused by the
// preview with the same status, code and detail; each refusal is paired with an accepted rule on
// the same request shape, which both create and preview accept.
func TestOccurrencePreview_RefusalParity(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	textRule := func(text string) map[string]any { return map[string]any{"repeats_rrule": text} }
	structuredRule := func(freq string) map[string]any { return map[string]any{"repeat": map[string]any{"freq": freq}} }
	for name, c := range map[string]struct{ refused, accepted map[string]any }{
		"policy":          {textRule("FREQ=HOURLY;BYHOUR=9"), textRule("FREQ=DAILY;BYHOUR=9")},
		"bounds":          {textRule("FREQ=DAILY;COUNT=10001"), textRule("FREQ=DAILY;COUNT=10000")},
		"RFC 5545":        {textRule("FREQ=DAILY;COUNT=2;UNTIL=20271231T235959Z"), textRule("FREQ=DAILY;COUNT=2")},
		"floating UNTIL":  {textRule("FREQ=DAILY;UNTIL=20271231T235959"), textRule("FREQ=DAILY;UNTIL=20271231T235959Z")},
		"structured freq": {structuredRule("bogus"), structuredRule("daily")},
	} {
		t.Run(name, func(t *testing.T) {
			withAnchor := func(rule map[string]any, extra map[string]any) map[string]any {
				body := map[string]any{"due_date": occAnchor}
				for k, v := range rule {
					body[k] = v
				}
				for k, v := range extra {
					body[k] = v
				}
				return body
			}

			created := humaRequest(t, e, http.MethodPost, occCreatePath, marshal(t, withAnchor(c.refused, map[string]any{"title": "refused"})), token, "")
			previewed := occPreview(t, e, token, withAnchor(c.refused, nil))
			createStatus, createCode, createDetail := occProblem(t, created)
			previewStatus, previewCode, previewDetail := occProblem(t, previewed)
			assert.Equal(t, http.StatusBadRequest, createStatus, created.Body.String())
			assert.Equal(t, models.ErrCodeInvalidData, createCode, created.Body.String())
			assert.NotEmpty(t, createDetail)
			assert.Equal(t, createStatus, previewStatus, previewed.Body.String())
			assert.Equal(t, createCode, previewCode)
			assert.Equal(t, createDetail, previewDetail)

			ruleTextDo(t, e, http.MethodPost, occCreatePath, marshal(t, withAnchor(c.accepted, map[string]any{"title": "accepted"})), token, "", http.StatusCreated)
			answer := occDecode(t, occPreview(t, e, token, withAnchor(c.accepted, nil)), http.StatusOK)
			assert.NotEmpty(t, answer.Occurrences)
		})
	}
}

// TestOccurrencePreview_RuleTooLong pins the length bound: 501 characters of rule text are
// refused with 2002 before any evaluation, as on create; the same rule shape, shorter, previews.
func TestOccurrencePreview_RuleTooLong(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	const prefix = "FREQ=DAILY;BYMONTHDAY="
	long := prefix + strings.Repeat("1,", 300)
	long = long[:501]
	require.Len(t, long, 501)

	previewed := occPreview(t, e, token, map[string]any{"repeats_rrule": long, "due_date": occAnchor})
	status, code, detail := occProblem(t, previewed)
	assert.Equal(t, http.StatusBadRequest, status, previewed.Body.String())
	assert.Equal(t, models.ErrCodeInvalidData, code)
	assert.Contains(t, detail, "501 characters")

	created := humaRequest(t, e, http.MethodPost, occCreatePath, marshal(t, map[string]any{"title": "long", "repeats_rrule": long}), token, "")
	createStatus, createCode, createDetail := occProblem(t, created)
	assert.Equal(t, []any{createStatus, createCode, createDetail}, []any{status, code, detail})

	occDecode(t, occPreview(t, e, token, map[string]any{"repeats_rrule": prefix + "1", "due_date": occAnchor}), http.StatusOK)
}

// TestOccurrencePreview_TimeZones: an explicit zone changes the wall-clock result across a
// daylight-saving change; without one the caller's setting applies, then the server's; an
// unknown zone is 400 with code 2003.
func TestOccurrencePreview_TimeZones(t *testing.T) {
	e, token := ruleTextSetup(t, "")
	// 09:00 in New York the day before its 2027 daylight-saving change (March 14).
	body := func(zone string) map[string]any {
		b := map[string]any{"repeats_rrule": "FREQ=DAILY", "due_date": "2027-03-13T14:00:00Z", "count": 3}
		if zone != "" {
			b["timezone"] = zone
		}
		return b
	}

	newYork := occDecode(t, occPreview(t, e, token, body("America/New_York")), http.StatusOK)
	assertSameInstants(t, []time.Time{
		time.Date(2027, 3, 14, 13, 0, 0, 0, time.UTC),
		time.Date(2027, 3, 15, 13, 0, 0, 0, time.UTC),
		time.Date(2027, 3, 16, 13, 0, 0, 0, time.UTC),
	}, newYork.Occurrences)

	// user1 has no zone set: the server's applies.
	anchor := time.Date(2027, 3, 13, 14, 0, 0, 0, time.UTC).In(config.GetTimeZone())
	server := occDecode(t, occPreview(t, e, token, body("")), http.StatusOK)
	assertSameInstants(t, []time.Time{anchor.AddDate(0, 0, 1), anchor.AddDate(0, 0, 2), anchor.AddDate(0, 0, 3)}, server.Occurrences)
	assert.False(t, server.Occurrences[0].Equal(newYork.Occurrences[0]), "the explicit zone changed the result")

	rec := occPreview(t, e, token, body("Mars/Olympus_Mons"))
	status, code, _ := occProblem(t, rec)
	assert.Equal(t, http.StatusBadRequest, status, rec.Body.String())
	assert.Equal(t, models.ErrCodeInvalidTimezone, code)

	ruleTextDo(t, e, http.MethodPatch, "/api/v2/user/settings/general", `{"timezone":"America/New_York"}`, token, ruleTextMerge, http.StatusOK)
	caller := occDecode(t, occPreview(t, e, token, body("")), http.StatusOK)
	assertSameInstants(t, newYork.Occurrences, caller.Occurrences)
}

// TestAPIToken_TaskOccurrences: the routes derive the token permissions tasks.occurrences and
// tasks.occurrences_preview, each covering only its own route, and a real token holding one of
// them can call that route and no other.
func TestAPIToken_TaskOccurrences(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	routes := models.GetAPITokenRoutes()["tasks"]
	require.NotNil(t, routes)
	assert.Equal(t, &models.RouteDetail{Path: "/api/v2/tasks/:task/occurrences", Method: http.MethodGet}, routes["occurrences"])
	assert.Equal(t, &models.RouteDetail{Path: "/api/v2/tasks/occurrences/preview", Method: http.MethodPost}, routes["occurrences_preview"])

	var all []tokenRoute
	for _, r := range e.Router().Routes() {
		if r.Method != echo.RouteNotFound && strings.HasPrefix(r.Path, "/api/") {
			all = append(all, tokenRoute{Method: r.Method, Path: r.Path})
		}
	}
	for perm, want := range map[string]tokenRoute{
		"occurrences":         {http.MethodGet, "/api/v2/tasks/:task/occurrences"},
		"occurrences_preview": {http.MethodPost, "/api/v2/tasks/occurrences/preview"},
	} {
		scoped := &models.APIToken{APIPermissions: models.APIPermissions{"tasks": {perm}}}
		var got []tokenRoute
		for _, r := range all {
			if models.CanDoAPIRoute(e.NewContext(httptest.NewRequest(r.Method, r.Path, nil), httptest.NewRecorder()), scoped) {
				got = append(got, r)
			}
		}
		assert.Equal(t, []tokenRoute{want}, got, "tasks.%s", perm)
	}

	issue := func(perms string) string {
		t.Helper()
		rec := humaRequest(t, e, http.MethodPost, "/api/v2/tokens",
			`{"title":"occurrences","permissions":`+perms+`,"expires_at":"2099-01-01T00:00:00Z"}`, token, "")
		require.Equal(t, http.StatusCreated, rec.Code, rec.Body.String())
		var created struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &created))
		return created.Token
	}
	preview := marshal(t, map[string]any{"repeats_rrule": "FREQ=DAILY", "due_date": occAnchor})

	occurrencesOnly := issue(`{"tasks":["occurrences"]}`)
	occDecode(t, humaRequest(t, e, http.MethodGet, occPath(1, ""), "", occurrencesOnly, ""), http.StatusOK)
	assert.Equal(t, http.StatusUnauthorized, humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1", "", occurrencesOnly, "").Code, "tasks.occurrences does not grant a task read")
	assert.Equal(t, http.StatusUnauthorized, humaRequest(t, e, http.MethodPost, occPreviewPath, preview, occurrencesOnly, "").Code, "nor the preview")

	readOnly := issue(`{"tasks":["read_one"]}`)
	assert.Equal(t, http.StatusOK, humaRequest(t, e, http.MethodGet, "/api/v2/tasks/1", "", readOnly, "").Code)
	assert.Equal(t, http.StatusUnauthorized, humaRequest(t, e, http.MethodGet, occPath(1, ""), "", readOnly, "").Code, "a task read does not grant the query")

	previewOnly := issue(`{"tasks":["occurrences_preview"]}`)
	occDecode(t, humaRequest(t, e, http.MethodPost, occPreviewPath, preview, previewOnly, ""), http.StatusOK)
}

// Preview and query rows (the status matrix, the initial-due and forecast rows and the GET answer). Every
// test below pins the engine clock with
// models.PinEngineClockForTests (never in parallel) and names its zones: the caller's, the
// project owner's and the server's (config.GetTimeZone, GMT in the test config) are always three
// different zones, so an answer in the wrong one cannot pass.

const (
	occBerlin   = "Europe/Berlin"
	occNewYork  = "America/New_York"
	occCaller   = "Asia/Tokyo"          // the caller's zone where a row needs one besides the evaluation zone
	occOwner    = "America/Los_Angeles" // user6, owner of project 9 (shared read-only with user1)
	occStranger = "Pacific/Auckland"    // user13, owner of project 20 (no access for user1)
)

// occPreviewAnswer is the preview's answer (schema PreviewAnswer); raw keeps the JSON keys so a
// row can assert that an absent field is absent rather than empty.
type occPreviewAnswer struct {
	Occurrences   []time.Time `json:"occurrences"`
	Reason        string      `json:"reason"`
	DueDate       *time.Time  `json:"due_date"`
	DueDateSource string      `json:"due_date_source"`
	TimeZone      string      `json:"timezone"`
	EvaluatedAt   time.Time   `json:"evaluated_at"`
	raw           map[string]json.RawMessage
}

// occPreviewDecode asserts a 200 preview answer carrying the effective zone and an evaluated_at
// equal to the pinned clock, and decodes it.
func occPreviewDecode(t *testing.T, rec *httptest.ResponseRecorder, zone string, now time.Time) occPreviewAnswer {
	t.Helper()
	occDecode(t, rec, http.StatusOK)
	answer := occPreviewAnswer{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &answer), rec.Body.String())
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &answer.raw))
	assert.Equal(t, zone, answer.TimeZone, "timezone: %s", rec.Body.String())
	assert.True(t, answer.EvaluatedAt.Equal(now), "evaluated_at %s is the pinned clock %s", answer.EvaluatedAt, now)
	return answer
}

// assertNoDueDate asserts a preview answer that selected no date: due_date and due_date_source
// are absent, occurrences is [] and reason says why.
func (a occPreviewAnswer) assertNoDueDate(t *testing.T, reason string) {
	t.Helper()
	assert.NotContains(t, a.raw, "due_date")
	assert.NotContains(t, a.raw, "due_date_source")
	assert.Empty(t, a.Occurrences)
	assert.Equal(t, reason, a.Reason)
}

// assertDue asserts the preview's anchor and its source.
func (a occPreviewAnswer) assertDue(t *testing.T, want time.Time, source string) {
	t.Helper()
	require.NotNil(t, a.DueDate, "due_date is present")
	assert.True(t, a.DueDate.Equal(want), "due_date: got %s, want %s", a.DueDate, want)
	assert.Equal(t, source, a.DueDateSource)
}

// occPinClock pins the engine clock to now for the rest of the test.
func occPinClock(t *testing.T, now time.Time) {
	t.Helper()
	t.Cleanup(models.PinEngineClockForTests(now))
}

// occSetZone stores a user's zone through the v2 settings route, which answers 304 when the zone
// is already the stored one.
func occSetZone(t *testing.T, e *echo.Echo, token, zone string) {
	t.Helper()
	rec := humaRequest(t, e, http.MethodPatch, "/api/v2/user/settings/general", marshal(t, map[string]any{"timezone": zone}), token, ruleTextMerge)
	require.Contains(t, []int{http.StatusOK, http.StatusNotModified}, rec.Code, rec.Body.String())
}

// occRequireDistinctZones guards the fixture: the zones a test tells apart must really differ,
// the server's included.
func occRequireDistinctZones(t *testing.T, zones ...string) {
	t.Helper()
	all := append([]string{config.GetTimeZone().String()}, zones...)
	seen := map[string]bool{}
	for _, z := range all {
		require.False(t, seen[z], "zone %s is used twice in %v", z, all)
		seen[z] = true
	}
}

func occZone(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	require.NoError(t, err)
	return loc
}

// TestOccurrencePreview_InitialDueRows covers rows P-1 to P-7c and P-13 below over HTTP: the
// due date initial_due_time selects at the pinned now, the occurrences strictly after it, and the
// reason when no date can be selected. Every row sends an explicit timezone that differs from the
// caller's stored zone and from the server's, and asserts timezone, evaluated_at and
// due_date_source.
//
// Black-box boundary: the HTTP API through the v2 router, real database and auth; the engine
// clock is substituted (outside the boundary); everything else is real.
func TestOccurrencePreview_InitialDueRows(t *testing.T) {
	berlin, ny := occZone(t, occBerlin), occZone(t, occNewYork)
	b := func(d, h, m, sec int) time.Time { return time.Date(2026, 10, d, h, m, sec, 0, berlin) }
	n := func(d, h, m int) time.Time { return time.Date(2023, 3, d, h, m, 0, 0, ny) }

	for _, row := range []struct {
		name, rule, hhmm, zone string
		now                    time.Time
		fromCompletion         bool
		count                  int
		wantDue                time.Time // zero: no due date
		want                   []time.Time
		wantReason             string
	}{
		{name: "P-1 DAILY 09:00, now 10:00", rule: "FREQ=DAILY", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 2,
			wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0), b(31, 9, 0, 0)}},
		{name: "P-2 now exactly 09:00:00 (strictly after)", rule: "FREQ=DAILY", hhmm: "09:00", zone: occBerlin, now: b(28, 9, 0, 0), count: 1,
			wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0)}},
		{name: "P-3 NY gap seed DAILY 02:30", rule: "FREQ=DAILY", hhmm: "02:30", zone: occNewYork, now: n(12, 1, 0), count: 1,
			wantDue: n(13, 2, 30), want: []time.Time{n(14, 2, 30)}},
		{name: "P-4 NY gap seed MINUTELY;INTERVAL=15", rule: "FREQ=MINUTELY;INTERVAL=15", hhmm: "02:30", zone: occNewYork, now: n(12, 1, 0), count: 2,
			wantDue: n(12, 3, 0), want: []time.Time{n(12, 3, 15), n(12, 3, 30)}},
		{name: "P-5 COUNT=1, seed before now", rule: "FREQ=DAILY;COUNT=1", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 10,
			wantDue: b(29, 9, 0, 0), want: []time.Time{}, wantReason: "count_exhausted"},
		{name: "P-6 COUNT=3, count 5", rule: "FREQ=DAILY;COUNT=3", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 5,
			wantDue: b(29, 9, 0, 0), want: []time.Time{b(30, 9, 0, 0), b(31, 9, 0, 0)}, wantReason: "count_exhausted"},
		// 2026-10-29 09:00 CET is 08:00Z: UNTIL is inclusive.
		{name: "P-7a UNTIL = first candidate", rule: "FREQ=DAILY;UNTIL=20261029T080000Z", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 3,
			wantDue: b(29, 9, 0, 0), want: []time.Time{}, wantReason: "no_occurrence"},
		{name: "P-7b now <= UNTIL < first candidate", rule: "FREQ=DAILY;UNTIL=20261029T075959Z", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 3,
			wantReason: "no_occurrence"},
		{name: "P-7c UNTIL < now", rule: "FREQ=DAILY;UNTIL=20261027T000000Z", hhmm: "09:00", zone: occBerlin, now: b(28, 10, 0, 0), count: 3,
			wantReason: "until_passed"},
		{name: "P-13 HOURLY from completion, 18:00 at 10:00", rule: "FREQ=HOURLY;INTERVAL=1", hhmm: "18:00", zone: occBerlin, now: b(28, 10, 0, 0), fromCompletion: true, count: 3,
			wantDue: b(28, 18, 0, 0), want: []time.Time{b(28, 19, 0, 0), b(28, 20, 0, 0), b(28, 21, 0, 0)}},
	} {
		t.Run(row.name, func(t *testing.T) {
			e, token := ruleTextSetup(t, "")
			occRequireDistinctZones(t, row.zone, occCaller)
			occSetZone(t, e, token, occCaller)
			occPinClock(t, row.now)

			answer := occPreviewDecode(t, occPreview(t, e, token, map[string]any{
				"repeats_rrule": row.rule, "initial_due_time": row.hhmm, "timezone": row.zone,
				"repeats_from_current_date": row.fromCompletion, "count": row.count,
			}), row.zone, row.now)

			if row.wantDue.IsZero() {
				answer.assertNoDueDate(t, row.wantReason)
				return
			}
			answer.assertDue(t, row.wantDue, models.DueDateSourceGenerated)
			assert.True(t, answer.DueDate.After(row.now), "a generated due date is strictly after now")
			assert.Equal(t, row.wantReason, answer.Reason)
			assertSameInstants(t, row.want, answer.Occurrences)
			prev := *answer.DueDate
			for _, o := range answer.Occurrences {
				assert.True(t, o.After(prev), "%s is strictly after %s", o, prev)
				prev = o
			}
		})
	}
}

// TestOccurrencePreview_ExplicitDueMatchesQuery: an explicit due_date previews with
// due_date_source explicit, echoes the due date to the second, and lists exactly what the occurrences query
// lists for a saved task with that rule and due date (a rule that does not repeat from the
// current date, so the on-time forecast does not apply). The caller's zone (Berlin) decides both; the server's
// (GMT) differs.
//
// Black-box boundary: the HTTP API through the v2 router, real database and auth; the engine
// clock is substituted (outside the boundary); everything else is real.
func TestOccurrencePreview_ExplicitDueMatchesQuery(t *testing.T) {
	berlin := occZone(t, occBerlin)
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, berlin)
	due := time.Date(2026, 10, 29, 9, 0, 0, 0, berlin)
	const rule = "FREQ=WEEKLY;BYDAY=MO,TH"

	e, token := ruleTextSetup(t, "")
	occRequireDistinctZones(t, occBerlin)
	occSetZone(t, e, token, occBerlin)
	occPinClock(t, now)

	answer := occPreviewDecode(t, occPreview(t, e, token, map[string]any{
		"repeats_rrule": rule, "due_date": due.Add(400 * time.Millisecond).Format(time.RFC3339Nano), "count": 4,
	}), occBerlin, now)
	answer.assertDue(t, due, models.DueDateSourceExplicit)
	assert.Empty(t, answer.Reason)

	id := occCreate(t, e, token, map[string]any{"due_date": due.Format(time.RFC3339), "repeats_rrule": rule})
	query := occDecode(t, humaRequest(t, e, http.MethodGet, occPath(id, "count=4"), "", token, ""), http.StatusOK)
	require.Len(t, query.Occurrences, 4, "positive control")
	assertSameInstants(t, query.Occurrences, answer.Occurrences)
	assertSameInstants(t, []time.Time{
		time.Date(2026, 11, 2, 9, 0, 0, 0, berlin), time.Date(2026, 11, 5, 9, 0, 0, 0, berlin),
		time.Date(2026, 11, 9, 9, 0, 0, 0, berlin), time.Date(2026, 11, 12, 9, 0, 0, 0, berlin),
	}, answer.Occurrences)
}

// TestOccurrencePreview_StatusMatrix pins every status-matrix row HTTP can reach without zone
// context: the new 422 rows (P-8 anchor forms and a malformed initial_due_time, P-12 project_id
// shapes) and the existing statuses (rule form XOR and count bounds 422, a refused rule 400/2002,
// an unknown zone 400/2003). Each refusal is paired with the same request made valid, which is 200.
//
// Black-box boundary: the HTTP API through the v2 router, real database and auth; the engine
// clock is substituted (outside the boundary); everything else is real.
func TestOccurrencePreview_StatusMatrix(t *testing.T) {
	berlin := occZone(t, occBerlin)
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, berlin)
	due := time.Date(2026, 10, 29, 9, 0, 0, 0, berlin).Format(time.RFC3339)

	e, token := ruleTextSetup(t, "")
	occRequireDistinctZones(t, occBerlin, occCaller)
	occSetZone(t, e, token, occCaller)
	occPinClock(t, now)

	with := func(fields map[string]any) map[string]any {
		body := map[string]any{"repeats_rrule": "FREQ=DAILY", "timezone": occBerlin, "count": 2}
		for k, v := range fields {
			if v == nil {
				delete(body, k)
				continue
			}
			body[k] = v
		}
		return body
	}

	// Positive controls: each anchor form alone, and project_id 1 alone, preview.
	occPreviewDecode(t, occPreview(t, e, token, with(map[string]any{"due_date": due})), occBerlin, now).
		assertDue(t, time.Date(2026, 10, 29, 9, 0, 0, 0, berlin), models.DueDateSourceExplicit)
	occPreviewDecode(t, occPreview(t, e, token, with(map[string]any{"initial_due_time": "09:00"})), occBerlin, now).
		assertDue(t, time.Date(2026, 10, 29, 9, 0, 0, 0, berlin), models.DueDateSourceGenerated)
	occPreviewDecode(t, occPreview(t, e, token, with(map[string]any{"initial_due_time": "23:59", "project_id": 1})), occBerlin, now)
	occPreviewDecode(t, occPreview(t, e, token, with(map[string]any{"initial_due_time": "00:00", "task_id": 1})), occBerlin, now)

	for name, c := range map[string]struct {
		body   map[string]any
		status int
		code   int // Vikunja code; 0 for a schema/shape 422
	}{
		"P-8 due_date and initial_due_time":           {with(map[string]any{"due_date": due, "initial_due_time": "09:00"}), http.StatusUnprocessableEntity, 0},
		"P-8 neither due_date nor initial_due_time":   {with(nil), http.StatusUnprocessableEntity, 0},
		"P-8 malformed initial_due_time 24:00":        {with(map[string]any{"initial_due_time": "24:00"}), http.StatusUnprocessableEntity, 0},
		"P-8 malformed initial_due_time 9:00":         {with(map[string]any{"initial_due_time": "9:00"}), http.StatusUnprocessableEntity, 0},
		"P-8 malformed initial_due_time 09:60":        {with(map[string]any{"initial_due_time": "09:60"}), http.StatusUnprocessableEntity, 0},
		"P-8 malformed initial_due_time 09:00:00":     {with(map[string]any{"initial_due_time": "09:00:00"}), http.StatusUnprocessableEntity, 0},
		"P-12 negative project_id":                    {with(map[string]any{"due_date": due, "project_id": -1}), http.StatusUnprocessableEntity, 0},
		"P-12 task_id and project_id":                 {with(map[string]any{"due_date": due, "task_id": 1, "project_id": 1}), http.StatusUnprocessableEntity, 0},
		"rule forms: both":                            {with(map[string]any{"due_date": due, "repeat": map[string]any{"freq": "daily"}}), http.StatusUnprocessableEntity, 0},
		"rule forms: neither":                         {with(map[string]any{"due_date": due, "repeats_rrule": nil}), http.StatusUnprocessableEntity, 0},
		"count 0":                                     {with(map[string]any{"due_date": due, "count": 0}), http.StatusUnprocessableEntity, 0},
		"count 101":                                   {with(map[string]any{"due_date": due, "count": models.MaxOccurrencesCount + 1}), http.StatusUnprocessableEntity, 0},
		"rule a save refuses (explicit due)":          {with(map[string]any{"due_date": due, "repeats_rrule": "FREQ=HOURLY;BYHOUR=9"}), http.StatusBadRequest, models.ErrCodeInvalidData},
		"rule a save refuses (initial_due_time)":      {with(map[string]any{"initial_due_time": "09:00", "repeats_rrule": "FREQ=HOURLY;BYHOUR=9"}), http.StatusBadRequest, models.ErrCodeInvalidData},
		"unknown timezone (explicit due)":             {with(map[string]any{"due_date": due, "timezone": "Mars/Olympus_Mons"}), http.StatusBadRequest, models.ErrCodeInvalidTimezone},
		"unknown timezone (initial_due_time)":         {with(map[string]any{"initial_due_time": "09:00", "timezone": "Mars/Olympus_Mons"}), http.StatusBadRequest, models.ErrCodeInvalidTimezone},
		"unknown timezone with project_id as context": {with(map[string]any{"due_date": due, "project_id": 1, "timezone": "Mars/Olympus_Mons"}), http.StatusBadRequest, models.ErrCodeInvalidTimezone},
	} {
		t.Run(name, func(t *testing.T) {
			rec := occPreview(t, e, token, c.body)
			require.Equal(t, c.status, rec.Code, rec.Body.String())
			if c.code != 0 {
				_, code, _ := occProblem(t, rec)
				assert.Equal(t, c.code, code, rec.Body.String())
			}
			assert.NotContains(t, rec.Body.String(), `"occurrences"`)
		})
	}

	// P-8's conflict names both fields, as the rule-form XOR does.
	rec := occPreview(t, e, token, with(map[string]any{"due_date": due, "initial_due_time": "09:00"}))
	assert.Contains(t, rec.Body.String(), `"location":"body.initial_due_time"`)
	assert.Contains(t, rec.Body.String(), `"location":"body.due_date"`)
}

// occForecastRow is one on-time forecast row run on both list surfaces.
type occForecastRow struct {
	name, zone, rule string
	due, now         time.Time
	generated        bool // F-7: the preview selects the due date from initial_due_time
	count            int
	want             []time.Time
	wantReason       string
}

// TestTaskOccurrences_ForecastBothSurfaces runs the forecast rows F-7 to F-14 below (both F-13 UNTIL
// variants) on the preview and on GET /tasks/{task}/occurrences for a saved task with the same
// rule, repeat-from-completion flag and due date. The caller's stored zone decides both surfaces
// (Berlin, New York for F-14); the server's (GMT) differs. Both give the same list and reason;
// F-13's reason is no_occurrence, as observed. The GET answer keeps exactly
// occurrences and reason.
//
// Black-box boundary: the HTTP API through the v2 router, real database and auth (the saved task
// is created through POST /projects/1/tasks); the engine clock is substituted (outside the
// boundary); everything else is real.
func TestTaskOccurrences_ForecastBothSurfaces(t *testing.T) {
	berlin, ny := occZone(t, occBerlin), occZone(t, occNewYork)
	b := func(h, m, sec int) time.Time { return time.Date(2026, 10, 28, h, m, sec, 0, berlin) }
	hours := func(from time.Time, n int) []time.Time {
		out := make([]time.Time, n)
		for i := range out {
			out[i] = from.Add(time.Duration(i) * time.Hour)
		}
		return out
	}
	const hourly = "FREQ=HOURLY;INTERVAL=1"

	rows := []occForecastRow{
		{name: "F-7 future generated", zone: occBerlin, rule: hourly, due: b(18, 0, 0), now: b(10, 0, 0), generated: true, count: 3, want: hours(b(19, 0, 0), 3)},
		{name: "F-8 future explicit", zone: occBerlin, rule: hourly, due: b(18, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(19, 0, 0), 3)},
		{name: "F-9 overdue explicit", zone: occBerlin, rule: hourly, due: b(8, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(11, 0, 0), 3)},
		{name: "F-10 due equal to now", zone: occBerlin, rule: hourly, due: b(10, 0, 0), now: b(10, 0, 0), count: 3, want: hours(b(11, 0, 0), 3)},
		{name: "F-11 COUNT=1", zone: occBerlin, rule: hourly + ";COUNT=1", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: []time.Time{}, wantReason: "count_exhausted"},
		{name: "F-12 COUNT=3, count 10", zone: occBerlin, rule: hourly + ";COUNT=3", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 2), wantReason: "count_exhausted"},
		// 20:00 CET is 19:00Z.
		{name: "F-13 UNTIL = 20:00", zone: occBerlin, rule: hourly + ";UNTIL=20261028T190000Z", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 2), wantReason: "no_occurrence"},
		{name: "F-13 UNTIL = 19:59:59", zone: occBerlin, rule: hourly + ";UNTIL=20261028T185959Z", due: b(18, 0, 0), now: b(10, 0, 0), count: 10, want: hours(b(19, 0, 0), 1), wantReason: "no_occurrence"},
		{name: "F-14 NY across spring forward", zone: occNewYork, rule: hourly,
			due: time.Date(2026, 3, 8, 1, 30, 0, 0, ny), now: time.Date(2026, 3, 7, 23, 0, 0, 0, ny), count: 3,
			want: []time.Time{time.Date(2026, 3, 8, 3, 30, 0, 0, ny), time.Date(2026, 3, 8, 4, 30, 0, 0, ny), time.Date(2026, 3, 8, 5, 30, 0, 0, ny)}},
	}
	for _, row := range rows {
		t.Run(row.name, func(t *testing.T) {
			e, token := ruleTextSetup(t, "")
			occRequireDistinctZones(t, row.zone)
			occSetZone(t, e, token, row.zone)
			occPinClock(t, row.now)

			body := map[string]any{"repeats_rrule": row.rule, "repeats_from_current_date": true, "count": row.count}
			source := models.DueDateSourceExplicit
			if row.generated {
				body["initial_due_time"] = row.due.Format("15:04")
				source = models.DueDateSourceGenerated
			} else {
				body["due_date"] = row.due.Format(time.RFC3339)
			}
			preview := occPreviewDecode(t, occPreview(t, e, token, body), row.zone, row.now)
			preview.assertDue(t, row.due, source)

			id := occCreate(t, e, token, map[string]any{
				"due_date": row.due.Format(time.RFC3339), "repeats_rrule": row.rule, "repeats_from_current_date": true,
			})
			rec := humaRequest(t, e, http.MethodGet, occPath(id, fmt.Sprintf("count=%d", row.count)), "", token, "")
			query := occDecode(t, rec, http.StatusOK)
			raw := map[string]json.RawMessage{}
			require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &raw))
			delete(raw, "$schema") // huma's schema link, not a field
			for key := range raw {
				assert.Contains(t, []string{"occurrences", "reason"}, key, "GET carries only occurrences and reason: %s", rec.Body.String())
			}

			for surface, got := range map[string]occAnswer{
				"preview": {Occurrences: preview.Occurrences, Reason: preview.Reason},
				"query":   query,
			} {
				assertSameInstants(t, row.want, got.Occurrences)
				assert.Equal(t, row.wantReason, got.Reason, surface)
				prev := row.due
				for _, o := range got.Occurrences {
					assert.True(t, o.After(prev), "%s: %s is strictly after %s", surface, o, prev)
					if row.zone == occNewYork {
						assert.NotEqual(t, 2, o.In(ny).Hour(), "%s: no 02:xx across spring forward", surface)
					}
					prev = o
				}
			}
			if row.zone == occNewYork {
				assert.Equal(t, time.Hour, preview.Occurrences[0].Sub(row.due), "03:30 EDT is one elapsed hour after 01:30 EST")
			}
		})
	}
}

// TestOccurrencePreview_ZoneContextAccess: task_id and project_id
// are zone context that needs read access, refused exactly as the task or project read is, with
// no zone in the refusal; a readable one evaluates in that project's zone. Zones: user1 (the
// caller) none at first, then Asia/Tokyo; user6 (owner of project 9, shared read-only with user1)
// America/Los_Angeles; user13 (owner of project 20, which user1 cannot read) Pacific/Auckland; the
// server GMT. With the caller zoneless, task #18 (project 9) answers in the owner's zone and a
// preview without context in the server's, so the owner zone can only come from the context.
//
// Black-box boundary: the HTTP API through the v2 router, real database, auth and link-share
// JWT; the engine clock is substituted (outside the boundary); everything else is real.
func TestOccurrencePreview_ZoneContextAccess(t *testing.T) {
	now := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	server := config.GetTimeZone().String()

	e, err := setupTestEnv()
	require.NoError(t, err)
	occRequireDistinctZones(t, occCaller, occOwner, occStranger)
	user1 := humaTokenFor(t, &testuser1)
	occSetZone(t, e, humaTokenFor(t, &testuser6), occOwner)
	occSetZone(t, e, humaTokenFor(t, &testuser13), occStranger)
	occPinClock(t, now)

	preview := func(token string, context map[string]any) *httptest.ResponseRecorder {
		body := map[string]any{"repeats_rrule": "FREQ=DAILY", "due_date": "2026-10-29T09:00:00Z", "count": 1}
		for k, v := range context {
			body[k] = v
		}
		return occPreview(t, e, token, body)
	}

	t.Run("readable task_id and project_id use that project's zone", func(t *testing.T) {
		occPreviewDecode(t, preview(user1, nil), server, now)
		occPreviewDecode(t, preview(user1, map[string]any{"task_id": 18}), occOwner, now)
		occPreviewDecode(t, preview(user1, map[string]any{"project_id": 9}), occOwner, now)
		// An explicit timezone still wins over the context.
		occPreviewDecode(t, preview(user1, map[string]any{"task_id": 18, "timezone": occBerlin}), occBerlin, now)
	})

	t.Run("the caller's stored zone comes before the owner's, as on a write", func(t *testing.T) {
		occSetZone(t, e, user1, occCaller)
		occPreviewDecode(t, preview(user1, map[string]any{"task_id": 18}), occCaller, now)
		occPreviewDecode(t, preview(user1, map[string]any{"project_id": 9}), occCaller, now)
		// The caller's own project: the owner is the caller.
		occPreviewDecode(t, preview(user1, map[string]any{"project_id": 1}), occCaller, now)
	})

	for name, c := range map[string]struct {
		context  map[string]any
		readPath string
		status   int
	}{
		"task_id unreadable":         {map[string]any{"task_id": 34}, "/api/v2/tasks/34", http.StatusForbidden},
		"task_id nonexistent":        {map[string]any{"task_id": 99999}, "/api/v2/tasks/99999", http.StatusNotFound},
		"task_id negative":           {map[string]any{"task_id": -5}, "/api/v2/tasks/-5", http.StatusNotFound},
		"project_id of another user": {map[string]any{"project_id": 20}, "/api/v2/projects/20", http.StatusForbidden},
		"project_id nonexistent":     {map[string]any{"project_id": 99999}, "/api/v2/projects/99999", http.StatusNotFound},
	} {
		t.Run(name+" is refused as the read is", func(t *testing.T) {
			read := humaRequest(t, e, http.MethodGet, c.readPath, "", user1, "")
			require.Equal(t, c.status, read.Code, read.Body.String())
			rec := preview(user1, c.context)
			require.Equal(t, c.status, rec.Code, rec.Body.String())
			assert.JSONEq(t, read.Body.String(), rec.Body.String(), "the same body as the read")
			// No access widening: nothing about any zone in the refusal.
			for _, leak := range []string{`"timezone"`, `"occurrences"`, occStranger, occOwner, server} {
				assert.NotContains(t, rec.Body.String(), leak)
			}
		})
	}

	t.Run("link share", func(t *testing.T) {
		// Observed and pinned: the v2 auth path accepts a read link
		// share's JWT for the preview. A link share has no zone of its own, so without context the
		// server's applies; with its project (or a task in it) as context, the project owner's
		// (resolveTaskZone falls through to the owner). Access is still checked: a project the
		// share does not cover is refused like any read.
		share := &models.LinkSharing{ID: 1, Hash: "test", ProjectID: 1, Permission: models.PermissionRead}
		shareToken, err := auth.NewLinkShareJWTAuthtoken(share)
		require.NoError(t, err)
		occSetZone(t, e, user1, occCaller) // user1 owns project 1

		occPreviewDecode(t, preview(shareToken, nil), server, now)
		occPreviewDecode(t, preview(shareToken, map[string]any{"project_id": 1}), occCaller, now)
		occPreviewDecode(t, preview(shareToken, map[string]any{"task_id": 1}), occCaller, now)
		refused := preview(shareToken, map[string]any{"project_id": 9})
		assert.Equal(t, http.StatusForbidden, refused.Code, refused.Body.String())
		assert.NotContains(t, refused.Body.String(), occOwner)
	})
}

// TestTaskOccurrences_NonPositiveIDMatchesRead is a regression test: GET
// /tasks/0/occurrences, and a negative id, answer with exactly the status and body GET /tasks/0
// (or the negative id) gives, 404.
//
// Black-box boundary: the HTTP API through the v2 router, real database and auth; the engine
// clock is substituted (outside the boundary); everything else is real.
func TestTaskOccurrences_NonPositiveIDMatchesRead(t *testing.T) {
	e, token := ruleTextSetup(t, "")
	occRequireDistinctZones(t, occCaller)
	occSetZone(t, e, token, occCaller)
	occPinClock(t, time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC))

	for _, id := range []int64{0, -1} {
		t.Run(fmt.Sprintf("id %d", id), func(t *testing.T) {
			read := humaRequest(t, e, http.MethodGet, fmt.Sprintf("/api/v2/tasks/%d", id), "", token, "")
			require.Equal(t, http.StatusNotFound, read.Code, read.Body.String())
			query := humaRequest(t, e, http.MethodGet, occPath(id, ""), "", token, "")
			assert.Equal(t, read.Code, query.Code, query.Body.String())
			assert.JSONEq(t, read.Body.String(), query.Body.String(), "the same body as tasks-read")
		})
	}
}
