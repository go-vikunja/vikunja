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

// Package errorreport derives stable Sentry fingerprints from errors.
//
// Every error we report reaches Sentry through one of three call sites, so the captured stack trace
// is the same for all of them. Sentry's default grouping therefore folds completely unrelated errors
// into a single issue. Fingerprinting by what the error *is* instead of where we reported it keeps
// those apart.
package errorreport

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/web"

	"github.com/getsentry/sentry-go"
	"github.com/go-sql-driver/mysql"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
)

// DefaultFingerprint tells Sentry to keep its own grouping for this event.
const DefaultFingerprint = "{{ default }}"

// maxNormalizedLen keeps a fingerprint component bounded: some driver errors carry a whole query.
const maxNormalizedLen = 120

var (
	urlRE    = regexp.MustCompile(`\b[a-z][a-z0-9+.-]*://\S+`)
	emailRE  = regexp.MustCompile(`[\w.+-]+@[\w-]+\.[\w.-]+`)
	uuidRE   = regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	hexRE    = regexp.MustCompile(`(?i)\b[0-9a-f]{16,}\b`)
	quotedRE = regexp.MustCompile("'[^']*'|\"[^\"]*\"|`[^`]*`")
	pathRE   = regexp.MustCompile(`(?:[a-zA-Z]:)?[/\\][^\s'"` + "`" + `]+`)
	numberRE = regexp.MustCompile(`\d+`)
	spaceRE  = regexp.MustCompile(`\s+`)
	letterRE = regexp.MustCompile(`\p{L}`)
)

// Normalize strips everything variable from an error message - ids, quoted identifiers and literals,
// paths, urls, mail addresses - so that the same failure always yields the same fingerprint and no
// user data ends up in the grouping key.
func Normalize(msg string) string {
	for _, re := range []*regexp.Regexp{urlRE, emailRE, uuidRE, hexRE, quotedRE, pathRE, numberRE} {
		msg = re.ReplaceAllString(msg, "?")
	}
	msg = strings.ToLower(strings.TrimSpace(spaceRE.ReplaceAllString(msg, " ")))

	runes := []rune(msg)
	if len(runes) > maxNormalizedLen {
		return string(runes[:maxNormalizedLen])
	}
	return msg
}

// TypeName returns the Go type of the innermost wrapped error.
func TypeName(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("%T", innermost(err))
}

// Fingerprint derives a stable grouping key from err, falling back to Sentry's default grouping when
// the error carries nothing usable.
func Fingerprint(err error) []string {
	if parts := fingerprintParts(err); len(parts) > 0 {
		return parts
	}
	return []string{DefaultFingerprint}
}

// Apply groups the event by what err is instead of where it was reported.
func Apply(scope *sentry.Scope, err error) {
	ApplyFingerprint(scope, err, Fingerprint(err)...)
}

// ApplyFingerprint is Apply with an explicit grouping key, for wrapper errors whose own message says
// nothing about the failure (a poisoned event, say).
func ApplyFingerprint(scope *sentry.Scope, err error, fingerprint ...string) {
	if len(fingerprint) == 0 {
		fingerprint = []string{DefaultFingerprint}
	}
	scope.SetFingerprint(fingerprint)
	if t := TypeName(err); t != "" {
		scope.SetTag("error.type", t)
	}
}

func fingerprintParts(err error) []string {
	if err == nil {
		return nil
	}

	// A recovered panic has a real stack of its own, but it still reaches Sentry through our
	// handler, so it needs the same treatment - just kept apart from the same error returned normally.
	var panicErr *middleware.PanicStackError
	if errors.As(err, &panicErr) && panicErr.Err != nil {
		return append([]string{"panic"}, fingerprintParts(panicErr.Err)...)
	}

	var domainErr web.HTTPErrorProcessor
	if errors.As(err, &domainErr) {
		if code := domainErr.HTTPError().Code; code != 0 {
			return []string{"vikunja", strconv.Itoa(code)}
		}
		return []string{"vikunja", fmt.Sprintf("%T", domainErr)}
	}

	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		return []string{"postgres", string(pgErr.Code), Normalize(pgErr.Message)}
	}

	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) {
		return []string{"mysql", strconv.FormatUint(uint64(myErr.Number), 10), Normalize(myErr.Message)}
	}

	var liteErr sqlite3.Error
	if errors.As(err, &liteErr) {
		return []string{"sqlite", strconv.Itoa(int(liteErr.Code)), strconv.Itoa(int(liteErr.ExtendedCode)), Normalize(liteErr.Error())}
	}

	var echoErr *echo.HTTPError
	if errors.As(err, &echoErr) {
		return []string{"echo", strconv.Itoa(echoErr.Code), Normalize(echoErr.Message)}
	}

	inner := innermost(err)
	message := Normalize(inner.Error())
	// A message that normalises to nothing but placeholders tells us nothing about the error.
	if !letterRE.MatchString(message) {
		return nil
	}
	return []string{fmt.Sprintf("%T", inner), message}
}

func innermost(err error) error {
	for {
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			return err
		}
		err = unwrapped
	}
}
