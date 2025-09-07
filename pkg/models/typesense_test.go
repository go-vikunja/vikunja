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
	"testing"
)

func TestConvertTaskDueDateSentinel(t *testing.T) {
	task := &Task{}
	tt := convertTaskToTypesenseTask(task, nil, nil)
	if tt.DueDate == nil || *tt.DueDate != dueDateSentinel {
		t.Fatalf("expected sentinel %d, got %v", dueDateSentinel, tt.DueDate)
	}
}

func TestConvertParsedFilterIncludeNulls(t *testing.T) {
	f := &taskFilter{field: "due_date", value: int64(42), comparator: taskFilterComparatorGreateEquals}
	out, err := convertParsedFilterToTypesense([]*taskFilter{f}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := fmt.Sprintf("(due_date:%d || due_date:>=42)", dueDateSentinel)
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}

	outNo, err := convertParsedFilterToTypesense([]*taskFilter{f}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outNo != "due_date:>=42" {
		t.Fatalf("unexpected filter %q", outNo)
	}
}

func TestConvertParsedFilterIncludeNullsWithOr(t *testing.T) {
	f1 := &taskFilter{field: "due_date", value: int64(1), comparator: taskFilterComparatorEquals, join: filterConcatOr}
	f2 := &taskFilter{field: "priority", value: int64(2), comparator: taskFilterComparatorEquals}
	out, err := convertParsedFilterToTypesense([]*taskFilter{f1, f2}, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := fmt.Sprintf("(due_date:%d || due_date:=1) || priority:=2", dueDateSentinel)
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}
