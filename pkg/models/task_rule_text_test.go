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

package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func strPtr(s string) *string { return &s }

// TestValidateRuleText pins rule-text validation. Each refused row sits next to an
// accepted row built from the same base rule, so a refusal is never vacuous.
func TestValidateRuleText(t *testing.T) {
	cases := []struct {
		name    string
		text    string
		want    string // canonical stored form; ignored when wantErr is set
		wantErr string // substring of the ErrInvalidData message
	}{
		// Prefix, spacing and case.
		{name: "prefix stripped", text: "RRULE:FREQ=WEEKLY;BYDAY=MO", want: "FREQ=WEEKLY;BYDAY=MO"},
		{name: "lowercase prefix and spaces", text: "  rrule:FREQ=WEEKLY;BYDAY=MO  ", want: "FREQ=WEEKLY;BYDAY=MO"},
		{name: "lowercase keys and values", text: "freq=weekly;byday=mo", want: "FREQ=WEEKLY;BYDAY=MO"},
		{name: "part order is canonicalised", text: "BYDAY=MO;INTERVAL=2;FREQ=WEEKLY", want: "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO"},
		{name: "empty is refused", text: "", wantErr: "empty"},
		{name: "a bare prefix is refused", text: "RRULE:", wantErr: "empty"},

		// BYEASTER: the documented extension.
		{name: "BYEASTER accepted", text: "FREQ=YEARLY;BYEASTER=1", want: "FREQ=YEARLY;BYEASTER=1"},
		{name: "BYEASTER lowercase accepted", text: "freq=yearly;byeaster=-2", want: "FREQ=YEARLY;BYEASTER=-2"},
		{name: "BYEASTER out of bounds", text: "FREQ=YEARLY;BYEASTER=300", wantErr: "BYEASTER value 300 is out of range"},

		// Tokenising.
		{name: "count with a value", text: "FREQ=DAILY;COUNT=5", want: "FREQ=DAILY;COUNT=5"},
		{name: "part without =", text: "FREQ=DAILY;COUNT", wantErr: `"COUNT" has no "="`},
		{name: "trailing semicolon", text: "FREQ=DAILY;", wantErr: `has no "="`},
		{name: "single freq", text: "FREQ=DAILY", want: "FREQ=DAILY"},
		{name: "freq given twice", text: "FREQ=DAILY;FREQ=WEEKLY", wantErr: "FREQ is given more than once"},
		{name: "a part given twice in different case", text: "FREQ=DAILY;INTERVAL=2;interval=3", wantErr: "INTERVAL is given more than once"},
		{name: "DTSTART part", text: "DTSTART=20270101T090000Z;FREQ=DAILY", wantErr: "DTSTART is not accepted"},
		{name: "DTSTART line", text: "DTSTART:20270101T090000Z\nRRULE:FREQ=DAILY", wantErr: "single line"},
		{name: "freq with an interval", text: "FREQ=DAILY;INTERVAL=2", want: "FREQ=DAILY;INTERVAL=2"},
		{name: "freq missing", text: "INTERVAL=2", wantErr: "FREQ is required"},

		// UNTIL must be a UTC date-time.
		{name: "UTC until", text: "FREQ=DAILY;UNTIL=20271231T235959Z", want: "FREQ=DAILY;UNTIL=20271231T235959Z"},
		{name: "floating until", text: "FREQ=DAILY;UNTIL=20271231T235959", wantErr: "UNTIL 20271231T235959 must be a UTC date-time"},
		{name: "date until", text: "FREQ=DAILY;UNTIL=20271231", wantErr: "UNTIL 20271231 must be a UTC date-time"},

		// RFC 5545 section 3.3.10 structure.
		{name: "count and until together", text: "FREQ=DAILY;COUNT=5;UNTIL=20271231T235959Z", wantErr: "COUNT and UNTIL cannot both be set"},
		{name: "count zero", text: "FREQ=DAILY;COUNT=0", wantErr: "COUNT 0 must be a positive integer"},
		{name: "count negative", text: "FREQ=DAILY;COUNT=-1", wantErr: "COUNT -1 must be a positive integer"},
		{name: "count not a number", text: "FREQ=DAILY;COUNT=five", wantErr: "COUNT FIVE must be a positive integer"},
		{name: "interval zero", text: "FREQ=DAILY;INTERVAL=0", wantErr: "INTERVAL 0 must be a positive integer"},
		{name: "interval with a sign", text: "FREQ=DAILY;INTERVAL=+2", wantErr: "INTERVAL +2 must be a positive integer"},
		{name: "BYWEEKNO with yearly", text: "FREQ=YEARLY;BYWEEKNO=20", want: "FREQ=YEARLY;BYWEEKNO=20"},
		{name: "BYWEEKNO with monthly", text: "FREQ=MONTHLY;BYWEEKNO=20", wantErr: "BYWEEKNO is only allowed with FREQ=YEARLY"},
		{name: "BYYEARDAY with yearly", text: "FREQ=YEARLY;BYYEARDAY=100", want: "FREQ=YEARLY;BYYEARDAY=100"},
		{name: "BYYEARDAY with monthly", text: "FREQ=MONTHLY;BYYEARDAY=100", wantErr: "BYYEARDAY is not allowed with FREQ=MONTHLY"},
		{name: "BYYEARDAY with weekly", text: "FREQ=WEEKLY;BYYEARDAY=100", wantErr: "BYYEARDAY is not allowed with FREQ=WEEKLY"},
		{name: "BYYEARDAY with daily", text: "FREQ=DAILY;BYYEARDAY=100", wantErr: "BYYEARDAY is not allowed with FREQ=DAILY"},
		{name: "BYMONTHDAY with monthly", text: "FREQ=MONTHLY;BYMONTHDAY=1", want: "FREQ=MONTHLY;BYMONTHDAY=1"},
		{name: "BYMONTHDAY with weekly", text: "FREQ=WEEKLY;BYMONTHDAY=1", wantErr: "BYMONTHDAY is not allowed with FREQ=WEEKLY"},
		{name: "plain BYDAY with yearly and BYWEEKNO", text: "FREQ=YEARLY;BYWEEKNO=20;BYDAY=MO", want: "FREQ=YEARLY;BYWEEKNO=20;BYDAY=MO"},
		{name: "numbered BYDAY with yearly and BYWEEKNO", text: "FREQ=YEARLY;BYWEEKNO=20;BYDAY=1MO", wantErr: "BYDAY 1MO has a position, which is not allowed with FREQ=YEARLY and BYWEEKNO"},
		{name: "numbered BYDAY with monthly", text: "FREQ=MONTHLY;BYDAY=1MO", want: "FREQ=MONTHLY;BYDAY=+1MO"},
		// Refused by checkRRuleWriteLimits, with the message a structured write gets.
		{name: "numbered BYDAY with weekly", text: "FREQ=WEEKLY;BYDAY=1MO", wantErr: "positions are only allowed for MONTHLY and YEARLY rules"},
		{name: "BYSETPOS with BYDAY", text: "FREQ=MONTHLY;BYDAY=MO,TU,WE,TH,FR;BYSETPOS=-1", want: "FREQ=MONTHLY;BYSETPOS=-1;BYDAY=MO,TU,WE,TH,FR"},
		{name: "BYSETPOS alone", text: "FREQ=MONTHLY;BYSETPOS=-1", wantErr: "BYSETPOS needs another BYxxx part"},

		// validateRRule: syntax, length, bounds and policy, as every write.
		{name: "plain hourly", text: "FREQ=HOURLY", want: "FREQ=HOURLY"},
		{name: "hourly with BYHOUR", text: "FREQ=HOURLY;BYHOUR=9", wantErr: "sub-daily frequencies cannot be combined with BY* components"},
		{name: "unknown frequency", text: "FREQ=FORTNIGHTLY", wantErr: "Invalid RRULE"},
		{name: "unknown part", text: "FREQ=DAILY;X-NAME=1", wantErr: "Invalid RRULE"},
		{name: "too long", text: "FREQ=YEARLY;BYMONTHDAY=1" + strings.Repeat(",1", 250), wantErr: "at most 500 characters"},
		{name: "too long before tokenising", text: strings.Repeat("FREQ=DAILY;", 50), wantErr: "the rule is 550 characters long; at most 500"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := ValidateRuleText(c.text)
			if c.wantErr != "" {
				require.Error(t, err)
				require.True(t, IsErrInvalidData(err), "want ErrInvalidData, got %T: %v", err, err)
				assert.Contains(t, err.(ErrInvalidData).Message, c.wantErr)
				assert.True(t, strings.HasPrefix(err.(ErrInvalidData).Message, "Invalid RRULE: "), "message style: %q", err.(ErrInvalidData).Message)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.want, got)
		})
	}

	t.Run("a policy refusal reads exactly as a structured write's", func(t *testing.T) {
		_, textErr := ValidateRuleText("FREQ=HOURLY;BYHOUR=9")
		_, structErr := ValidateRepeat(&TaskRepeat{Freq: "hourly", Interval: 1, ByHour: []int{9}})
		require.Error(t, textErr)
		require.Error(t, structErr)
		assert.Equal(t, structErr.(ErrInvalidData).Message, textErr.(ErrInvalidData).Message)
	})
}

// TestValidateRepeat pins that a structured repeat is validated exactly as a structured write
// (resolveIncomingRepeat's rule 5) validates it.
func TestValidateRepeat(t *testing.T) {
	text, err := ValidateRepeat(nil)
	require.NoError(t, err)
	assert.Empty(t, text)

	text, err = ValidateRepeat(&TaskRepeat{Freq: "weekly", Interval: 2, ByDay: []string{"mo"}})
	require.NoError(t, err)
	assert.Equal(t, "FREQ=WEEKLY;INTERVAL=2;BYDAY=MO", text)
	decision, err := resolveIncomingRepeat("", &TaskRepeat{Freq: "weekly", Interval: 2, ByDay: []string{"mo"}}, "")
	require.NoError(t, err)
	assert.Equal(t, decision.Repeats, text, "the same text a structured create stores")

	_, err = ValidateRepeat(&TaskRepeat{Freq: "fortnightly"})
	require.Error(t, err)
	assert.Contains(t, err.(ErrInvalidData).Message, "Invalid frequency")

	_, err = ValidateRepeat(&TaskRepeat{Freq: "minutely", Interval: 1, ByDay: []string{"mo"}})
	require.Error(t, err)
	assert.Contains(t, err.(ErrInvalidData).Message, "sub-daily")
}

// TestResolveV2Repeat pins the v2 write-precedence rows, one case per row plus the
// echo variants.
func TestResolveV2Repeat(t *testing.T) {
	const (
		weekly      = "FREQ=WEEKLY;BYDAY=MO"
		countTwo    = "FREQ=DAILY;COUNT=2"
		easter      = "FREQ=YEARLY;BYEASTER=0"
		withStart   = "FREQ=DAILY;DTSTART=20260101T090000Z"
		disagreeMsg = "repeat and repeats_rrule disagree: repeats_rrule clears the recurrence but repeat changes it; send only one of them."
	)
	changedRepeat := &TaskRepeat{Freq: "daily", Interval: 3}

	cases := []struct {
		name       string
		stored     string
		repeat     *TaskRepeat
		raw        string
		text       *string
		wantRepeat *TaskRepeat
		wantRaw    string
		wantErr    string
	}{
		// Row 1: text absent leaves every caller as it is today.
		{name: "absent text passes repeat and raw through", stored: weekly, repeat: changedRepeat, raw: "FREQ=DAILY", wantRepeat: changedRepeat, wantRaw: "FREQ=DAILY"},
		{name: "absent text with nothing else", stored: weekly},

		// Row 2: a text echo hands the decision to repeat.
		{name: "echo with nil repeat (the frontend clear shape)", stored: weekly, text: strPtr(weekly)},
		{name: "echo with the projection (a PATCH echo)", stored: weekly, repeat: taskRepeatFromRRule(weekly), text: strPtr(weekly), wantRepeat: taskRepeatFromRRule(weekly)},
		{name: "echo with a changed repeat lets repeat decide", stored: weekly, repeat: changedRepeat, text: strPtr(weekly), wantRepeat: changedRepeat},
		{name: "echo with the RRULE: prefix", stored: weekly, text: strPtr("RRULE:" + weekly)},
		{name: "echo spelled differently", stored: weekly, text: strPtr("freq=weekly;byday=MO")},
		{name: "echo in another part order", stored: weekly, text: strPtr("BYDAY=MO;FREQ=WEEKLY")},
		{name: "echo of a read-only rule", stored: easter, repeat: taskRepeatFromRRule(easter), text: strPtr(easter), wantRepeat: taskRepeatFromRRule(easter)},
		// The stale-count rule, applied to text: a higher COUNT on otherwise-equal text is a
		// client echoing what it read before a completion lowered COUNT, so it is an echo. Raising
		// COUNT through rule text alone is therefore ignored (clear and set the rule to raise it).
		{name: "stale higher count is an echo", stored: countTwo, text: strPtr("FREQ=DAILY;COUNT=5")},
		{name: "lower count is a change", stored: countTwo, text: strPtr("FREQ=DAILY;COUNT=1"), wantRaw: "FREQ=DAILY;COUNT=1"},
		{name: "empty on create is an echo of no rule", stored: "", repeat: changedRepeat, text: strPtr(""), wantRepeat: changedRepeat},
		{name: "empty on create with no repeat", stored: "", text: strPtr("")},

		// Rows 3 and 4: "" clears, unless repeat changes the rule.
		{name: "empty with nil repeat clears", stored: weekly, text: strPtr("")},
		{name: "empty prefix only clears", stored: weekly, text: strPtr("RRULE:")},
		{name: "empty with an echoed repeat clears", stored: weekly, repeat: taskRepeatFromRRule(weekly), text: strPtr("")},
		{name: "empty with a stale-count echo clears", stored: countTwo, repeat: &TaskRepeat{Freq: "daily", Interval: 1, Count: 4}, text: strPtr("")},
		{name: "empty with a changed repeat is refused", stored: weekly, repeat: changedRepeat, text: strPtr(""), wantErr: disagreeMsg},

		// Rows 5 and 6: changed text is validated and wins over a nil or echoed repeat.
		{name: "changed text with nil repeat", stored: weekly, text: strPtr("freq=weekly;byday=tu"), wantRaw: "FREQ=WEEKLY;BYDAY=TU"},
		{name: "changed text with the old projection", stored: weekly, repeat: taskRepeatFromRRule(weekly), text: strPtr("FREQ=WEEKLY;BYDAY=TU"), wantRaw: "FREQ=WEEKLY;BYDAY=TU"},
		{name: "changed text with a read-only projection", stored: easter, repeat: taskRepeatFromRRule(easter), text: strPtr("FREQ=YEARLY;BYEASTER=1"), wantRaw: "FREQ=YEARLY;BYEASTER=1"},
		{name: "changed text with a changed repeat passes both on", stored: weekly, repeat: changedRepeat, text: strPtr("FREQ=WEEKLY;BYDAY=TU"), wantRepeat: changedRepeat, wantRaw: "FREQ=WEEKLY;BYDAY=TU"},
		{name: "changed text on create", stored: "", text: strPtr("FREQ=YEARLY;BYEASTER=1"), wantRaw: "FREQ=YEARLY;BYEASTER=1"},
		{name: "changed invalid text is refused", stored: weekly, text: strPtr("FREQ=WEEKLY;BYMONTHDAY=1"), wantErr: "BYMONTHDAY is not allowed with FREQ=WEEKLY"},
		// rrule-go would let the later FREQ win and read this as the stored rule.
		{name: "a duplicated part is never an echo", stored: "FREQ=WEEKLY", text: strPtr("FREQ=DAILY;FREQ=WEEKLY"), wantErr: "FREQ is given more than once"},
		// RRuleString drops DTSTART, so dropping it must not read as an echo.
		{name: "text dropping a stored DTSTART is a change", stored: withStart, text: strPtr("FREQ=DAILY"), wantRaw: "FREQ=DAILY"},
		{name: "exact echo of a stored DTSTART rule", stored: withStart, text: strPtr(withStart)},

		// Only text that passes ValidateRuleText can be a respelled or stale-COUNT echo.
		// The exact text stays an echo without validation.
		{name: "a floating UNTIL is refused, not read as the stored UTC one", stored: "FREQ=DAILY;UNTIL=20271231T235959Z", repeat: taskRepeatFromRRule("FREQ=DAILY;UNTIL=20271231T235959Z"), text: strPtr("FREQ=DAILY;UNTIL=20271231T235959"), wantErr: "must be a UTC date-time"},
		{name: "a lowercase UTC UNTIL is a respelled echo", stored: "FREQ=DAILY;UNTIL=20271231T235959Z", text: strPtr("freq=daily;until=20271231t235959z")},
		{name: "COUNT=0 is refused, not read as the stored unlimited rule", stored: "FREQ=DAILY", text: strPtr("FREQ=DAILY;COUNT=0"), wantErr: "COUNT 0 must be a positive integer"},
		{name: "a stale COUNT above the limit is refused", stored: countTwo, text: strPtr("FREQ=DAILY;COUNT=10001"), wantErr: "COUNT 10001 is too large"},
		{name: "a respelled policy-rejected rule is refused", stored: "FREQ=HOURLY;BYHOUR=9", text: strPtr("byhour=9;freq=hourly"), wantErr: "sub-daily"},
		{name: "the exact policy-rejected rule is an echo", stored: "FREQ=HOURLY;BYHOUR=9", text: strPtr("FREQ=HOURLY;BYHOUR=9")},
		{name: "an explicit INTERVAL=1 is an echo of none", stored: "FREQ=DAILY;INTERVAL=1", text: strPtr("FREQ=DAILY")},
		{name: "a reordered BYDAY list is an echo", stored: "FREQ=WEEKLY;BYDAY=MO,FR", text: strPtr("FREQ=WEEKLY;BYDAY=FR,MO")},
		{name: "a different interval is a change", stored: "FREQ=DAILY;INTERVAL=1", text: strPtr("FREQ=DAILY;INTERVAL=2"), wantRaw: "FREQ=DAILY;INTERVAL=2"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			gotRepeat, gotRaw, err := resolveV2Repeat(c.stored, c.repeat, c.raw, c.text)
			if c.wantErr != "" {
				require.Error(t, err)
				require.True(t, IsErrInvalidData(err), "want ErrInvalidData, got %T: %v", err, err)
				assert.Contains(t, err.(ErrInvalidData).Message, c.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, c.wantRepeat, gotRepeat)
			assert.Equal(t, c.wantRaw, gotRaw)
		})
	}

	t.Run("absent text returns the same repeat pointer", func(t *testing.T) {
		gotRepeat, _, err := resolveV2Repeat(weekly, changedRepeat, "", nil)
		require.NoError(t, err)
		assert.Same(t, changedRepeat, gotRepeat)
	})
}
