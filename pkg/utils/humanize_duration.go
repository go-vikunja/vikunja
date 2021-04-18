// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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
)

// HumanizeDuration formats a time.Duration in a human-friendly format.
// Based on https://gist.github.com/harshavardhana/327e0577c4fed9211f65
func HumanizeDuration(duration time.Duration) string {
	years := int64(duration.Hours() / 24 / 365)
	days := int64(duration.Hours()/24) - years*365
	weeks := days / 7
	days -= weeks * 7

	hours := int64(math.Mod(duration.Hours(), 24))
	minutes := int64(math.Mod(duration.Minutes(), 60))

	chunks := []struct {
		singularName string
		amount       int64
	}{
		{"year", years},
		{"week", weeks},
		{"day", days},
		{"hour", hours},
		{"minute", minutes},
	}

	parts := []string{}

	for _, chunk := range chunks {
		switch chunk.amount {
		case 0:
			continue
		case 1:
			parts = append(parts, fmt.Sprintf("one %s", chunk.singularName))
		default:
			parts = append(parts, fmt.Sprintf("%d %ss", chunk.amount, chunk.singularName))
		}
	}

	if len(parts) > 1 {
		return strings.Join(parts[:len(parts)-1], ", ") + " and " + parts[len(parts)-1]
	}

	return strings.Join(parts, ", ")
}
