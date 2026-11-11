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

package doctor

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"

	"github.com/fatih/color"
)

var (
	greenCheck = color.New(color.FgGreen).SprintFunc()
	redCross   = color.New(color.FgRed).SprintFunc()
	bold       = color.New(color.Bold).SprintFunc()
)

// PrintHeader writes the report header to the given writer.
func PrintHeader(w io.Writer) {
	fmt.Fprintln(w, bold("Vikunja Doctor"))
	fmt.Fprintln(w, "==============")
	fmt.Fprintln(w)
}

// PrintGroup writes a single check group to the given writer with colored output.
func PrintGroup(w io.Writer, group CheckGroup) {
	fmt.Fprintln(w, bold(group.Name))
	for _, result := range group.Results {
		printResult(w, result)
	}
	fmt.Fprintln(w)
}

func printResult(w io.Writer, result CheckResult) {
	marker := greenCheck("✓")
	value := result.Value
	if !result.Passed {
		marker = redCross("✗")
		if result.Error != "" {
			value = result.Error
		}
	}

	fmt.Fprintf(w, "  %s %s: %s\n", marker, escapeControl(result.Name), escapeControl(value))

	for _, line := range result.Lines {
		fmt.Fprintf(w, "      %s\n", escapeControl(line))
	}
}

// Values can come from the database or remote providers; keep them on one line and free of terminal escapes.
func escapeControl(s string) string {
	if strings.IndexFunc(s, unicode.IsControl) == -1 {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if unicode.IsControl(r) {
			q := strconv.QuoteRune(r)
			b.WriteString(q[1 : len(q)-1])
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// CountFailed returns the number of failed checks across all groups.
func CountFailed(groups []CheckGroup) int {
	count := 0
	for _, group := range groups {
		for _, result := range group.Results {
			if !result.Passed {
				count++
			}
		}
	}
	return count
}
