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
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

// This file holds the Jalali (Solar Hijri) recurrence math for task repeat
// modes JalaliMonth (3) and JalaliYear (4). The conversion is a JDN-based
// port of the jalaali-js algorithm, kept dependency-free on purpose.

// jalaliBreaks are the 33/29-year cycle boundaries the Jalali leap pattern restarts at.
var jalaliBreaks = []int{-61, 9, 38, 199, 426, 686, 756, 818, 1111, 1181, 1210, 1635, 2060, 2097, 2192, 2262, 2324, 2394, 2456, 3178}

// jalCal resolves a Jalali year to its Gregorian counterpart and Nowruz offset.
// leap is 0 in leap years, 1 in the year right after one; ok is false outside jalaliBreaks.
func jalCal(jy int, withoutLeap bool) (leap, gy, march int, ok bool) {
	bl := len(jalaliBreaks)
	gy = jy + 621
	leapJ := -14
	jp := jalaliBreaks[0]
	if jy < jp || jy >= jalaliBreaks[bl-1] {
		return 0, 0, 0, false
	}
	jump := 0
	for i := 1; i < bl; i++ {
		jm := jalaliBreaks[i]
		jump = jm - jp
		if jy < jm {
			break
		}
		leapJ += (jump/33)*8 + (jump%33)/4
		jp = jm
	}
	n := jy - jp
	leapJ += (n/33)*8 + ((n%33)+3)/4
	if jump%33 == 4 && jump-n == 4 {
		leapJ++
	}
	leapG := gy/4 - ((gy/100+1)*3)/4 - 150
	march = 20 + leapJ - leapG
	if withoutLeap {
		return 0, gy, march, true
	}
	if jump-n < 6 {
		n = n - jump + ((jump+4)/33)*33
	}
	leap = (((n + 1) % 33) - 1) % 4
	if leap == -1 {
		leap = 4
	}
	return leap, gy, march, true
}

func gregorianToJDN(gy, gm, gd int) int {
	d := ((gy+(gm-8)/6+100100)*1461)/4 + (153*((gm+9)%12)+2)/5 + gd - 34840408
	d -= (((gy + 100100 + (gm-8)/6) / 100) * 3) / 4
	d += 752
	return d
}

func jdnToGregorian(jdn int) (gy, gm, gd int) {
	j := 4*jdn + 139361631
	j += ((((4*jdn+183187720)/146097)*3)/4)*4 - 3908
	i := ((j%1461)/4)*5 + 308
	gd = ((i % 153) / 5) + 1
	gm = ((i / 153) % 12) + 1
	gy = j/1461 - 100100 + (8-gm)/6
	return gy, gm, gd
}

func jalaliToJDN(jy, jm, jd int) (int, bool) {
	_, gy, march, ok := jalCal(jy, true)
	if !ok {
		return 0, false
	}
	return gregorianToJDN(gy, 3, march) + (jm-1)*31 - (jm/7)*(jm-7) + jd - 1, true
}

func jdnToJalali(jdn int) (jy, jm, jd int, ok bool) {
	gy, _, _ := jdnToGregorian(jdn)
	jy = gy - 621
	leap, _, march, ok := jalCal(jy, false)
	if !ok {
		return 0, 0, 0, false
	}
	jdn1f := gregorianToJDN(gy, 3, march)
	k := jdn - jdn1f
	if k >= 0 {
		if k <= 185 {
			return jy, 1 + k/31, (k % 31) + 1, true
		}
		k -= 186
	} else {
		jy--
		k += 179
		if leap == 1 { // previous year was leap, so its Esfand had 30 days
			k++
		}
	}
	return jy, 7 + k/30, (k % 30) + 1, true
}

func gregorianToJalali(gy int, gm time.Month, gd int) (jy, jm, jd int, ok bool) {
	return jdnToJalali(gregorianToJDN(gy, int(gm), gd))
}

func jalaliToGregorian(jy, jm, jd int) (gy int, gm time.Month, gd int, ok bool) {
	jdn, ok := jalaliToJDN(jy, jm, jd)
	if !ok {
		return 0, 0, 0, false
	}
	y, m, d := jdnToGregorian(jdn)
	return y, time.Month(m), d, true
}

// jalaliMonthLength reports the days in jm of jy; Esfand is 30 in leap years, 29 otherwise.
func jalaliMonthLength(jy, jm int) int {
	switch {
	case jm <= 6:
		return 31
	case jm <= 11:
		return 30
	default:
		if leap, _, _, ok := jalCal(jy, false); ok && leap == 0 {
			return 30
		}
		return 29
	}
}

func floorDiv(a, b int) int {
	q := a / b
	if a%b != 0 && (a < 0) != (b < 0) {
		q--
	}
	return q
}

func floorMod(a, b int) int {
	return a - floorDiv(a, b)*b
}

// addJalaliMonthsToDate steps d forward by months Jalali months in loc's
// wall-clock, clamping the day to the target month instead of overflowing.
func addJalaliMonthsToDate(d time.Time, months int, loc *time.Location) time.Time {
	if loc == nil {
		loc = config.GetTimeZone()
	}
	ld := d.In(loc)
	jy, jm, jd, ok := gregorianToJalali(ld.Year(), ld.Month(), ld.Day())
	if !ok {
		return d
	}
	total := jy*12 + (jm - 1) + months
	njy := floorDiv(total, 12)
	njm := floorMod(total, 12) + 1
	if l := jalaliMonthLength(njy, njm); jd > l {
		jd = l
	}
	gy, gm, gd, ok := jalaliToGregorian(njy, njm, jd)
	if !ok {
		return d
	}
	return time.Date(gy, gm, gd, ld.Hour(), ld.Minute(), ld.Second(), ld.Nanosecond(), loc)
}

// addJalaliYearsToDate steps d forward by years Jalali years, clamping Esfand 30 to 29 in common years.
func addJalaliYearsToDate(d time.Time, years int, loc *time.Location) time.Time {
	if loc == nil {
		loc = config.GetTimeZone()
	}
	ld := d.In(loc)
	jy, jm, jd, ok := gregorianToJalali(ld.Year(), ld.Month(), ld.Day())
	if !ok {
		return d
	}
	njy := jy + years
	if l := jalaliMonthLength(njy, jm); jd > l {
		jd = l
	}
	gy, gm, gd, ok := jalaliToGregorian(njy, jm, jd)
	if !ok {
		return d
	}
	return time.Date(gy, gm, gd, ld.Hour(), ld.Minute(), ld.Second(), ld.Nanosecond(), loc)
}

// repeatLocationForDoer anchors Jalali recurrence in the acting user's timezone, falling back to the service timezone.
func repeatLocationForDoer(s *xorm.Session, a web.Auth) *time.Location {
	if s != nil && a != nil {
		if u, err := GetUserOrLinkShareUser(s, a); err == nil && u != nil && u.Timezone != "" {
			if loc, err := time.LoadLocation(u.Timezone); err == nil && loc != nil {
				return loc
			}
		}
	}
	return config.GetTimeZone()
}

func setTaskDatesJalaliMonthRepeat(oldTask, newTask *Task, loc *time.Location) {
	if loc == nil {
		loc = config.GetTimeZone()
	}
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = addJalaliMonthsToDate(oldTask.DueDate, 1, loc)
	}

	newTask.Reminders = oldTask.Reminders
	if len(oldTask.Reminders) > 0 {
		for in, r := range oldTask.Reminders {
			newTask.Reminders[in].Reminder = addJalaliMonthsToDate(r.Reminder, 1, loc)
		}
	}

	if !oldTask.StartDate.IsZero() && !oldTask.EndDate.IsZero() {
		diff := oldTask.EndDate.Sub(oldTask.StartDate)
		newTask.StartDate = addJalaliMonthsToDate(oldTask.StartDate, 1, loc)
		newTask.EndDate = newTask.StartDate.Add(diff)
	} else {
		if !oldTask.StartDate.IsZero() {
			newTask.StartDate = addJalaliMonthsToDate(oldTask.StartDate, 1, loc)
		}

		if !oldTask.EndDate.IsZero() {
			newTask.EndDate = addJalaliMonthsToDate(oldTask.EndDate, 1, loc)
		}
	}

	newTask.Done = false
}

func setTaskDatesJalaliYearRepeat(oldTask, newTask *Task, loc *time.Location) {
	if loc == nil {
		loc = config.GetTimeZone()
	}
	if !oldTask.DueDate.IsZero() {
		newTask.DueDate = addJalaliYearsToDate(oldTask.DueDate, 1, loc)
	}

	newTask.Reminders = oldTask.Reminders
	if len(oldTask.Reminders) > 0 {
		for in, r := range oldTask.Reminders {
			newTask.Reminders[in].Reminder = addJalaliYearsToDate(r.Reminder, 1, loc)
		}
	}

	if !oldTask.StartDate.IsZero() && !oldTask.EndDate.IsZero() {
		diff := oldTask.EndDate.Sub(oldTask.StartDate)
		newTask.StartDate = addJalaliYearsToDate(oldTask.StartDate, 1, loc)
		newTask.EndDate = newTask.StartDate.Add(diff)
	} else {
		if !oldTask.StartDate.IsZero() {
			newTask.StartDate = addJalaliYearsToDate(oldTask.StartDate, 1, loc)
		}

		if !oldTask.EndDate.IsZero() {
			newTask.EndDate = addJalaliYearsToDate(oldTask.EndDate, 1, loc)
		}
	}

	newTask.Done = false
}
