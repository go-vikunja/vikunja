import type {TaskRepeat} from '@/client/generated'
import {canonicalRepeat, classifyRepeat, repeatToForm, validateRepeatForm} from '@/helpers/rrule'

type PreviewHiddenReason = 'notEditable' | 'hourly' | 'fromCompletion' | 'noDueDate' | 'overdue' | 'noZone'

interface RepeatPreview {
	dates: Date[] | null
	hiddenReason: PreviewHiddenReason | null
}

type Freq = 'hourly' | 'daily' | 'weekly' | 'monthly' | 'yearly'

interface Rule {
	freq: Freq
	interval: number
	byDay: string[]
	byMonthDay: number[]
	count: number | null
	until: number | null
}

/**
 * A wall-clock reading in a zone written as a UTC timestamp, so it has no DST. Calendar
 * repeats are stepped on it and only become instants again through `localise`.
 */
type Floating = number

interface Occurrence {
	at: number
	/** The wall clock the occurrence stands for; for a gap-shifted instant, the unshifted one. */
	f: Floating
}

interface ZoneClock {
	float(at: number): Floating
	offset(at: number): number
	localise(f: Floating): number
}

const PREVIEW_LENGTH = 3
const SECOND_MS = 1000
const HOUR_MS = 3_600_000
const DAY_MS = 86_400_000
// No transition moves a wall clock by more than a day, so occurrences more than 48 h apart
// on the floating timeline normalise in the same order.
const LOOK_AROUND_MS = 2 * DAY_MS
// rrule-go ends a rule without UNTIL at DTSTART plus Go's largest time.Duration (about 292 years).
const MAX_DURATION_NS = BigInt('9223372036854775807')
const IMPLICIT_UNTIL_MS = Number(MAX_DURATION_NS / BigInt(1_000_000_000)) * SECOND_MS
const MAX_YEAR = 9999
const MIN_STORABLE_YEAR = 1000
const WEEKDAYS = ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su']
const NTH_WEEKDAY = /^(-?\d)(mo|tu|we|th|fr|sa|su)$/

// dayjs.tz builds a new formatter per call (~100 µs), too slow for a live preview.
const clocks = new Map<string, ZoneClock | null>()

function clockFor(zone: string | null): ZoneClock | null {
	if (!zone) {
		return null
	}
	if (!clocks.has(zone)) {
		let clock: ZoneClock | null
		try {
			clock = zoneClock(new Intl.DateTimeFormat('en-US', {
				timeZone: zone,
				hourCycle: 'h23',
				year: 'numeric',
				month: 'numeric',
				day: 'numeric',
				hour: 'numeric',
				minute: 'numeric',
				second: 'numeric',
			}))
		} catch (_e) {
			clock = null
		}
		clocks.set(zone, clock)
	}
	return clocks.get(zone) ?? null
}

function zoneClock(formatter: Intl.DateTimeFormat): ZoneClock {
	function float(at: number): Floating {
		const parts: Record<string, number> = {}
		for (const {type, value} of formatter.formatToParts(at)) {
			parts[type] = Number(value)
		}
		const wholeSecond = Math.floor(at / SECOND_MS) * SECOND_MS
		const date = new Date(0)
		date.setUTCFullYear(parts.year, parts.month - 1, parts.day)
		date.setUTCHours(parts.hour % 24, parts.minute, parts.second, at - wholeSecond)
		return date.getTime()
	}
	function offset(at: number): number {
		return float(at) - at
	}
	// An overlap takes the earlier instant; a gap moves forward by its length (the
	// pre-transition offset). Go's time.Date resolves both differently and is not mirrored.
	function localise(f: Floating): number {
		const c1 = f - offset(f - LOOK_AROUND_MS)
		const c2 = f - offset(f + LOOK_AROUND_MS)
		const valid1 = float(c1) === f
		const valid2 = float(c2) === f
		if (valid1 && valid2) {
			return Math.min(c1, c2)
		}
		return valid2 && !valid1 ? c2 : c1
	}
	return {float, offset, localise}
}

function browserZone(): string {
	return Intl.DateTimeFormat().resolvedOptions().timeZone
}

function dayOf(f: Floating): number {
	return Math.floor(f / DAY_MS)
}

function timeOfDay(f: Floating): number {
	return f - dayOf(f) * DAY_MS
}

function truncateToSecond(ms: number): number {
	return Math.floor(ms / SECOND_MS) * SECOND_MS
}

function calendarOf(day: number): {year: number, month: number, date: number, weekday: number} {
	const d = new Date(day * DAY_MS)
	return {
		year: d.getUTCFullYear(),
		month: d.getUTCMonth() + 1,
		date: d.getUTCDate(),
		weekday: (d.getUTCDay() + 6) % 7,
	}
}

function dayNumber(year: number, month: number, date: number): number {
	const d = new Date(0)
	d.setUTCFullYear(year, month - 1, date)
	return d.getTime() / DAY_MS
}

function daysInMonth(year: number, month: number): number {
	return new Date(dayNumber(year, month + 1, 0) * DAY_MS).getUTCDate()
}

function isStorable(at: number): boolean {
	const year = new Date(at).getUTCFullYear()
	return year >= MIN_STORABLE_YEAR && year <= MAX_YEAR
}

function ruleOf(repeat: TaskRepeat): Rule {
	const canonical = canonicalRepeat(repeat)
	const until = canonical.until as number | undefined
	const count = canonical.count as number | undefined
	return {
		freq: canonical.freq as Freq,
		interval: canonical.interval as number,
		byDay: (canonical.by_day as string[] | undefined) ?? [],
		byMonthDay: (canonical.by_month_day as number[] | undefined) ?? [],
		count: count !== undefined && count > 0 ? count : null,
		until: until === undefined ? null : truncateToSecond(until),
	}
}

// The server refuses to advance a rule whose interval spans more than 100 years; the form names that limit.
function exceedsIntervalHorizon(repeat: TaskRepeat, dueDate: Date | null): boolean {
	return validateRepeatForm(repeatToForm(repeat, dueDate, 'UTC'), null, 'UTC')
		.some(error => error.field === 'interval')
}

function wholePeriodDays(rule: Rule): number {
	if ((rule.freq !== 'daily' && rule.freq !== 'weekly') ||
		rule.byDay.length > 0 || rule.byMonthDay.length > 0 || rule.count !== null || rule.until !== null) {
		return 0
	}
	return rule.interval * (rule.freq === 'weekly' ? 7 : 1)
}

function monthDays(rule: Rule, year: number, month: number, startDate: number): number[] {
	const lastDate = daysInMonth(year, month)
	if (rule.byDay.length > 0) {
		return rule.byDay.flatMap(code => {
			const match = NTH_WEEKDAY.exec(code)
			if (!match) {
				return []
			}
			const nth = Number(match[1])
			const weekday = WEEKDAYS.indexOf(match[2]!)
			if (nth < 0) {
				const lastWeekday = calendarOf(dayNumber(year, month, lastDate)).weekday
				return [lastDate - (lastWeekday - weekday + 7) % 7 + (nth + 1) * 7]
			}
			const firstWeekday = calendarOf(dayNumber(year, month, 1)).weekday
			return [1 + (weekday - firstWeekday + 7) % 7 + (nth - 1) * 7]
		})
	}
	const wanted = rule.byMonthDay.length > 0 ? rule.byMonthDay : [startDate]
	// Months without the day are skipped, never clamped, exactly like the scheduler.
	return wanted.map(d => d < 0 ? lastDate + d + 1 : d)
}

// Calendar days on which the rule fires, from the start's day to lastDay, counting
// intervals from the start's period (weeks begin on Monday, as WKST=MO). Periods wholly
// before fromDay are skipped without changing the phase.
function* candidateDays(rule: Rule, startDay: number, fromDay: number, lastDay: number): Generator<number> {
	const start = calendarOf(startDay)
	const periodsBefore = (gap: number, length: number) => Math.max(0, Math.floor(gap / length))
	if (rule.freq === 'daily') {
		const first = startDay + Math.max(0, Math.ceil((fromDay - startDay) / rule.interval)) * rule.interval
		for (let day = first; day <= lastDay; day += rule.interval) {
			yield day
		}
		return
	}
	if (rule.freq === 'weekly') {
		const weekdays = rule.byDay.length > 0
			? rule.byDay.map(code => WEEKDAYS.indexOf(code)).sort((a, b) => a - b)
			: [start.weekday]
		const step = 7 * rule.interval
		const firstMonday = startDay - start.weekday
		for (let monday = firstMonday + periodsBefore(fromDay - firstMonday, step) * step; monday <= lastDay; monday += step) {
			for (const weekday of weekdays) {
				const day = monday + weekday
				if (day >= startDay && day <= lastDay) {
					yield day
				}
			}
		}
		return
	}
	const from = calendarOf(Math.max(fromDay, startDay))
	if (rule.freq === 'monthly') {
		const startIndex = start.year * 12 + start.month - 1
		const skip = periodsBefore(from.year * 12 + from.month - 1 - startIndex, rule.interval)
		for (let monthIndex = startIndex + skip * rule.interval; ; monthIndex += rule.interval) {
			const year = Math.floor(monthIndex / 12)
			const month = monthIndex % 12 + 1
			if (dayNumber(year, month, 1) > lastDay) {
				return
			}
			const lastDate = daysInMonth(year, month)
			const days = [...new Set(monthDays(rule, year, month, start.date))]
				.filter(date => date >= 1 && date <= lastDate)
				.map(date => dayNumber(year, month, date))
				.filter(day => day >= startDay && day <= lastDay)
				.sort((a, b) => a - b)
			yield* days
		}
	}
	if (rule.freq === 'yearly') {
		const skip = periodsBefore(from.year - start.year, rule.interval)
		for (let year = start.year + skip * rule.interval; dayNumber(year, 1, 1) <= lastDay; year += rule.interval) {
			if (start.date <= daysInMonth(year, start.month)) {
				const day = dayNumber(year, start.month, start.date)
				if (day >= startDay && day <= lastDay) {
					yield day
				}
			}
		}
	}
}

interface Search {
	rule: Rule
	clock: ZoneClock
	/** DTSTART: the anchor's wall clock, at whole seconds as rrule-go reads it. */
	dtstart: Floating
}

interface FloatingCandidates {
	values: Generator<Floating>
	withinUntil: (at: number) => boolean
}

// rrule-go iterates to UNTIL, or without one to DTSTART plus its implicit window, never past the
// year 9999. UNTIL is widened by the look-around so no occurrence that normalises inside it is cut off.
function floatingCandidates({rule, clock, dtstart}: Search, fromDay: number): FloatingCandidates {
	const limit = rule.until !== null ? clock.float(rule.until) + LOOK_AROUND_MS : dtstart + IMPLICIT_UNTIL_MS
	const lastDay = Math.min(dayOf(limit), dayNumber(MAX_YEAR, 12, 31))
	const time = timeOfDay(dtstart)
	function* values(): Generator<Floating> {
		for (const day of candidateDays(rule, dayOf(dtstart), fromDay, lastDay)) {
			const f = day * DAY_MS + time
			if (f > limit) {
				return
			}
			yield f
		}
	}
	const until = rule.until
	return {
		values: values(),
		withinUntil: at => until === null || truncateToSecond(at) <= until,
	}
}

// Of two floating values that normalise to one instant, the unshifted one stands for it.
function representative(clock: ZoneClock, kept: Occurrence, other: Occurrence): Occurrence {
	return clock.float(kept.at) !== kept.f && clock.float(other.at) === other.f ? other : kept
}

// The earliest normalised occurrence strictly after `from` (and within UNTIL). The scan stops
// once the floating values pass the best candidate's wall clock by the look-around.
function nextIterated(search: Search, from: number): Occurrence | null {
	const {clock} = search
	const skip = clock.float(from) - LOOK_AROUND_MS
	const {values, withinUntil} = floatingCandidates(search, dayOf(skip))
	let best: Occurrence | null = null
	let stop = 0
	for (const f of values) {
		if (best !== null && f > stop) {
			break
		}
		if (f <= skip) {
			continue
		}
		const at = clock.localise(f)
		if (!withinUntil(at) || at <= from) {
			continue
		}
		if (best === null || at < best.at) {
			best = {at, f}
			stop = clock.float(at) + LOOK_AROUND_MS
		} else if (at === best.at) {
			best = representative(clock, best, {at, f})
		}
	}
	return best
}

// COUNT ranks the distinct normalised instants from DTSTART, so a collision counts once. An
// instant's rank is final once the floating values have passed its wall clock by the look-around.
function nextCounted(search: Search, count: number, from: number): Occurrence | null {
	const {clock} = search
	const {values, withinUntil} = floatingCandidates(search, dayOf(search.dtstart))
	const pending: Occurrence[] = []
	let lastFinal = 0
	let rank = 0
	const settle = (f: Floating | null): Occurrence | null | undefined => {
		while (pending.length > 0 && (f === null || f > clock.float(pending[0]!.at) + LOOK_AROUND_MS)) {
			const o = pending.shift()!
			if (++rank > count) {
				return null
			}
			if (o.at > from) {
				return o
			}
			lastFinal = o.at
		}
		return undefined
	}
	for (const f of values) {
		const settled = settle(f)
		if (settled !== undefined) {
			return settled
		}
		const at = clock.localise(f)
		if (!withinUntil(at) || (rank > 0 && at <= lastFinal)) {
			continue
		}
		let i = pending.length
		while (i > 0 && pending[i - 1]!.at > at) {
			i--
		}
		if (i > 0 && pending[i - 1]!.at === at) {
			pending[i - 1] = representative(clock, pending[i - 1]!, {at, f})
		} else {
			pending.splice(i, 0, {at, f})
		}
	}
	return settle(null) ?? null
}

// Plain daily and weekly repeats take whole local periods from the anchor, at least one.
function nextWholePeriod(clock: ZoneClock, anchor: Floating, periodDays: number, from: number): Occurrence {
	const periods = Math.floor((dayOf(clock.float(from)) - dayOf(anchor)) / periodDays)
	for (let j = periods > 2 ? periods - 1 : 1; ; j++) {
		const f = anchor + j * periodDays * DAY_MS
		const at = clock.localise(f)
		if (at > from) {
			return {at, f}
		}
	}
}

// The anchor of the following completion. A gap-shifted due date is stored with its intended time
// of day, which the scheduler honours only when the gap is no longer than the shift; otherwise,
// and when nothing was shifted, it re-anchors on the stored instant's wall clock.
function nextAnchor(clock: ZoneClock, {at, f}: Occurrence): Floating {
	const stored = clock.float(at)
	const intended = truncateToSecond(timeOfDay(f))
	if (timeOfDay(stored) === timeOfDay(f)) {
		return stored
	}
	const gap = clock.offset(at) - clock.offset(at - DAY_MS)
	const displacement = (truncateToSecond(timeOfDay(stored)) - intended + DAY_MS) % DAY_MS
	return gap > 0 && displacement > 0 && displacement <= gap ? stored - displacement : stored
}

function hiddenReasonFor(
	repeat: TaskRepeat | null,
	dueDate: Date | null,
	now: Date,
	repeatsFromCurrentDate: boolean,
	zone: string | null,
): PreviewHiddenReason | null {
	if (repeat === null || classifyRepeat(repeat, dueDate) !== 'editable') {
		return 'notEditable'
	}
	if (ruleOf(repeat).freq === 'hourly') {
		return 'hourly'
	}
	if (exceedsIntervalHorizon(repeat, dueDate)) {
		return 'notEditable'
	}
	if (repeatsFromCurrentDate) {
		return 'fromCompletion'
	}
	if (dueDate === null) {
		return 'noDueDate'
	}
	if (dueDate.getTime() <= now.getTime()) {
		return 'overdue'
	}
	if (clockFor(zone) === null) {
		return 'noZone'
	}
	return null
}

/**
 * The current due date followed by the dates the scheduler assigns on the next
 * completions, each completed before it is due. `zone` is the user's timezone setting
 * only (null when unset): the scheduler's zone is then unknown, so nothing is previewed.
 *
 * The due date's own intended time is not known here: a due date that was itself moved
 * by a DST gap anchors the preview at its shifted time.
 */
export function previewDates(
	repeat: TaskRepeat | null,
	dueDate: Date | null,
	now: Date,
	repeatsFromCurrentDate: boolean,
	zone: string | null,
): RepeatPreview {
	const hiddenReason = hiddenReasonFor(repeat, dueDate, now, repeatsFromCurrentDate, zone)
	const clock = clockFor(zone)
	if (hiddenReason !== null || repeat === null || dueDate === null || clock === null) {
		return {dates: null, hiddenReason}
	}
	const rule = ruleOf(repeat)
	const periodDays = wholePeriodDays(rule)
	const dates = [new Date(dueDate.getTime())]
	// The due date is in the future, so the scheduler's anchor clamp (for anchors over a year old) never applies.
	let previous: Occurrence = {at: dueDate.getTime(), f: clock.float(dueDate.getTime())}
	let anchor = previous.f
	// COUNT is the number of occurrences left, including the current one.
	let count = rule.count
	while (dates.length < PREVIEW_LENGTH && (count === null || count > 1)) {
		let next: Occurrence | null
		if (periodDays > 0) {
			next = nextWholePeriod(clock, anchor, periodDays, previous.at)
		} else {
			const search = {rule, clock, dtstart: truncateToSecond(anchor)}
			next = count === null ? nextIterated(search, previous.at) : nextCounted(search, count, previous.at)
		}
		if (next === null || !isStorable(next.at)) {
			break
		}
		dates.push(new Date(next.at))
		anchor = nextAnchor(clock, next)
		previous = next
		count = count === null ? null : count - 1
	}
	return {dates, hiddenReason: null}
}

/**
 * The first occurrence strictly after `after`, used as the due date for a task that
 * has none. The time of day is the anchor's wall-clock time in `zone` (the browser's
 * zone when empty); the anchor also supplies the implicit weekday, day and month.
 * COUNT does not limit the first occurrence. The search ends where the scheduler's does:
 * at UNTIL, at its search window, or at the last storable year. Null then, or when the
 * editor cannot represent the repeat.
 */
export function firstOccurrenceAfter(
	repeat: TaskRepeat,
	anchor: Date,
	after: Date,
	zone: string | null,
): Date | null {
	if (classifyRepeat(repeat, anchor) !== 'editable') {
		return null
	}
	const rule = {...ruleOf(repeat), count: null}
	const afterMs = after.getTime()
	let first: number | null

	if (rule.freq === 'hourly') {
		const step = rule.interval * HOUR_MS
		const start = anchor.getTime()
		first = start > afterMs ? start : start + (Math.floor((afterMs - start) / step) + 1) * step
		if (rule.until !== null && truncateToSecond(first) > rule.until) {
			first = null
		}
	} else {
		const clock = clockFor(zone) ?? clockFor(browserZone())
		if (clock === null) {
			return null
		}
		first = nextIterated({rule, clock, dtstart: truncateToSecond(clock.float(anchor.getTime()))}, afterMs)?.at ?? null
	}
	return first !== null && isStorable(first) ? new Date(first) : null
}
