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

package caldav

import (
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func todoWithAlarms(count, titleLength int) *Todo {
	alarms := make([]Alarm, count)
	for i := range alarms {
		alarms[i] = Alarm{Time: time.Unix(int64(1_700_000_000+i*60), 0)}
	}
	return &Todo{
		UID:       "uid",
		Summary:   strings.Repeat("a", titleLength),
		Timestamp: time.Unix(1_700_000_000, 0),
		Alarms:    alarms,
	}
}

func TestParseTodosAllocatesLinearly(t *testing.T) {
	todo := todoWithAlarms(500, 1000)

	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	out := ParseTodos(&Config{Name: "test"}, []*Todo{todo})
	runtime.ReadMemStats(&after)

	assert.Equal(t, 500, strings.Count(out, "BEGIN:VALARM"))
	assert.Less(t, after.TotalAlloc-before.TotalAlloc, uint64(10*len(out)))
}

func BenchmarkParseTodosAlarms(b *testing.B) {
	for _, count := range []int{100, 200, 400} {
		todo := todoWithAlarms(count, 1000)
		b.Run(strconv.Itoa(count), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				ParseTodos(&Config{Name: "test"}, []*Todo{todo})
			}
		})
	}
}
