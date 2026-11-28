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
	"testing"

	"code.vikunja.io/api/pkg/models"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The v2 rule text (repeats_rrule) through real HTTP requests: create, read, merge-patch,
// JSON Patch (the frontend's shape) and PUT, against the write-precedence table. Every
// subtest reloads the fixtures; the rule-on-a-fixture cases work on task #1 (user1, project 1),
// whose rule is seeded directly, as CalDAV or the RRULE migration would store it.
const (
	ruleTextTaskPath  = "/api/v2/tasks/1"
	ruleTextCreate    = "/api/v2/projects/1/tasks"
	ruleTextMerge     = "application/merge-patch+json"
	ruleTextJSONPatch = "application/json-patch+json"
	ruleTextDaily     = "FREQ=DAILY;INTERVAL=1"
	ruleTextEaster    = "FREQ=YEARLY;BYEASTER=0"
)

func ruleTextSetup(t *testing.T, rule string) (*echo.Echo, string) {
	t.Helper()
	e, err := setupTestEnv()
	require.NoError(t, err)
	if rule != "" {
		seedStoredRule(t, 1, rule)
	}
	return e, humaTokenFor(t, &testuser1)
}

func ruleTextBody(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	body := map[string]any{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body), rec.Body.String())
	return body
}

// ruleTextRead GETs a task on v2 and returns its body.
func ruleTextRead(t *testing.T, e *echo.Echo, token, path string) map[string]any {
	t.Helper()
	rec := humaRequest(t, e, http.MethodGet, path, "", token, "")
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	return ruleTextBody(t, rec)
}

func ruleTextDo(t *testing.T, e *echo.Echo, method, path, body, token, contentType string, want int) map[string]any {
	t.Helper()
	rec := humaRequest(t, e, method, path, body, token, contentType)
	require.Equal(t, want, rec.Code, rec.Body.String())
	return ruleTextBody(t, rec)
}

// ruleTextRefused asserts a 400 with domain code 2002 and returns the error detail.
func ruleTextRefused(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	require.Equal(t, http.StatusBadRequest, rec.Code, rec.Body.String())
	body := ruleTextBody(t, rec)
	assert.InDelta(t, float64(models.ErrCodeInvalidData), body["code"], 0, rec.Body.String())
	return rec.Body.String()
}

func ruleTextRepeat(t *testing.T, body map[string]any) map[string]any {
	t.Helper()
	repeat, ok := body["repeat"].(map[string]any)
	require.True(t, ok, "repeat in %v", body)
	return repeat
}

// TestTaskRuleText_EasterStory: an agent saves "every Easter Monday", reads the same
// rule back, changes it and clears it, all through v2 with no CalDAV client.
func TestTaskRuleText_EasterStory(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	created := ruleTextDo(t, e, http.MethodPost, ruleTextCreate,
		`{"title":"Easter Monday","due_date":"2027-03-29T09:00:00Z","repeats_rrule":"FREQ=YEARLY;BYEASTER=1"}`,
		token, "", http.StatusCreated)
	text, ok := created["repeats_rrule"].(string)
	require.True(t, ok, "the create response carries repeats_rrule: %v", created)
	assert.Contains(t, text, "FREQ=YEARLY")
	assert.Contains(t, text, "BYEASTER=1")
	assert.NotContains(t, created, "max_permission", "the write response adds only repeats_rrule")
	id := int64(created["id"].(float64))
	path := fmt.Sprintf("/api/v2/tasks/%d", id)
	assert.Equal(t, text, storedRule(t, id), "the response shows exactly the stored rule")

	read := ruleTextRead(t, e, token, path)
	assert.Equal(t, text, read["repeats_rrule"])
	assert.Equal(t, true, ruleTextRepeat(t, read)["read_only"], "BYEASTER has no structured form")
	byIndex := ruleTextRead(t, e, token, fmt.Sprintf("/api/v2/projects/1/tasks/by-index/%d", int64(created["index"].(float64))))
	assert.Equal(t, text, byIndex["repeats_rrule"], "the by-index read carries it too")

	changed := ruleTextDo(t, e, http.MethodPatch, path, `{"repeats_rrule":"FREQ=YEARLY;BYEASTER=0"}`, token, ruleTextMerge, http.StatusOK)
	stored := storedRule(t, id)
	assert.Contains(t, stored, "BYEASTER=0")
	assert.NotContains(t, stored, "BYEASTER=1")
	assert.Equal(t, stored, changed["repeats_rrule"], "the update response shows the new stored rule")
	assert.NotContains(t, changed, "max_permission", "the write response adds only repeats_rrule")
	assert.Equal(t, stored, ruleTextRead(t, e, token, path)["repeats_rrule"])

	cleared := ruleTextDo(t, e, http.MethodPatch, path, `{"repeats_rrule":""}`, token, ruleTextMerge, http.StatusOK)
	assert.Empty(t, storedRule(t, id))
	assert.NotContains(t, cleared, "repeats_rrule")
	read = ruleTextRead(t, e, token, path)
	assert.NotContains(t, read, "repeats_rrule", "absent once the task no longer repeats")
	assert.Nil(t, read["repeat"])
}

// TestTaskRuleText_FrontendJSONPatchClear: the frontend clears with a JSON Patch add /repeat null.
// AutoPatch applies it to the read body, so the PUT carries the echoed repeats_rrule and a nil
// repeat; the echo hands the decision to repeat, which clears.
func TestTaskRuleText_FrontendJSONPatchClear(t *testing.T) {
	for name, rule := range map[string]string{"representable": ruleTextDaily, "read-only": ruleTextEaster} {
		t.Run(name+" rule clears", func(t *testing.T) {
			e, token := ruleTextSetup(t, rule)
			require.Equal(t, rule, ruleTextRead(t, e, token, ruleTextTaskPath)["repeats_rrule"], "the GET body the patch applies to echoes the rule")

			body := ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, `[{"op":"add","path":"/repeat","value":null}]`, token, ruleTextJSONPatch, http.StatusOK)
			assert.Empty(t, storedRule(t, 1))
			assert.NotContains(t, body, "repeats_rrule")
		})
	}
}

// TestTaskRuleText_TitleOnlyEditKeepsRule: a title edit echoes the rule text and the structured
// projection, and the rule stays byte-identical, a read-only one included.
func TestTaskRuleText_TitleOnlyEditKeepsRule(t *testing.T) {
	edits := map[string]struct{ body, contentType string }{
		"merge-patch title": {`{"title":"renamed"}`, ruleTextMerge},
		"JSON Patch title":  {`[{"op":"replace","path":"/title","value":"renamed"}]`, ruleTextJSONPatch},
	}
	for ruleName, rule := range map[string]string{"representable": ruleTextDaily, "read-only": ruleTextEaster} {
		for editName, edit := range edits {
			t.Run(editName+" keeps a "+ruleName+" rule", func(t *testing.T) {
				e, token := ruleTextSetup(t, rule)
				body := ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, edit.body, token, edit.contentType, http.StatusOK)
				assert.Equal(t, "renamed", body["title"])
				assert.Equal(t, rule, storedRule(t, 1), "byte-identical")
				assert.Equal(t, rule, body["repeats_rrule"])
			})
		}
	}
}

// TestTaskRuleText_StructuredEditWithEchoedText: a structured edit through PATCH carries the
// echoed text; the echo hands the decision to repeat, so the structured change is stored.
func TestTaskRuleText_StructuredEditWithEchoedText(t *testing.T) {
	e, token := ruleTextSetup(t, ruleTextDaily)

	body := ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeat":{"freq":"weekly","by_day":["tu"]}}`, token, ruleTextMerge, http.StatusOK)
	stored := storedRule(t, 1)
	assert.Contains(t, stored, "FREQ=WEEKLY")
	assert.Contains(t, stored, "BYDAY=TU")
	assert.Equal(t, stored, body["repeats_rrule"])
}

// TestTaskRuleText_PUT: PUT replaces every field, repeat included. Omitting repeat clears the rule
// even though the body echoes repeats_rrule, exactly as before the field existed; a changed
// repeats_rrule is the change; an unchanged body keeps the rule; null means absent.
func TestTaskRuleText_PUT(t *testing.T) {
	readBody := func(t *testing.T, e *echo.Echo, token string) map[string]any {
		t.Helper()
		body := ruleTextRead(t, e, token, ruleTextTaskPath)
		require.Equal(t, ruleTextDaily, body["repeats_rrule"])
		return body
	}
	put := func(t *testing.T, e *echo.Echo, token string, body map[string]any) map[string]any {
		t.Helper()
		raw, err := json.Marshal(body)
		require.NoError(t, err)
		return ruleTextDo(t, e, http.MethodPut, ruleTextTaskPath, string(raw), token, "", http.StatusOK)
	}

	t.Run("full read body unchanged keeps the rule", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		resp := put(t, e, token, readBody(t, e, token))
		assert.Equal(t, ruleTextDaily, storedRule(t, 1))
		assert.Equal(t, ruleTextDaily, resp["repeats_rrule"])
	})

	t.Run("full read body without repeat clears, as PUT replaces repeat (echoed text is ignored)", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		body := readBody(t, e, token)
		delete(body, "repeat")
		resp := put(t, e, token, body)
		assert.Empty(t, storedRule(t, 1))
		assert.NotContains(t, resp, "repeats_rrule")
	})

	t.Run("full read body without repeat but with changed text stores the text", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		body := readBody(t, e, token)
		delete(body, "repeat")
		body["repeats_rrule"] = "FREQ=YEARLY;BYEASTER=1"
		resp := put(t, e, token, body)
		assert.Contains(t, storedRule(t, 1), "BYEASTER=1")
		assert.Equal(t, storedRule(t, 1), resp["repeats_rrule"])
	})

	t.Run("repeats_rrule null is treated as absent, not as a clear", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		body := readBody(t, e, token)
		body["repeats_rrule"] = nil
		put(t, e, token, body)
		assert.Equal(t, ruleTextDaily, storedRule(t, 1), "null with an echoed repeat keeps; an empty string would have cleared")
	})

	t.Run("repeats_rrule empty string with an echoed repeat clears", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		body := readBody(t, e, token)
		body["repeats_rrule"] = ""
		put(t, e, token, body)
		assert.Empty(t, storedRule(t, 1))
	})
}

// TestTaskRuleText_Disagreement: changing both forms to different rules is refused; changing both
// to the same rule is accepted.
func TestTaskRuleText_Disagreement(t *testing.T) {
	t.Run("both changed and disagreeing is refused", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		rec := humaRequest(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeat":{"freq":"weekly"},"repeats_rrule":"FREQ=MONTHLY"}`, token, ruleTextMerge)
		detail := ruleTextRefused(t, rec)
		assert.Contains(t, detail, "(repeat)")
		assert.Contains(t, detail, "(repeats_rrule)")
		assert.Equal(t, ruleTextDaily, storedRule(t, 1), "a refused write leaves the rule alone")
	})

	t.Run("clearing the text while changing repeat is refused", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		rec := humaRequest(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeat":{"freq":"weekly"},"repeats_rrule":""}`, token, ruleTextMerge)
		detail := ruleTextRefused(t, rec)
		assert.Contains(t, detail, "repeats_rrule")
		assert.Equal(t, ruleTextDaily, storedRule(t, 1))
	})

	t.Run("both changed and agreeing is stored", func(t *testing.T) {
		e, token := ruleTextSetup(t, ruleTextDaily)
		ruleTextDo(t, e, http.MethodPut, ruleTextTaskPath, `{"title":"task #1","repeat":{"freq":"monthly","interval":1},"repeats_rrule":"FREQ=MONTHLY"}`, token, "", http.StatusOK)
		assert.Contains(t, storedRule(t, 1), "FREQ=MONTHLY")
	})
}

// TestTaskRuleText_Refusals: rule text RFC 5545 or the creation policy forbids is refused with
// 2002 and the stored rule stays; an accepted text on the same task is stored.
func TestTaskRuleText_Refusals(t *testing.T) {
	cases := map[string]struct{ text, names string }{
		"COUNT with UNTIL": {"FREQ=DAILY;COUNT=2;UNTIL=20271231T235959Z", "UNTIL"},
		"floating UNTIL":   {"FREQ=DAILY;UNTIL=20271231T235959", "UNTIL"},
		"DTSTART part":     {"DTSTART=20270101T090000Z;FREQ=DAILY", "DTSTART"},
		"policy rejected":  {"FREQ=HOURLY;BYHOUR=9", "RRULE"},
	}
	for name, c := range cases {
		t.Run(name+" is refused on update", func(t *testing.T) {
			e, token := ruleTextSetup(t, ruleTextDaily)
			rec := humaRequest(t, e, http.MethodPatch, ruleTextTaskPath, marshal(t, map[string]string{"repeats_rrule": c.text}), token, ruleTextMerge)
			assert.Contains(t, ruleTextRefused(t, rec), c.names)
			assert.Equal(t, ruleTextDaily, storedRule(t, 1), "a refused write leaves the rule alone")

			// Paired acceptance on the same task: a valid text replaces the rule.
			ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeats_rrule":"FREQ=DAILY;UNTIL=20271231T235959Z"}`, token, ruleTextMerge, http.StatusOK)
			assert.Contains(t, storedRule(t, 1), "UNTIL=20271231T235959Z")
		})
		t.Run(name+" is refused on create", func(t *testing.T) {
			e, token := ruleTextSetup(t, "")
			rec := humaRequest(t, e, http.MethodPost, ruleTextCreate, marshal(t, map[string]string{"title": "refused", "repeats_rrule": c.text}), token, "")
			assert.Contains(t, ruleTextRefused(t, rec), c.names)
		})
	}
}

// TestTaskRuleText_StaleCountKept is the stale-count rule applied to text, documented in the field's doc:
// a higher COUNT on otherwise unchanged text counts as an echo, so COUNT cannot be raised through
// repeats_rrule alone; a lower COUNT is a change and is stored.
func TestTaskRuleText_StaleCountKept(t *testing.T) {
	const stored = "FREQ=DAILY;INTERVAL=1;COUNT=3"

	t.Run("a higher COUNT is a stale echo and keeps COUNT=3", func(t *testing.T) {
		e, token := ruleTextSetup(t, stored)
		ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeats_rrule":"FREQ=DAILY;COUNT=5"}`, token, ruleTextMerge, http.StatusOK)
		assert.Equal(t, stored, storedRule(t, 1))
	})

	t.Run("a lower COUNT is a change and is stored", func(t *testing.T) {
		e, token := ruleTextSetup(t, stored)
		ruleTextDo(t, e, http.MethodPatch, ruleTextTaskPath, `{"repeats_rrule":"FREQ=DAILY;COUNT=2"}`, token, ruleTextMerge, http.StatusOK)
		assert.Contains(t, storedRule(t, 1), "COUNT=2")
	})
}

// TestTaskRuleText_RepresentableTextShowsStructured: rule text the structured form can
// represent also reads back as an editable repeat; BYEASTER, on the same request shape, does not.
func TestTaskRuleText_RepresentableTextShowsStructured(t *testing.T) {
	e, token := ruleTextSetup(t, "")
	create := func(text string) map[string]any {
		created := ruleTextDo(t, e, http.MethodPost, ruleTextCreate,
			marshal(t, map[string]string{"title": "rule", "due_date": "2027-03-29T09:00:00Z", "repeats_rrule": text}), token, "", http.StatusCreated)
		return ruleTextRead(t, e, token, fmt.Sprintf("/api/v2/tasks/%d", int64(created["id"].(float64))))
	}

	weekly := create("FREQ=WEEKLY;BYDAY=MO")
	repeat := ruleTextRepeat(t, weekly)
	assert.Equal(t, "weekly", repeat["freq"])
	assert.Equal(t, []any{"mo"}, repeat["by_day"])
	assert.NotEqual(t, true, repeat["read_only"], "a representable rule is editable")
	text, ok := weekly["repeats_rrule"].(string)
	require.True(t, ok)
	assert.Contains(t, text, "FREQ=WEEKLY")
	assert.Contains(t, text, "BYDAY=MO")

	easter := create("FREQ=YEARLY;BYEASTER=1")
	assert.Equal(t, true, ruleTextRepeat(t, easter)["read_only"])
	assert.Contains(t, easter["repeats_rrule"], "BYEASTER=1")
}

// TestTaskRuleText_ClientWithoutTheField: a v2 client that never sends repeats_rrule writes the
// rule as before and sees it as one added field; v1 never shows it.
func TestTaskRuleText_ClientWithoutTheField(t *testing.T) {
	e, token := ruleTextSetup(t, "")

	created := ruleTextDo(t, e, http.MethodPost, ruleTextCreate, `{"title":"structured","repeat":{"freq":"daily","interval":2}}`, token, "", http.StatusCreated)
	id := int64(created["id"].(float64))
	assert.Contains(t, storedRule(t, id), "INTERVAL=2")
	assert.Equal(t, storedRule(t, id), created["repeats_rrule"])

	path := fmt.Sprintf("/api/v2/tasks/%d", id)
	ruleTextDo(t, e, http.MethodPatch, path, `{"repeat":{"freq":"weekly","interval":1}}`, token, ruleTextMerge, http.StatusOK)
	assert.Contains(t, storedRule(t, id), "FREQ=WEEKLY")

	ruleTextDo(t, e, http.MethodPatch, path, `{"repeat":null}`, token, ruleTextMerge, http.StatusOK)
	assert.Empty(t, storedRule(t, id))

	seedStoredRule(t, 1, ruleTextEaster)
	v1 := readTaskV1(t, e, token)
	assert.NotContains(t, v1, "repeats_rrule", "v1's wire shape is frozen")
}
