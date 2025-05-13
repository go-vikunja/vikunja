// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package utils

import (
	"fmt"
	"math"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/i18n"
)

// HumanizeDuration formats a time.Duration in a human-friendly format.
// Based on https://gist.github.com/harshavardhana/327e0577c4fed9211f65
func HumanizeDuration(duration time.Duration, lang string) string {
	years := int64(duration.Hours() / 24 / 365)
	days := int64(duration.Hours()/24) - years*365
	weeks := days / 7
	days -= weeks * 7

	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))

	chunks := []struct {
		pluralFormatKey string
		oneKey          string
		amount          int64
	}{
		{"time.many_years", "time.one_year", years},
		{"time.many_weeks", "time.one_week", weeks},
		{"time.many_days", "time.one_day", days},
		{"time.many_hours", "time.one_hour", hours},
		{"time.many_minutes", "time.one_minute", minutes},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, i18n.T(lang, chunk.oneKey))
		default:
			parts = append(parts, fmt.Sprintf(i18n.T(lang, chunk.pluralFormatKey), chunk.amount))
		}
	}

	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ", ") + " " + i18n.T(lang, "time.list_last_separator") + " " + parts[len(parts)-1]
	}

	return strings.Join(parts, ", ")
}
