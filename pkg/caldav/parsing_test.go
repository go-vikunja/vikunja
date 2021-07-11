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

package caldav

import (
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/models"
	"gopkg.in/d4l3k/messagediff.v1"
)

func TestParseTaskFromVTODO(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name      string
		args      args
		wantVTask *models.Task
		wantErr   bool
	}{
		{
			name: "normal",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
		{
			name: "With priority",
			args: args{content: `BEGIN:VCALENDAR
VERSION:2.0
METHOD:PUBLISH
X-PUBLISHED-TTL:PT4H
X-WR-CALNAME:test
PRODID:-//RandomProdID which is not random//EN
BEGIN:VTODO
UID:randomuid
DTSTAMP:20181201T011204
SUMMARY:Todo #1
DESCRIPTION:Lorem Ipsum
PRIORITY:9
LAST-MODIFIED:00010101T000000
END:VTODO
END:VCALENDAR`,
			},
			wantVTask: &models.Task{
				Title:       "Todo #1",
				UID:         "randomuid",
				Description: "Lorem Ipsum",
				Priority:    1,
				Updated:     time.Unix(1543626724, 0).In(config.GetTimeZone()),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTaskFromVTODO(tt.args.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTaskFromVTODO() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff, equal := messagediff.PrettyDiff(got, tt.wantVTask); !equal {
				t.Errorf("ParseTaskFromVTODO() gotVTask = %v, want %v, diff = %s", got, tt.wantVTask, diff)
			}
		})
	}
}
