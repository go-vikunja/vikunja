/**
 * Helper functions for working with structured task repeat objects.
 * The API provides structured repeat data (TaskRepeat) instead of raw RRULE strings.
 */

import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

import type {TaskRepeat} from '@/client/generated'
import {parseDateOrNull} from '@/helpers/parseDateOrNull'

dayjs.extend(utc)
dayjs.extend(timezone)

export const REPEAT_FREQUENCIES = {
	Hours: 'hours',
	Days: 'days',
	Weeks: 'weeks',
	Months: 'months',
	Years: 'years',
} as const

export type RepeatFrequency = typeof REPEAT_FREQUENCIES[keyof typeof REPEAT_FREQUENCIES]

const UI_TO_FREQ: Record<RepeatFrequency, string> = {
	hours: 'hourly',
	days: 'daily',
	weeks: 'weekly',
	months: 'monthly',
	years: 'yearly',
}

function uiFreqToFreq(freq: RepeatFrequency): string {
	return UI_TO_FREQ[freq] || 'daily'
}

/**
 * Creates a structured repeat object from UI-friendly settings.
 */
export function repeatFromSettings(amount: number, freq: RepeatFrequency, bymonthday?: number): TaskRepeat | null {
	if (amount <= 0) {
		return null
	}

	const repeat: TaskRepeat = {
		freq: uiFreqToFreq(freq),
		interval: amount,
	}

	if (bymonthday !== undefined && bymonthday > 0) {
		repeat.by_month_day = [bymonthday]
	}

	return repeat
}

/**
 * Checks if a task has a valid repeat configuration.
 */
export function isRepeating(repeat: TaskRepeat | null | undefined): boolean {
	return repeat != null && !!repeat.freq
}

// Repeat editor model. Components go through these helpers and never read TaskRepeat fields directly.

type RepeatClass = 'none' | 'editable' | 'foreign' | 'locked'
export type RepeatUnit = 'hour' | 'day' | 'week' | 'month' | 'year'
type RepeatMonthMode = 'day' | 'nthWeekday' | 'lastDay'
type RepeatEnd = 'never' | 'count' | 'until'
export type WeekdayCode = 'mo' | 'tu' | 'we' | 'th' | 'fr' | 'sa' | 'su'
export type RepeatOrdinal = 1 | 2 | 3 | 4 | -1

export interface RepeatForm {
	unit: RepeatUnit
	interval: number
	/** `null` = implicit: the due date's weekday (no by_day stored). */
	weekdays: WeekdayCode[] | null
	monthMode: RepeatMonthMode
	/** `null` = implicit: the due date's day (no by_month_day stored). */
	monthDay: number | null
	ordinal: RepeatOrdinal
	monthWeekday: WeekdayCode
	end: RepeatEnd
	count: number
	/** 'YYYY-MM-DD', a calendar date in the recurrence zone. */
	untilDate: string | null
	/** The stored RFC 3339 value, kept verbatim while untilDate is unchanged. */
	untilInstant: string | null
}

type RepeatPresetId =
	| 'none'
	| 'daily'
	| 'weekdays'
	| 'weeklyOnDay'
	| 'monthlyOnDate'
	| 'monthlyOnNthWeekday'
	| 'yearlyOnDate'

export interface RepeatPreset {
	id: RepeatPresetId
	labelKey: string
	labelParams: Record<string, unknown>
	repeat: TaskRepeat | null
}

interface ZonedDay {
	year: number
	month: number
	day: number
	weekday: WeekdayCode
	ordinal: RepeatOrdinal
}

interface RepeatFormError {
	field: keyof RepeatForm
	key: string
	params?: Record<string, unknown>
}

const WEEKDAY_CODES: readonly WeekdayCode[] = ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su']
const DAYJS_DAY_TO_WEEKDAY: readonly WeekdayCode[] = ['su', 'mo', 'tu', 'we', 'th', 'fr', 'sa']

const FREQ_TO_UNIT: Record<string, RepeatUnit> = {
	hourly: 'hour',
	daily: 'day',
	weekly: 'week',
	monthly: 'month',
	yearly: 'year',
}

const UNIT_TO_FREQ: Record<RepeatUnit, string> = {
	hour: 'hourly',
	day: 'daily',
	week: 'weekly',
	month: 'monthly',
	year: 'yearly',
}

// The server refuses an INTERVAL spanning more than 100 years and a COUNT above 10,000.
const MAX_INTERVAL: Record<RepeatUnit, number> = {
	hour: 876_600,
	day: 36_525,
	week: 5_218,
	month: 1_200,
	year: 100,
}
const MAX_COUNT = 10_000

// read_only is response-only metadata, not part of the rule.
const RESPONSE_ONLY_FIELD = 'read_only'

const REPEAT_FIELDS: ReadonlySet<string> = new Set<keyof TaskRepeat>([
	'freq', 'interval', 'by_day', 'by_month_day', 'by_month', 'by_year_day', 'by_week_no',
	'by_set_pos', 'by_hour', 'by_minute', 'by_second', 'count', 'until', 'wkst', RESPONSE_ONLY_FIELD,
])

const UNSUPPORTED_LIST_FIELDS = ['by_month', 'by_year_day', 'by_week_no', 'by_set_pos', 'by_hour', 'by_minute', 'by_second'] as const

const RFC3339_INSTANT = /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})(?:\.\d+)?(?:Z|[+-](\d{2}):(\d{2}))$/i
const DAYS_IN_MONTH = [31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31]
const PLAIN_WEEKDAY = /^(mo|tu|we|th|fr|sa|su)$/
const NTH_WEEKDAY = /^(-1|[1-4])(mo|tu|we|th|fr|sa|su)$/

function hasValue(value: unknown): boolean {
	return value !== null && value !== undefined && !(Array.isArray(value) && value.length === 0)
}

function isIntegerInRange(value: unknown, min: number, max = Number.MAX_SAFE_INTEGER): boolean {
	return Number.isInteger(value) && (value as number) >= min && (value as number) <= max
}

function isLeapYear(year: number): boolean {
	return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0)
}

// Date.parse rolls impossible components over (Feb 30 -> Mar 2), so an UNTIL it accepts may not be the stored one.
export function isStrictInstant(value: unknown): boolean {
	const match = typeof value === 'string' ? RFC3339_INSTANT.exec(value) : null
	if (!match) {
		return false
	}
	const [year, month, day, hour, minute, second] = match.slice(1, 7).map(Number) as [number, number, number, number, number, number]
	const offsetHour = Number(match[7] ?? 0)
	const offsetMinute = Number(match[8] ?? 0)
	if (month < 1 || month > 12) {
		return false
	}
	const daysInMonth = month === 2 && isLeapYear(year) ? 29 : DAYS_IN_MONTH[month - 1]!
	return day >= 1 && day <= daysInMonth &&
		hour <= 23 && minute <= 59 && second <= 59 &&
		offsetHour <= 23 && offsetMinute <= 59
}

function weekdayRank(code: unknown): number {
	const rank = WEEKDAY_CODES.indexOf(code as WeekdayCode)
	return rank === -1 ? WEEKDAY_CODES.length : rank
}

function sortedWeekdays(codes: readonly string[]): WeekdayCode[] {
	return WEEKDAY_CODES.filter(code => codes.includes(code))
}

function lowerCaseAll(values: unknown[]): unknown[] {
	return values.map(value => typeof value === 'string' ? value.toLowerCase() : value)
}

function canonicalByDay(byDay: unknown, freq: unknown): unknown {
	if (!Array.isArray(byDay)) {
		return byDay
	}
	const codes = lowerCaseAll(byDay)
	return freq === 'weekly' ? codes.sort((a, b) => weekdayRank(a) - weekdayRank(b)) : codes
}

// Unknown keys are kept so that they make a repeat foreign instead of silently disappearing.
export function canonicalRepeat(r: TaskRepeat): Record<string, unknown> {
	const freq = typeof r.freq === 'string' ? r.freq.toLowerCase() : r.freq
	const canonical: Record<string, unknown> = {interval: 1, wkst: 'mo'}
	for (const [key, value] of Object.entries(r)) {
		if (!hasValue(value) || key === RESPONSE_ONLY_FIELD) {
			continue
		}
		switch (key) {
			case 'freq':
				canonical.freq = freq
				break
			case 'interval':
				canonical.interval = value === 0 ? 1 : value
				break
			case 'by_day':
				canonical.by_day = canonicalByDay(value, freq)
				break
			case 'until':
				canonical.until = typeof value === 'string' ? Date.parse(value) : value
				break
			default:
				canonical[key] = Array.isArray(value) ? [...value] : value
		}
	}
	return canonical
}

function canonicalEqual(a: unknown, b: unknown): boolean {
	if (Array.isArray(a) || Array.isArray(b)) {
		return Array.isArray(a) && Array.isArray(b) && a.length === b.length && a.every((value, i) => canonicalEqual(value, b[i]))
	}
	if (a !== null && b !== null && typeof a === 'object' && typeof b === 'object') {
		const objectA = a as Record<string, unknown>
		const objectB = b as Record<string, unknown>
		const keys = Object.keys(objectA).sort()
		return canonicalEqual(keys, Object.keys(objectB).sort()) && keys.every(key => canonicalEqual(objectA[key], objectB[key]))
	}
	return Object.is(a, b)
}

function hasEditableScalars(r: TaskRepeat & {freq: string}): boolean {
	if (Object.keys(r).some(key => !REPEAT_FIELDS.has(key))) {
		return false
	}
	if (!(r.freq.toLowerCase() in FREQ_TO_UNIT)) {
		return false
	}
	if (hasValue(r.interval) && !isIntegerInRange(r.interval, 1)) {
		return false
	}
	if (hasValue(r.wkst) && r.wkst !== 'mo') {
		return false
	}
	if (hasValue(r.count) && !isIntegerInRange(r.count, 1)) {
		return false
	}
	if (hasValue(r.until) && !isStrictInstant(r.until)) {
		return false
	}
	return !(hasValue(r.count) && hasValue(r.until)) && !UNSUPPORTED_LIST_FIELDS.some(field => hasValue(r[field]))
}

function hasEditableSelectors(freq: string, byDay: unknown, byMonthDay: unknown): boolean {
	const days = hasValue(byDay) ? byDay : null
	const monthDays = hasValue(byMonthDay) ? byMonthDay : null
	if ((days !== null && !Array.isArray(days)) || (monthDays !== null && !Array.isArray(monthDays))) {
		return false
	}
	if (freq === 'weekly') {
		if (monthDays !== null) {
			return false
		}
		if (days === null) {
			return true
		}
		const codes = lowerCaseAll(days)
		return codes.every(code => typeof code === 'string' && PLAIN_WEEKDAY.test(code)) && new Set(codes).size === codes.length
	}
	if (freq === 'monthly') {
		if (days !== null && monthDays !== null) {
			return false
		}
		if (days !== null) {
			const [code] = lowerCaseAll(days)
			return days.length === 1 && typeof code === 'string' && NTH_WEEKDAY.test(code)
		}
		if (monthDays !== null) {
			return monthDays.length === 1 && (monthDays[0] === -1 || isIntegerInRange(monthDays[0], 1, 31))
		}
		return true
	}
	return days === null && monthDays === null
}

function isStructurallyEditable(r: TaskRepeat): boolean {
	return typeof r.freq === 'string' &&
		hasEditableScalars(r as TaskRepeat & {freq: string}) &&
		hasEditableSelectors(r.freq.toLowerCase(), r.by_day, r.by_month_day)
}

function ordinalOfDay(day: number): RepeatOrdinal {
	const ordinal = Math.ceil(day / 7)
	// The 5th such weekday of a month is always its last one.
	return ordinal >= 5 ? -1 : ordinal as RepeatOrdinal
}

function browserZone(): string {
	return Intl.DateTimeFormat().resolvedOptions().timeZone
}

function resolveZone(zone: string): string {
	return zone || browserZone()
}

function isLoadableZone(zone: string): boolean {
	try {
		Intl.DateTimeFormat('en-US', {timeZone: zone})
		return true
	} catch (_e) {
		return false
	}
}

function formatYmd({year, month, day}: {year: number, month: number, day: number}): string {
	return `${String(year).padStart(4, '0')}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
}

interface ZoneSettings {
	timezone?: string | null
}

// An unloadable setting counts as empty everywhere, so the preview, its hint and the fallback zone agree.
function usableZoneSetting(settings: ZoneSettings): string | null {
	const zone = settings.timezone
	return zone && isLoadableZone(zone) ? zone : null
}

// The scheduler evaluates repeats in the user's timezone setting, so every calendar day the editor derives uses it too.
export function recurrenceZone(settings: ZoneSettings): string {
	return usableZoneSetting(settings) ?? browserZone()
}

// Without a setting the scheduler's zone is unknown, so zone-dependent previews stay hidden.
export function previewZone(settings: ZoneSettings): string | null {
	return usableZoneSetting(settings)
}

export function dueDateOf(task: {due_date?: string | null} | null | undefined): Date | null {
	return parseDateOrNull(task?.due_date)
}

export function dayInZone(date: Date, zone: string): ZonedDay {
	const zoned = dayjs(date).tz(resolveZone(zone))
	const day = zoned.date()
	return {
		year: zoned.year(),
		month: zoned.month() + 1,
		day,
		weekday: DAYJS_DAY_TO_WEEKDAY[zoned.day()]!,
		ordinal: ordinalOfDay(day),
	}
}

// Local noon, so the picker shows the same calendar date in every browser zone.
export function untilDateToPickerDate(untilDate: string | null): Date | null {
	if (!untilDate) {
		return null
	}
	const [year, month, day] = untilDate.split('-').map(Number) as [number, number, number]
	return new Date(year, month - 1, day, 12)
}

export function pickerDateToUntilDate(date: Date): string {
	return formatYmd({year: date.getFullYear(), month: date.getMonth() + 1, day: date.getDate()})
}

function storedUntil(untilDate: string, untilInstant: string | null, zone: string): string {
	if (untilInstant && formatYmd(dayInZone(new Date(untilInstant), zone)) === untilDate) {
		return untilInstant
	}
	// The last second of the picked date, in the zone the rule is evaluated in, keeps that date inclusive.
	return dayjs.tz(`${untilDate} 23:59:59`, resolveZone(zone)).toISOString()
}

function applyMonthSelector(form: RepeatForm, byDay: string[], byMonthDay: number[]): void {
	if (byMonthDay[0] === -1) {
		form.monthMode = 'lastDay'
		return
	}
	if (byMonthDay.length > 0) {
		form.monthDay = byMonthDay[0]!
		return
	}
	const nth = byDay.length > 0 ? NTH_WEEKDAY.exec(byDay[0]!) : null
	if (nth) {
		form.monthMode = 'nthWeekday'
		form.ordinal = Number(nth[1]) as RepeatOrdinal
		form.monthWeekday = nth[2] as WeekdayCode
	}
}

function mapRepeatToForm(r: TaskRepeat, dueDate: Date | null, zone: string): RepeatForm {
	const anchorDay = dayInZone(dueDate ?? new Date(), zone)
	const freq = (r.freq ?? '').toLowerCase()
	const byDay = (r.by_day ?? []).map(code => code.toLowerCase())
	const form: RepeatForm = {
		unit: FREQ_TO_UNIT[freq]!,
		interval: r.interval !== undefined && r.interval >= 1 ? r.interval : 1,
		weekdays: freq === 'weekly' && byDay.length > 0 ? sortedWeekdays(byDay) : null,
		monthMode: 'day',
		monthDay: null,
		ordinal: anchorDay.ordinal,
		monthWeekday: anchorDay.weekday,
		end: 'never',
		count: 1,
		untilDate: null,
		untilInstant: null,
	}
	if (freq === 'monthly') {
		applyMonthSelector(form, byDay, r.by_month_day ?? [])
	}
	if (r.count) {
		form.end = 'count'
		form.count = r.count
	} else if (r.until) {
		form.end = 'until'
		form.untilInstant = r.until
		form.untilDate = formatYmd(dayInZone(new Date(r.until), zone))
	}
	return form
}

function monthSelector(form: RepeatForm): Pick<TaskRepeat, 'by_day' | 'by_month_day'> {
	switch (form.monthMode) {
		case 'lastDay':
			return {by_month_day: [-1]}
		case 'nthWeekday':
			return {by_day: [`${form.ordinal}${form.monthWeekday}`]}
		default:
			return form.monthDay === null ? {} : {by_month_day: [form.monthDay]}
	}
}

function mapFormToRepeat(form: RepeatForm, zone: string): TaskRepeat {
	const repeat: TaskRepeat = {freq: UNIT_TO_FREQ[form.unit]}
	// interval is omitted when it is 1: the backend defaults an absent interval to 1 and GET always returns it.
	if (form.interval > 1) {
		repeat.interval = form.interval
	}
	if (form.unit === 'week' && form.weekdays !== null && form.weekdays.length > 0) {
		repeat.by_day = sortedWeekdays(form.weekdays)
	}
	if (form.unit === 'month') {
		Object.assign(repeat, monthSelector(form))
	}
	if (form.end === 'count' && form.count >= 1) {
		repeat.count = form.count
	}
	if (form.end === 'until' && form.untilDate) {
		repeat.until = storedUntil(form.untilDate, form.untilInstant, zone)
	}
	return repeat
}

// Fails closed: a repeat the editor does not recognise, or cannot reproduce exactly, is foreign. A rule the server
// marks read-only only accepts an unchanged echo or removal, and its projection may omit parts, so it is locked.
export function classifyRepeat(r: TaskRepeat | null | undefined, dueDate: Date | null): RepeatClass {
	if (!r) {
		return 'none'
	}
	if (r.read_only === true) {
		return 'locked'
	}
	if (!r.freq) {
		return 'none'
	}
	if (!isStructurallyEditable(r)) {
		return 'foreign'
	}
	// The zone only affects newly picked end dates; untouched instants are written verbatim.
	const roundTripped = mapFormToRepeat(mapRepeatToForm(r, dueDate, 'UTC'), 'UTC')
	return canonicalEqual(canonicalRepeat(roundTripped), canonicalRepeat(r)) ? 'editable' : 'foreign'
}

// dueDate only supplies defaults for fields the repeat does not use; it never changes a stored field.
export function repeatToForm(r: TaskRepeat, dueDate: Date | null, zone: string): RepeatForm {
	if (import.meta.env.DEV && classifyRepeat(r, dueDate) !== 'editable') {
		throw new Error('repeatToForm: the editor cannot represent this repeat')
	}
	return mapRepeatToForm(r, dueDate, zone)
}

export function formToRepeat(form: RepeatForm, dueDate: Date | null, zone: string): TaskRepeat {
	return mapFormToRepeat(form, zone)
}

// An emptied weekday selection falls back to the due date's weekday rather than being an error.
export function normalizeRepeatForm(form: RepeatForm): RepeatForm {
	return {
		...form,
		weekdays: form.weekdays !== null && form.weekdays.length > 0 ? [...form.weekdays] : null,
	}
}

export function defaultCustomForm(anchor: Date, zone: string): RepeatForm {
	const {ordinal, weekday} = dayInZone(anchor, zone)
	return {
		unit: 'week',
		interval: 1,
		weekdays: null,
		monthMode: 'day',
		monthDay: null,
		ordinal,
		monthWeekday: weekday,
		end: 'never',
		count: 1,
		untilDate: null,
		untilInstant: null,
	}
}

export function repeatPresets(anchor: Date, zone: string): RepeatPreset[] {
	const {month, day, weekday, ordinal} = dayInZone(anchor, zone)
	return [
		{id: 'none', labelKey: 'task.repeat.preset.none', labelParams: {}, repeat: null},
		{id: 'daily', labelKey: 'task.repeat.preset.daily', labelParams: {}, repeat: {freq: 'daily'}},
		{
			id: 'weekdays',
			labelKey: 'task.repeat.preset.weekdays',
			labelParams: {},
			repeat: {freq: 'weekly', by_day: ['mo', 'tu', 'we', 'th', 'fr']},
		},
		{id: 'weeklyOnDay', labelKey: 'task.repeat.preset.weeklyOnDay', labelParams: {weekday}, repeat: {freq: 'weekly'}},
		{id: 'monthlyOnDate', labelKey: 'task.repeat.preset.monthlyOnDate', labelParams: {day}, repeat: {freq: 'monthly'}},
		{
			id: 'monthlyOnNthWeekday',
			labelKey: ordinal === -1 ? 'task.repeat.preset.monthlyOnLastWeekday' : 'task.repeat.preset.monthlyOnNthWeekday',
			labelParams: {ordinal, weekday},
			repeat: {freq: 'monthly', by_day: [`${ordinal}${weekday}`]},
		},
		{id: 'yearlyOnDate', labelKey: 'task.repeat.preset.yearlyOnDate', labelParams: {month, day}, repeat: {freq: 'yearly'}},
	]
}

// Presets depend on the anchor, so a shape match alone would mark e.g. "2nd Tuesday" as "4th Friday".
export function presetIdFor(r: TaskRepeat | null | undefined, anchor: Date, zone: string): RepeatPresetId | null {
	if (!r || !r.freq) {
		return 'none'
	}
	const stored = canonicalRepeat(r)
	const match = repeatPresets(anchor, zone)
		.find(preset => preset.repeat !== null && canonicalEqual(canonicalRepeat(preset.repeat), stored))
	return match?.id ?? null
}

// The server keeps the stored COUNT when only a higher one arrives, reading it as a stale echo (isStaleCount in
// pkg/models/task_repeat_update.go), so such an edit would silently not save.
function isCountOnlyRaise(form: RepeatForm, stored: TaskRepeat, zone: string): boolean {
	const storedCount = stored.count
	if (!isIntegerInRange(storedCount, 1) || form.end !== 'count' || form.count <= storedCount!) {
		return false
	}
	const echo = mapFormToRepeat({...form, count: storedCount!}, zone)
	return canonicalEqual(canonicalRepeat(echo), canonicalRepeat(stored))
}

export function validateRepeatForm(form: RepeatForm, dueDate: Date | null, zone: string, stored?: TaskRepeat | null): RepeatFormError[] {
	const errors: RepeatFormError[] = []
	if (!isIntegerInRange(form.interval, 1)) {
		errors.push({field: 'interval', key: 'task.repeat.error.interval'})
	} else if (form.interval > MAX_INTERVAL[form.unit]) {
		errors.push({field: 'interval', key: 'task.repeat.error.intervalMax', params: {max: MAX_INTERVAL[form.unit]}})
	}
	if (form.end === 'count') {
		if (!isIntegerInRange(form.count, 1)) {
			errors.push({field: 'count', key: 'task.repeat.error.count'})
		} else if (form.count > MAX_COUNT) {
			errors.push({field: 'count', key: 'task.repeat.error.countMax', params: {max: MAX_COUNT}})
		} else if (stored && isCountOnlyRaise(form, stored, zone)) {
			errors.push({field: 'count', key: 'task.repeat.error.countRaise'})
		}
	}
	if (form.end === 'until') {
		if (!form.untilDate) {
			errors.push({field: 'untilDate', key: 'task.repeat.error.untilMissing'})
		} else if (dueDate !== null && form.untilDate < formatYmd(dayInZone(dueDate, zone))) {
			errors.push({field: 'untilDate', key: 'task.repeat.error.untilBeforeDue'})
		}
	}
	if (form.unit === 'month' && form.monthMode === 'day' && form.monthDay !== null && !isIntegerInRange(form.monthDay, 1, 31)) {
		errors.push({field: 'monthDay', key: 'task.repeat.error.monthDay'})
	}
	return errors
}
