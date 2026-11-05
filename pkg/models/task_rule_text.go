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
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	// untilUTCDateTime is an RFC 5545 UTC DATE-TIME (form #2, section 3.3.5). A due date is a
	// zone-bound DATE-TIME, so section 3.3.10 requires UNTIL in this form.
	untilUTCDateTime = regexp.MustCompile(`^\d{8}T\d{6}Z$`)
	// numberedWeekday is a BYDAY value with an ordinal, such as 1MO or -1FR.
	numberedWeekday = regexp.MustCompile(`^[+-]?\d+(MO|TU|WE|TH|FR|SA|SU)$`)
	// positiveInteger is RFC 5545's 1*DIGIT, which COUNT and INTERVAL must be (and above 0).
	positiveInteger = regexp.MustCompile(`^\d+$`)
)

// stripRulePrefix trims spaces and one leading, case-insensitive "RRULE:".
func stripRulePrefix(text string) string {
	text = strings.TrimSpace(text)
	const prefix = "RRULE:"
	if len(text) >= len(prefix) && strings.EqualFold(text[:len(prefix)], prefix) {
		text = strings.TrimSpace(text[len(prefix):])
	}
	return text
}

// ruleTextParts tokenises a recur value (the text after "RRULE:") into its parts, keys and
// values upper-cased: RFC 5545 section 3.1 makes names and enumerated values case-insensitive,
// and rrule-go only reads upper case. It returns the parts in order, keyed for lookup, and
// refuses what rrule-go would silently accept: a part without "=", a part given twice (rrule-go
// lets the later one win), a DTSTART part (a separate property, not a recur part; the anchor is
// the due date) and a multi-line value.
func ruleTextParts(text string) (order []string, parts map[string]string, err error) {
	if strings.ContainsAny(text, "\r\n") {
		return nil, nil, ErrInvalidData{Message: "Invalid RRULE: the rule must be a single line of rule parts; DTSTART is not accepted (the due date is the start)"}
	}
	parts = map[string]string{}
	for _, part := range strings.Split(text, ";") {
		key, value, ok := strings.Cut(part, "=")
		if !ok {
			return nil, nil, ErrInvalidData{Message: "Invalid RRULE: the part \"" + part + "\" has no \"=\""}
		}
		key = strings.ToUpper(key)
		if key == "DTSTART" {
			return nil, nil, ErrInvalidData{Message: "Invalid RRULE: DTSTART is not accepted; the due date is the start"}
		}
		if _, dup := parts[key]; dup {
			return nil, nil, ErrInvalidData{Message: "Invalid RRULE: " + key + " is given more than once"}
		}
		parts[key] = strings.ToUpper(value)
		order = append(order, key)
	}
	if _, ok := parts["FREQ"]; !ok {
		return nil, nil, ErrInvalidData{Message: "Invalid RRULE: FREQ is required"}
	}
	return order, parts, nil
}

// checkRuleTextStructure refuses the combinations RFC 5545 section 3.3.10 forbids, and an UNTIL
// that is not a UTC DATE-TIME. A numbered BYDAY outside MONTHLY and YEARLY is left to
// checkRRuleWriteLimits, which refuses it for every write.
func checkRuleTextStructure(parts map[string]string) error {
	freq := parts["FREQ"]
	_, hasCount := parts["COUNT"]
	until, hasUntil := parts["UNTIL"]

	if hasCount && hasUntil {
		return ErrInvalidData{Message: "Invalid RRULE: COUNT and UNTIL cannot both be set"}
	}
	if hasUntil && !untilUTCDateTime.MatchString(until) {
		return ErrInvalidData{Message: "Invalid RRULE: UNTIL " + until + " must be a UTC date-time such as 20271231T235959Z"}
	}
	for _, key := range []string{"COUNT", "INTERVAL"} {
		value, ok := parts[key]
		if !ok {
			continue
		}
		if n, err := strconv.Atoi(value); !positiveInteger.MatchString(value) || err != nil || n < 1 {
			return ErrInvalidData{Message: "Invalid RRULE: " + key + " " + value + " must be a positive integer"}
		}
	}
	if _, ok := parts["BYWEEKNO"]; ok && freq != "YEARLY" {
		return ErrInvalidData{Message: "Invalid RRULE: BYWEEKNO is only allowed with FREQ=YEARLY"}
	}
	if _, ok := parts["BYYEARDAY"]; ok && (freq == "DAILY" || freq == "WEEKLY" || freq == "MONTHLY") {
		return ErrInvalidData{Message: "Invalid RRULE: BYYEARDAY is not allowed with FREQ=" + freq}
	}
	if _, ok := parts["BYMONTHDAY"]; ok && freq == "WEEKLY" {
		return ErrInvalidData{Message: "Invalid RRULE: BYMONTHDAY is not allowed with FREQ=WEEKLY"}
	}
	if byDay, ok := parts["BYDAY"]; ok && freq == "YEARLY" {
		if _, weekNo := parts["BYWEEKNO"]; weekNo {
			for _, day := range strings.Split(byDay, ",") {
				if numberedWeekday.MatchString(day) {
					return ErrInvalidData{Message: "Invalid RRULE: BYDAY " + day + " has a position, which is not allowed with FREQ=YEARLY and BYWEEKNO"}
				}
			}
		}
	}
	if _, ok := parts["BYSETPOS"]; ok {
		other := false
		for key := range parts {
			if key != "BYSETPOS" && strings.HasPrefix(key, "BY") {
				other = true
				break
			}
		}
		if !other {
			return ErrInvalidData{Message: "Invalid RRULE: BYSETPOS needs another BYxxx part to pick from"}
		}
	}
	return nil
}

// ValidateRuleText checks RFC 5545 rule text sent by a v2 client and returns the canonical form
// to store: rrule-go's RRuleString, the spelling toRRule and CalDAV's NormalizeRRule store, so
// reads show one spelling. A leading "RRULE:" is stripped. Beyond what every write checks
// (validateRRule: syntax, length, bounds, policy), it refuses what RFC 5545 section 3.3.10
// forbids and rrule-go accepts: a repeated part, COUNT with UNTIL, a COUNT or INTERVAL that is
// not a positive integer, BY* parts the frequency does not allow, BYSETPOS on its own, a DTSTART,
// and an UNTIL that is not a UTC date-time. BYEASTER, rrule-go's extension, is accepted. Every
// refusal is ErrInvalidData naming the part. Empty text is refused; clearing is the caller's
// decision (resolveV2Repeat).
func ValidateRuleText(text string) (string, error) {
	stripped := stripRulePrefix(text)
	if stripped == "" {
		return "", ErrInvalidData{Message: "Invalid RRULE: the rule is empty"}
	}
	// validateRRule checks this too; checking first spares tokenising an oversized body.
	if len(stripped) > maxRRuleLength {
		return "", ErrInvalidData{Message: fmt.Sprintf("Invalid RRULE: the rule is %d characters long; at most %d characters are allowed", len(stripped), maxRRuleLength)}
	}
	order, parts, err := ruleTextParts(stripped)
	if err != nil {
		return "", err
	}
	if err := checkRuleTextStructure(parts); err != nil {
		return "", err
	}
	rule := joinRuleTextParts(order, parts)
	if err := validateRRule(rule); err != nil {
		return "", err
	}
	opt, err := parseRuleText(rule)
	if err != nil {
		// validateRRule parsed the same text.
		return "", ErrInvalidData{Message: "Invalid RRULE: " + err.Error()}
	}
	return opt.RRuleString(), nil
}

// ValidateRepeat checks a structured repeat exactly as a structured write does and returns the
// rule text it would store ("" for nil or no frequency).
func ValidateRepeat(r *TaskRepeat) (string, error) {
	text, err := r.toRRule()
	if err != nil {
		return "", err
	}
	if err := validateRRule(text); err != nil {
		return "", err
	}
	return text, nil
}

// joinRuleTextParts rebuilds the recur value from ruleTextParts' output.
func joinRuleTextParts(order []string, parts map[string]string) string {
	joined := make([]string, len(order))
	for i, key := range order {
		joined[i] = key + "=" + parts[key]
	}
	return strings.Join(joined, ";")
}

// isValidatedRuleEcho reports whether canon, rule text that passed ValidateRuleText (in its
// canonical form), is the stored rule sent back unchanged: the same rule spelled differently
// (case, part order, an explicit INTERVAL=1, list order; compared as normaliseROption, so a
// stored DTSTART or BYEASTER counts), or the stored rule with a stale, higher COUNT
// (isStaleRawCount, the stale-count rule: so COUNT cannot be raised through rule text alone either).
// Only validated text gets this lenient comparison, so text ValidateRuleText refuses (a
// floating UNTIL, COUNT=0, a duplicated part) is never read as the stored rule.
func isValidatedRuleEcho(stored, canon string) bool {
	if stored == "" {
		return false
	}
	storedOpt, err := parseRuleText(stored)
	if err != nil {
		return false
	}
	canonOpt, err := parseRuleText(canon)
	if err != nil {
		return false
	}
	return normaliseROption(canonOpt) == normaliseROption(storedOpt) || isStaleRawCount(stored, canon)
}

// resolveV2Repeat maps a v2 write (structured repeat and/or rule text) onto
// resolveIncomingRepeat's inputs. text == nil means the v2 field was absent and returns
// (repeat, raw) unchanged, so every other caller (CalDAV, import, duplication, bulk) is
// untouched.
//
// Otherwise, with "unchanged" defined by the resolver's own helpers:
//   - text equal to the stored rule hands the decision to repeat, exactly as without the field:
//     a nil repeat clears, an echoed one keeps. This is the frontend's clear path, whose PUT
//     carries the echoed text and a nil repeat. The exact text is an echo without validation,
//     so a stored rule a later policy rejects never fails an unrelated update; a respelled or
//     stale-COUNT echo must first pass ValidateRuleText (isValidatedRuleEcho);
//   - "" clears, unless repeat changes the rule, which is refused as two conflicting intents;
//   - changed text is validated (ValidateRuleText) and passed as the raw rule. A nil or echoed
//     repeat drops out, so the text wins; a changed repeat is passed along and the resolver's
//     rule 1 refuses it unless it agrees with the text.
func resolveV2Repeat(stored string, repeat *TaskRepeat, raw string, text *string) (*TaskRepeat, string, error) {
	if text == nil {
		return repeat, raw, nil
	}
	stripped := stripRulePrefix(*text)
	if stripped == stored {
		return repeat, "", nil
	}

	structuredEcho := false
	if repeat != nil && stored != "" {
		if projection := taskRepeatFromRRule(stored); projection != nil {
			structuredEcho = repeatsEqual(repeat, projection) || isStaleCount(repeat, projection)
		}
	}

	if stripped == "" {
		if repeat == nil || structuredEcho {
			return nil, "", nil
		}
		return nil, "", ErrInvalidData{Message: "repeat and repeats_rrule disagree: repeats_rrule clears the recurrence but repeat changes it; send only one of them."}
	}

	canon, err := ValidateRuleText(stripped)
	if err != nil {
		return nil, "", err
	}
	if isValidatedRuleEcho(stored, canon) {
		return repeat, "", nil
	}
	if repeat == nil || structuredEcho {
		return nil, canon, nil
	}
	return repeat, canon, nil
}
