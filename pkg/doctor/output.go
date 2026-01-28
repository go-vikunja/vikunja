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

	"github.com/fatih/color"
)

var (
	greenCheck = color.New(color.FgGreen).SprintFunc()
	redCross   = color.New(color.FgRed).SprintFunc()
	bold       = color.New(color.Bold).SprintFunc()
)

// PrintResults writes all check groups to the given writer with colored output.
func PrintResults(w io.Writer, groups []CheckGroup) {
	fmt.Fprintln(w, bold("Vikunja Doctor"))
	fmt.Fprintln(w, "==============")
	fmt.Fprintln(w)

	for _, group := range groups {
		fmt.Fprintln(w, bold(group.Name))
		for _, result := range group.Results {
			printResult(w, result)
		}
		fmt.Fprintln(w)
	}
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

	fmt.Fprintf(w, "  %s %s: %s\n", marker, result.Name, value)

	for _, line := range result.Lines {
		fmt.Fprintf(w, "      %s\n", line)
	}
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
