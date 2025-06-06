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

package utils

import (
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
		key    string
		amount int64
	}{
		{"time.since_years", years},
		{"time.since_weeks", weeks},
		{"time.since_days", days},
		{"time.since_hours", hours},
		{"time.since_minutes", minutes},
	}

	parts := []string{}

	for _, chunk := range chunks {
		if chunk.amount <= 0 {
			continue
		}

		translatedText := i18n.TP(lang, chunk.key, chunk.amount, chunk.amount)
		parts = append(parts, translatedText)
	}

	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ", ") + " " + i18n.T(lang, "time.list_last_separator") + " " + parts[len(parts)-1]
	}

	return strings.Join(parts, ", ")
}
