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

package migration

import (
	"errors"
	"regexp"
)

// Some clients (e.g. Trello's) send credentials as query parameters, and a
// transport error's message contains the full request url.
var secretQueryParamRE = regexp.MustCompile(`(?i)([?&](?:token|key|api_?key|access_token|refresh_token|client_secret|secret|password|code)=)[^&#\s"']+`)

func redactSecrets(s string) string {
	return secretQueryParamRE.ReplaceAllString(s, "${1}[REDACTED]")
}

// RedactError returns an error without credentials as is, so errors.Is/As keep working on it.
func RedactError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	redacted := redactSecrets(msg)
	if redacted == msg {
		return err
	}
	return errors.New(redacted)
}
