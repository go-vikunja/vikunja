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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/kentonium3/rrule-go"
)

// parseRuleText parses stored or incoming RRULE text into an option exactly as rrule-go v1.8.3
// did, on the stricter v1.9.0 parser. Every non-test parse in this package goes through it
// (TestParseSitesUseParseRuleText).
//
// v1.9.0 refuses a DTSTART whose time form does not match UNTIL's, and re-spells a floating or
// DATE UNTIL without its Z. Vikunja never uses a rule-text DTSTART (the task's due date is the
// anchor), but older rows, imports and v1 writes still carry one, and normaliseROption still
// reads it (a stored DTSTART keeps a rule read-only and makes it differ from a DTSTART-free
// echo). So this reads the DTSTART the way v1.8.3 did and hands v1.9.0 only the rest:
//
//   - at most two lines; in two-line text the first is the DTSTART property, DTSTART:<value> or
//     DTSTART;TZID=<zone>:<value>, and the value is a DATE, floating or UTC date-time, read in UTC
//     or the TZID's zone;
//   - an inline DTSTART=<value> part of the rule line (key exactly DTSTART) is removed and read
//     in UTC; the last one read wins;
//   - every UNTIL part is read too, so a malformed earlier UNTIL is refused, as v1.8.3 did
//     (v1.9.0 reads only the last);
//   - the remainder must not change under the parser's own trimming (edge whitespace, a second
//     RRULE: prefix), which v1.8.3 refused.
//
// The grammar is frozen to teambition/rrule-go v1.8.3 (str.go: StrToROptionInLocation,
// StrToDtStart, strToTimeInLoc, parseTZID, processRRuleName; MIT licence). Do not extend it.
// testdata/rrule_v183_parse_golden.json is v1.8.3's own parse of 637 texts, the contract this
// must keep. It can go once no stored rule carries a DTSTART and the library is not stricter
// than v1.8.3 for the remaining texts. Error messages are this
// function's own and reach v1 clients inside "Invalid RRULE: ..." (validateRRuleSyntax); no test,
// frontend code or known client depends on their wording (checked 2026-10-29).
func parseRuleText(text string) (*rrule.ROption, error) {
	lines := strings.Split(strings.TrimSpace(text), "\n")
	var dtstartLine, ruleLine string
	switch len(lines) {
	case 1:
		ruleLine = lines[0]
	case 2:
		dtstartLine, ruleLine = lines[0], lines[1]
	default:
		return nil, errors.New("invalid RRULE string")
	}

	var dtstart time.Time
	if dtstartLine != "" {
		name, err := rulePropertyName(dtstartLine)
		if err != nil {
			return nil, err
		}
		if name != "DTSTART" {
			return nil, fmt.Errorf("expect DTSTART but: %s", name)
		}
		if dtstart, err = parseDTStartValue(dtstartLine[len(name)+1:]); err != nil {
			return nil, fmt.Errorf("invalid DTSTART: %w", err)
		}
	}

	ruleLine = strings.TrimPrefix(ruleLine, "RRULE:")
	parts := strings.Split(ruleLine, ";")
	kept := make([]string, 0, len(parts))
	for _, part := range parts {
		kv := strings.Split(part, "=")
		if len(kv) == 2 && kv[0] == "UNTIL" && kv[1] != "" {
			if _, err := parseRuleTime(kv[1], time.UTC); err != nil {
				return nil, fmt.Errorf("invalid UNTIL: %w", err)
			}
		}
		if len(kv) == 2 && kv[0] == "DTSTART" {
			if kv[1] == "" {
				return nil, errors.New("DTSTART option has no value")
			}
			t, err := parseRuleTime(kv[1], time.UTC)
			if err != nil {
				return nil, fmt.Errorf("invalid DTSTART: %w", err)
			}
			dtstart = t
			continue
		}
		kept = append(kept, part)
	}

	rest := strings.Join(kept, ";")
	if strings.TrimSpace(rest) != rest || strings.HasPrefix(rest, "RRULE:") {
		return nil, fmt.Errorf("unknown RRULE property in %q", rest)
	}
	opt, err := rrule.StrToROption(rest)
	if err != nil {
		return nil, err
	}
	opt.Dtstart = dtstart
	return opt, nil
}

// rulePropertyName is the upper-cased name before a property line's first ';' or ':'.
func rulePropertyName(line string) (string, error) {
	line = strings.ToUpper(strings.TrimSpace(line))
	n := strings.IndexAny(line, ";:")
	if n <= 0 || strings.Contains(line[:n], "=") {
		return "", fmt.Errorf("bad format %v", line)
	}
	return line[:n], nil
}

// parseDTStartValue reads a DTSTART property's parameters and value: <value> in UTC, or
// TZID=<zone>:<value> in that zone. No other parameter is accepted.
func parseDTStartValue(s string) (time.Time, error) {
	fields := strings.Split(s, ":")
	switch len(fields) {
	case 1:
		return parseRuleTime(fields[0], time.UTC)
	case 2:
		zone, ok := strings.CutPrefix(fields[0], "TZID=")
		if !ok || zone == "" {
			return time.Time{}, errors.New("bad TZID parameter format")
		}
		loc, err := time.LoadLocation(zone)
		if err != nil {
			return time.Time{}, err
		}
		return parseRuleTime(fields[1], loc)
	default:
		return time.Time{}, errors.New("bad format")
	}
}

// parseRuleTime reads a DATE or floating date-time in loc, or a UTC date-time, by length.
func parseRuleTime(s string, loc *time.Location) (time.Time, error) {
	switch len(s) {
	case len("20060102"):
		return time.ParseInLocation("20060102", s, loc)
	case len("20060102T150405"):
		return time.ParseInLocation("20060102T150405", s, loc)
	}
	return time.Parse("20060102T150405Z", s)
}
