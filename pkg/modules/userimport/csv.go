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

package userimport

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Column sizes of the users table. Longer values are rejected (id, mail) or truncated (the rest),
// because PostgreSQL and MySQL refuse over-length values and one bad row would roll back the run.
const (
	maxMailLength  = 250
	maxFieldLength = 250
	maxIDLength    = 64
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Person is one valid row of the user list.
type Person struct {
	// ID is the Microsoft Entra object id.
	ID string
	// Mail is the email address the person is matched on, trimmed but in its original case.
	Mail              string
	DisplayName       string
	JobTitle          string
	Department        string
	UserPrincipalName string
	// Enabled is false when the accountEnabled column says so. Such a person is treated as not listed.
	Enabled bool
}

// ParseResult is the outcome of reading a user list.
type ParseResult struct {
	People           []*Person
	SkippedInvalid   int
	SkippedDuplicate int
}

// ParseCSV reads the user list. The first row is the header; columns are found by name,
// case-insensitively, and unknown columns are ignored. The columns id, mail and displayName are
// required; jobTitle, department, userPrincipalName and accountEnabled are optional.
//
// Rows without a valid id or mail are skipped and counted. Rows that repeat an earlier id or
// mail (case-insensitive) are skipped and counted too, the first one wins.
func ParseCSV(r io.Reader) (*ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("could not read the user list: %w", err)
	}
	// Exports from Microsoft Graph and Excel start with a byte order mark.
	data = bytes.TrimPrefix(data, utf8BOM)

	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1 // rows may be shorter or longer than the header
	reader.TrimLeadingSpace = true

	header, err := reader.Read()
	if errors.Is(err, io.EOF) {
		return nil, errors.New("the user list is empty")
	}
	if err != nil {
		return nil, fmt.Errorf("could not read the header row: %w", err)
	}

	columns := make(map[string]int, len(header))
	for i, name := range header {
		key := strings.ToLower(strings.TrimSpace(clean(name)))
		if _, exists := columns[key]; !exists {
			columns[key] = i
		}
	}
	for _, required := range []string{"id", "mail", "displayname"} {
		if _, ok := columns[required]; !ok {
			return nil, fmt.Errorf("the required column %q is missing from the header row", required)
		}
	}

	field := func(record []string, name string) string {
		idx, ok := columns[name]
		if !ok || idx >= len(record) {
			return ""
		}
		return strings.TrimSpace(clean(record[idx]))
	}

	result := &ParseResult{}
	seenMail := make(map[string]struct{})
	seenID := make(map[string]struct{})

	for row := 2; ; row++ {
		record, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not read row %d: %w", row, err)
		}

		id := field(record, "id")
		mailAddress := field(record, "mail")
		if id == "" || len(id) > maxIDLength || !validMail(mailAddress) {
			result.SkippedInvalid++
			continue
		}

		mailKey := normalizeMail(mailAddress)
		_, mailSeen := seenMail[mailKey]
		_, idSeen := seenID[id]
		if mailSeen || idSeen {
			result.SkippedDuplicate++
			continue
		}
		seenMail[mailKey] = struct{}{}
		seenID[id] = struct{}{}

		result.People = append(result.People, &Person{
			ID:                id,
			Mail:              mailAddress,
			DisplayName:       truncateRunes(field(record, "displayname"), maxFieldLength),
			JobTitle:          truncateRunes(field(record, "jobtitle"), maxFieldLength),
			Department:        truncateRunes(field(record, "department"), maxFieldLength),
			UserPrincipalName: truncateRunes(field(record, "userprincipalname"), maxFieldLength),
			Enabled:           parseEnabled(field(record, "accountenabled")),
		})
	}

	return result, nil
}

// normalizeMail is the form emails are compared in.
func normalizeMail(m string) string {
	return strings.ToLower(strings.TrimSpace(m))
}

// parseEnabled reads the accountEnabled column. Only an explicit false disables a person: an empty
// or unknown value leaves them enabled, so a formatting surprise can never deactivate anyone.
func parseEnabled(v string) bool {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "false", "0", "no", "n":
		return false
	}
	return true
}

func validMail(m string) bool {
	if m == "" || len(m) > maxMailLength || strings.ContainsAny(m, " \t<>,;") {
		return false
	}
	addr, err := mail.ParseAddress(m)
	return err == nil && addr.Address == m
}

// clean removes what databases refuse to store: invalid UTF-8 (for example a file saved as
// Windows-1252) and NUL bytes.
func clean(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, "")
	}
	return s
}

func truncateRunes(s string, maxRunes int) string {
	if utf8.RuneCountInString(s) <= maxRunes {
		return s
	}
	return string([]rune(s)[:maxRunes])
}
