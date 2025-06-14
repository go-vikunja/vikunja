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

// In caldav, priority values are an int from 0 to 9 where 1 is the highest priority and 9 the lowest. 0 is "unset".
// Vikunja only has priorites from 0 to 5 where 0 is unset and 5 is the highest
// See https://icalendar.org/iCalendar-RFC-5545/3-8-1-9-priority.html
func mapPriorityToCaldav(priority int64) (caldavPriority int) {
	switch priority {
	case 0:
		return 0
	case 1: // Low
		return 9
	case 2: // Medium
		return 5
	case 3: // High
		return 3
	case 4: // Urgent
		return 2
	case 5: // DO NOW
		return 1
	}
	return 0
}

// See mapPriorityToCaldav
func parseVTODOPriority(priority int64) (vikunjaPriority int64) {
	switch priority {
	case 0:
		return 0
	case 1:
		return 5
	case 2:
		return 4
	case 3:
		return 3
	case 4:
		return 3
	case 5:
		return 2
	case 6:
		return 1
	case 7:
		return 1
	case 8:
		return 1
	case 9:
		return 1
	}

	return 0
}
