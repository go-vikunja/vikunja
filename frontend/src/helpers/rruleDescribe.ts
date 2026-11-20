import dayjs from 'dayjs'
import utc from 'dayjs/plugin/utc'
import timezone from 'dayjs/plugin/timezone'

import type {TaskRepeat} from '@/client/generated'
import {canonicalRepeat, classifyRepeat, dayInZone, isStrictInstant, repeatToForm, type RepeatForm} from '@/helpers/rrule'

dayjs.extend(utc)
dayjs.extend(timezone)

export interface DescribeOptions {
	t: (key: string, params?: Record<string, unknown>, plural?: number) => string
	locale: string
	/** 0 = Sunday. Only changes the order weekdays are listed in. */
	weekStart: number
	/** The recurrence zone: every calendar day and end date is read in it. */
	zone: string
	dueDate?: Date | null
}

type ZonedDateStyle = 'date' | 'monthDay' | 'monthDayYear' | 'dateTime' | 'dateTimeYear'

type Translate = DescribeOptions['t']

const KEY = 'task.repeat.describe'

const WEEKDAYS = ['mo', 'tu', 'we', 'th', 'fr', 'sa', 'su'] as const
const WORKWEEK = ['mo', 'tu', 'we', 'th', 'fr']
const KNOWN_FREQS = ['secondly', 'minutely', 'hourly', 'daily', 'weekly', 'monthly', 'yearly']
const ORDINAL_CATEGORIES = ['one', 'two', 'few', 'other']
const BY_DAY_CODE = /^([+-]?\d{1,2})?(mo|tu|we|th|fr|sa|su)$/

const ZONED_DATE_STYLES: Record<ZonedDateStyle, Intl.DateTimeFormatOptions> = {
	date: {weekday: 'short', month: 'short', day: 'numeric'},
	monthDay: {month: 'short', day: 'numeric'},
	monthDayYear: {year: 'numeric', month: 'short', day: 'numeric'},
	dateTime: {month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'},
	dateTimeYear: {year: 'numeric', month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit'},
}

const SELECTORS = ['by_month', 'by_week_no', 'by_year_day', 'by_month_day', 'by_day', 'by_set_pos', 'by_hour', 'by_minute', 'by_second'] as const

type Selector = typeof SELECTORS[number]

const NUMBER_PARTS: Record<string, {key: string, param: string, min: number, max: number}> = {
	by_year_day: {key: 'byYearDay', param: 'days', min: 1, max: 366},
	by_month_day: {key: 'byMonthDay', param: 'days', min: 1, max: 31},
	by_hour: {key: 'byHour', param: 'values', min: 0, max: 23},
}

// read_only is response metadata, not a part of the rule to describe.
const KNOWN_FIELDS = new Set<string>(['freq', 'interval', ...SELECTORS, 'wkst', 'count', 'until', 'read_only'])

interface ListFormatter {
	format(items: string[]): string
}

// Intl.ListFormat is ES2021 and the vitest tsconfig project compiles without that lib, so it is typed locally.
const ListFormat = (Intl as unknown as {ListFormat: new (locale: string, options: {type: 'conjunction' | 'unit'}) => ListFormatter}).ListFormat

// Intl constructors are expensive and these run on every render of a summary.
const formatterCache = new Map<string, unknown>()

function cached<T>(key: string, create: () => T): T {
	if (!formatterCache.has(key)) {
		formatterCache.set(key, create())
	}
	return formatterCache.get(key) as T
}

function dateTimeFormat(locale: string, options: Intl.DateTimeFormatOptions): Intl.DateTimeFormat {
	return cached(`dtf|${locale}|${JSON.stringify(options)}`, () => new Intl.DateTimeFormat(locale, options))
}

function hasValue(value: unknown): boolean {
	return value !== null && value !== undefined && !(Array.isArray(value) && value.length === 0)
}

function isPositiveInteger(value: unknown): value is number {
	return Number.isInteger(value) && (value as number) >= 1
}

export function weekdayName(code: string, locale: string, style: 'long' | 'short' | 'narrow' = 'long'): string {
	const index = WEEKDAYS.indexOf(code.toLowerCase() as typeof WEEKDAYS[number])
	if (index === -1) {
		return code
	}
	// 2024-01-01 was a Monday.
	return dateTimeFormat(locale, {weekday: style, timeZone: 'UTC'}).format(Date.UTC(2024, 0, 1 + index))
}

function monthName(n: number, locale: string): string {
	if (!Number.isInteger(n) || n < 1 || n > 12) {
		return String(n)
	}
	return dateTimeFormat(locale, {month: 'long', timeZone: 'UTC'}).format(Date.UTC(2024, n - 1, 1))
}

function formatList(items: string[], locale: string): string {
	return cached(`list|${locale}`, () => new ListFormat(locale, {type: 'conjunction'})).format(items)
}

export function formatOrdinal(n: number, {t, locale}: Pick<DescribeOptions, 't' | 'locale'>): string {
	const category = cached(`ordinal|${locale}`, () => new Intl.PluralRules(locale, {type: 'ordinal'})).select(n)
	// Only these categories have keys; a locale whose ordinals need others gets its "other" form.
	const key = ORDINAL_CATEGORIES.includes(category) ? category : 'other'
	return t(`${KEY}.ordinal.${key}`, {n})
}

export function formatInZone(date: Date, zone: string, locale: string, style: ZonedDateStyle = 'date'): string {
	// An empty zone means the browser's own.
	return dateTimeFormat(locale, {...ZONED_DATE_STYLES[style], timeZone: zone || undefined}).format(date)
}

function sortedForDisplay(codes: string[], weekStart: number): string[] {
	const rank = (code: string) => (WEEKDAYS.indexOf(code as typeof WEEKDAYS[number]) + 1 - weekStart + 7) % 7
	return [...new Set(codes)].sort((a, b) => rank(a) - rank(b))
}

function weekdayList(codes: string[], opts: DescribeOptions): string {
	return formatList(sortedForDisplay(codes, opts.weekStart).map(code => weekdayName(code, opts.locale)), opts.locale)
}

function isWorkweek(codes: readonly unknown[]): boolean {
	return codes.length === WORKWEEK.length && WORKWEEK.every(code => codes.includes(code))
}

function isPlainWeekdayList(codes: unknown): codes is string[] {
	return Array.isArray(codes) && codes.every(code => WEEKDAYS.includes(code))
}

function formatMonthDay(month: number, day: number, locale: string): string | null {
	// A leap year, so February 29 exists; a day the month never has would roll over into the next month.
	const date = new Date(Date.UTC(2024, month - 1, day))
	if (date.getUTCMonth() !== month - 1) {
		return null
	}
	return dateTimeFormat(locale, {month: 'long', day: 'numeric', timeZone: 'UTC'}).format(date)
}

// An end date picked in the editor is the last second of that day in the zone, so it reads as a plain date.
function formatUntil(until: unknown, opts: DescribeOptions): string | null {
	// Date parsing rolls impossible values over (Feb 30 -> Mar 2), which would name a day the rule does not.
	if (!isStrictInstant(until)) {
		return null
	}
	const instant = new Date(until as string)
	const zoned = dayjs(instant).tz(opts.zone || Intl.DateTimeFormat().resolvedOptions().timeZone)
	const isEndOfDay = zoned.hour() === 23 && zoned.minute() === 59 && zoned.second() === 59
	const sameYear = dayInZone(instant, opts.zone).year === dayInZone(opts.dueDate ?? new Date(), opts.zone).year
	const style: ZonedDateStyle = isEndOfDay
		? (sameYear ? 'monthDay' : 'monthDayYear')
		: (sameYear ? 'dateTime' : 'dateTimeYear')
	return formatInZone(instant, opts.zone, opts.locale, style)
}

function ending(count: unknown, until: unknown, opts: DescribeOptions): string {
	if (isPositiveInteger(count)) {
		return opts.t(`${KEY}.endCount`, {n: count}, count)
	}
	const date = formatUntil(until, opts)
	return date === null ? '' : opts.t(`${KEY}.endUntil`, {date})
}

function everyN(key: string, n: number, t: Translate, params: Record<string, unknown> = {}): string {
	return t(`${KEY}.${key}`, {n, ...params}, n)
}

function describeEditable(form: RepeatForm, opts: DescribeOptions): string {
	const {t, locale, zone, dueDate} = opts
	const n = form.interval
	const due = dueDate ? dayInZone(dueDate, zone) : null
	switch (form.unit) {
		case 'hour':
			return everyN('hourly', n, t)
		case 'day':
			return everyN('daily', n, t)
		case 'week':
			if (form.weekdays === null) {
				return due ? everyN('weeklyOn', n, t, {days: weekdayName(due.weekday, locale)}) : everyN('weekly', n, t)
			}
			if (n === 1 && isWorkweek(form.weekdays)) {
				return t(`${KEY}.weekdays`)
			}
			return everyN('weeklyOn', n, t, {days: weekdayList(form.weekdays, opts)})
		case 'month':
			if (form.monthMode === 'lastDay') {
				return everyN('monthlyOnLastDay', n, t)
			}
			if (form.monthMode === 'nthWeekday') {
				const weekday = weekdayName(form.monthWeekday, locale)
				return form.ordinal === -1
					? everyN('monthlyOnLastWeekday', n, t, {weekday})
					: everyN('monthlyOnNth', n, t, {ordinal: formatOrdinal(form.ordinal, opts), weekday})
			}
			if (form.monthDay !== null) {
				return everyN('monthlyOnDay', n, t, {day: formatOrdinal(form.monthDay, opts)})
			}
			return due ? everyN('monthlyOnDay', n, t, {day: formatOrdinal(due.day, opts)}) : everyN('monthly', n, t)
		case 'year':
			return due
				? everyN('yearlyOn', n, t, {date: formatMonthDay(due.month, due.day, locale)})
				: everyN('yearly', n, t)
	}
}

// The legacy monthly migration turns a task due on the 29th to 31st into BYMONTHDAY=28..D;BYSETPOS=-1: the
// last of those days each month has. D=31 is always the last day; D=29 or 30 falls back in shorter months.
function describeMonthEnd(days: unknown, setPos: unknown, n: number, opts: DescribeOptions): string | null {
	if (!Array.isArray(setPos) || setPos.length !== 1 || setPos[0] !== -1 || !uniqueIntegersIn(days, 28, 31)) {
		return null
	}
	const sorted = ascending(days)
	const last = sorted[sorted.length - 1]
	if (sorted[0] !== 28 || last < 29 || sorted.length !== last - 27) {
		return null
	}
	return last === 31
		? everyN('monthlyOnLastDay', n, opts.t)
		: everyN('monthlyOnDayOrLastDay', n, opts.t, {day: formatOrdinal(last, opts)})
}

// Shapes the editor cannot build but that still read as one plain sentence.
function describeExtended(c: Record<string, unknown>, r: TaskRepeat, opts: DescribeOptions): string | null {
	const {count, until, interval, freq} = c
	const extraKeys = Object.keys(c).filter(key => !['freq', 'interval', 'wkst', 'count', 'until'].includes(key))
	const scalarsPlain = c.wkst === 'mo' && isPositiveInteger(interval) &&
		(count === undefined || isPositiveInteger(count)) &&
		(until === undefined || formatUntil(r.until, opts) !== null) &&
		!(count !== undefined && until !== undefined)
	if (!scalarsPlain) {
		return null
	}
	const has = (...keys: string[]) => extraKeys.length === keys.length && keys.every(key => extraKeys.includes(key))
	const {t} = opts
	let sentence: string | null = null

	if (freq === 'daily' && has('by_day') && isPlainWeekdayList(c.by_day)) {
		sentence = interval === 1 && isWorkweek(c.by_day)
			? t(`${KEY}.weekdays`)
			: everyN('dailyOn', interval, t, {days: weekdayList(c.by_day, opts)})
	} else if (freq === 'monthly' && has('by_day', 'by_set_pos') && isPlainWeekdayList(c.by_day) && isWorkweek(c.by_day) &&
		(c.by_set_pos as unknown[]).length === 1 && (c.by_set_pos as unknown[])[0] === -1) {
		sentence = everyN('lastWeekdayOfMonth', interval, t)
	} else if (freq === 'monthly' && has('by_month_day', 'by_set_pos') && isPositiveInteger(interval)) {
		sentence = describeMonthEnd(c.by_month_day, c.by_set_pos, interval, opts)
	} else if (freq === 'yearly' && has('by_month') && opts.dueDate && (c.by_month as unknown[]).length === 1) {
		const month = (c.by_month as unknown[])[0]
		const date = Number.isInteger(month)
			? formatMonthDay(month as number, dayInZone(opts.dueDate, opts.zone).day, opts.locale)
			: null
		sentence = date === null ? null : everyN('yearlyOn', interval, t, {date})
	}
	return sentence === null ? null : sentence + ending(count, r.until, opts)
}

function raw(value: unknown): string {
	if (Array.isArray(value)) {
		return value.map(raw).join(',')
	}
	return typeof value === 'object' && value !== null ? JSON.stringify(value) : String(value)
}

interface ByDayCode {
	weekday: string
	n: number | null
}

// Codes the scheduler would reject (an ordinal of 0 or beyond 53) are not parsed.
function parseByDay(code: unknown): ByDayCode | null {
	const match = typeof code === 'string' ? BY_DAY_CODE.exec(code.toLowerCase()) : null
	if (!match) {
		return null
	}
	const n = match[1] === undefined ? null : Number(match[1])
	if (n !== null && (n === 0 || Math.abs(n) > 53)) {
		return null
	}
	return {weekday: match[2], n}
}

function nthWeekdayPhrase(code: string, n: number, opts: DescribeOptions): string {
	const weekday = weekdayName(code, opts.locale)
	if (n === -1) {
		return opts.t(`${KEY}.part.lastWeekday`, {weekday})
	}
	return n > 0
		? opts.t(`${KEY}.part.nthWeekday`, {ordinal: formatOrdinal(n, opts), weekday})
		: opts.t(`${KEY}.part.nthLastWeekday`, {ordinal: formatOrdinal(-n, opts), weekday})
}

function rawPart(field: string, r: TaskRepeat): string {
	return `${field}=${raw((r as unknown as Record<string, unknown>)[field])}`
}

function isKnownFreq(freq: unknown): freq is string {
	return typeof freq === 'string' && KNOWN_FREQS.includes(freq)
}

// Duplicates are refused: the scheduler emits a repeated time twice and counts it in set positions.
function uniqueIntegersIn(value: unknown, min: number, max: number): value is number[] {
	return Array.isArray(value) && new Set(value).size === value.length &&
		value.every(v => Number.isInteger(v) && v >= min && v <= max)
}

function ascending(values: number[]): number[] {
	return [...values].sort((a, b) => a - b)
}

function byDayCodes(value: unknown): ByDayCode[] | null {
	const codes = Array.isArray(value) ? value.map(parseByDay) : null
	if (codes === null || codes.some(code => code === null)) {
		return null
	}
	const parsed = codes as ByDayCode[]
	return new Set(parsed.map(code => `${code.n}${code.weekday}`)).size === parsed.length ? parsed : null
}

function numberPart(field: string, c: Record<string, unknown>, opts: DescribeOptions): string | null {
	const {key, param, min, max} = NUMBER_PARTS[field]!
	const value = c[field]
	return uniqueIntegersIn(value, min, max)
		? opts.t(`${KEY}.part.${key}`, {[param]: formatList(ascending(value).map(String), opts.locale)})
		: null
}

function monthPart(c: Record<string, unknown>, opts: DescribeOptions): string | null {
	return uniqueIntegersIn(c.by_month, 1, 12)
		? opts.t(`${KEY}.part.byMonth`, {months: formatList(ascending(c.by_month).map(m => monthName(m, opts.locale)), opts.locale)})
		: null
}

function everyWeekdayPart(codes: ByDayCode[], opts: DescribeOptions): string | null {
	return codes.every(code => code.n === null)
		? opts.t(`${KEY}.part.byDayEvery`, {days: weekdayList(codes.map(code => code.weekday), opts)})
		: null
}

// Within a month only ordinals up to 5 are safe: the scheduler indexes its weekday mask out of range beyond that.
const MONTH_ORDINAL_LIMIT = 5
const YEAR_ORDINAL_LIMIT = 53

function nthWeekdaysPart(codes: ByDayCode[], limit: number, opts: DescribeOptions): string | null {
	return codes.every(code => code.n !== null && Math.abs(code.n) <= limit)
		? opts.t(`${KEY}.part.byDay`, {days: formatList(codes.map(code => nthWeekdayPhrase(code.weekday, code.n as number, opts)), opts.locale)})
		: null
}

function weekdaysAlonePart(c: Record<string, unknown>, opts: DescribeOptions): string | null {
	const codes = byDayCodes(c.by_day)
	if (codes === null) {
		return null
	}
	if (c.freq === 'daily' || c.freq === 'weekly') {
		// Below monthly the scheduler drops the ordinal, so "2mo" and "mo" are the same Monday.
		const weekdays = codes.map(code => code.weekday)
		return new Set(weekdays).size === weekdays.length
			? opts.t(`${KEY}.part.byDay`, {days: weekdayList(weekdays, opts)})
			: null
	}
	// Monthly and yearly rules match plain and ordinal weekdays conjunctively, so a mixed list has no sentence.
	return everyWeekdayPart(codes, opts) ??
		nthWeekdaysPart(codes, c.freq === 'monthly' ? MONTH_ORDINAL_LIMIT : YEAR_ORDINAL_LIMIT, opts)
}

interface ExactSelectors {
	freqs: string[]
	fields: Selector[]
	build: (c: Record<string, unknown>, opts: DescribeOptions) => (string | null)[]
}

// The only selector combinations described in words, each checked against rrule-go's source and a probe of the
// occurrences it produces. Anything else (including every by_week_no, by_set_pos, by_minute and by_second) is shown as stored;
// rrule-go miscounts week numbers around the new year, so no week number is put into words.
const EXACT_SELECTORS: ExactSelectors[] = [
	{freqs: ['daily', 'weekly', 'monthly', 'yearly'], fields: ['by_month'], build: (c, o) => [monthPart(c, o)]},
	{freqs: ['daily', 'weekly', 'monthly', 'yearly'], fields: ['by_day'], build: (c, o) => [weekdaysAlonePart(c, o)]},
	{freqs: ['monthly'], fields: ['by_month_day'], build: (c, o) => [numberPart('by_month_day', c, o)]},
	{freqs: ['yearly'], fields: ['by_year_day'], build: (c, o) => [numberPart('by_year_day', c, o)]},
	{freqs: ['daily'], fields: ['by_hour'], build: (c, o) => [numberPart('by_hour', c, o)]},
	{
		freqs: ['yearly'],
		fields: ['by_month', 'by_day'],
		build: (c, o) => {
			const codes = byDayCodes(c.by_day)
			return [monthPart(c, o), codes?.length === 1 ? nthWeekdaysPart(codes, MONTH_ORDINAL_LIMIT, o) : null]
		},
	},
]

// Fails closed: one selector without exact wording puts every selector in its stored form.
function selectorParts(c: Record<string, unknown>, r: TaskRepeat, opts: DescribeOptions): string[] {
	const present = SELECTORS.filter(field => field in c)
	const exact = EXACT_SELECTORS.find(entry => entry.freqs.includes(c.freq as string) &&
		entry.fields.length === present.length && entry.fields.every(field => present.includes(field)))
	const parts = exact?.build(c, opts)
	if (parts !== undefined && parts.every(part => part !== null)) {
		return parts as string[]
	}
	return present.map(field => rawPart(field, r))
}

function frequencyParts(c: Record<string, unknown>, r: TaskRepeat, t: Translate): string[] {
	const {freq, interval} = c
	if (!isKnownFreq(freq)) {
		return interval === 1 ? [rawPart('freq', r)] : [rawPart('freq', r), rawPart('interval', r)]
	}
	if (Number.isInteger(interval) && (interval as number) >= 2) {
		return [t(`${KEY}.part.every.${freq}`, {n: interval}, interval as number)]
	}
	const parts = [t(`${KEY}.part.freq`, {freq: t(`${KEY}.freq.${freq}`)})]
	// The canonical form already reads 0 as 1; a negative interval makes the scheduler reject the rule.
	if (interval !== 1) {
		parts.push(rawPart('interval', r))
	}
	return parts
}

function trailingPart(field: 'wkst' | 'count' | 'until', c: Record<string, unknown>, r: TaskRepeat, opts: DescribeOptions): string {
	const {t, locale} = opts
	switch (field) {
		case 'wkst':
			return typeof r.wkst === 'string' && WEEKDAYS.includes(r.wkst.toLowerCase() as typeof WEEKDAYS[number])
				? t(`${KEY}.part.wkst`, {weekday: weekdayName(r.wkst, locale)})
				: rawPart(field, r)
		case 'count':
			if (isPositiveInteger(c.count)) {
				return t(`${KEY}.part.count`, {n: c.count}, c.count)
			}
			// The scheduler reads a count of 0 or less as no limit on the number of occurrences.
			return Number.isInteger(c.count) ? t(`${KEY}.part.countUnlimited`) : rawPart(field, r)
		case 'until':
			return t(`${KEY}.part.until`, {date: formatUntil(r.until, opts) ?? raw(r.until)})
	}
}

function unknownKeys(r: TaskRepeat): string[] {
	return Object.keys(r).filter(key => !KNOWN_FIELDS.has(key))
}

function describeGeneric(c: Record<string, unknown>, r: TaskRepeat, opts: DescribeOptions): string {
	const parts = [...frequencyParts(c, r, opts.t), ...selectorParts(c, r, opts)]
	for (const field of ['wkst', 'count', 'until'] as const) {
		// The canonical form fills in wkst; only a stored one is described.
		const present = field === 'wkst' ? hasValue(r.wkst) : field in c
		if (present) {
			parts.push(trailingPart(field, c, r, opts))
		}
	}
	// Read from the stored object: canonicalisation drops empty values, but an unknown key is shown even when empty.
	for (const key of unknownKeys(r)) {
		const value = (r as unknown as Record<string, unknown>)[key]
		parts.push(`${key}=${typeof value === 'string' ? value : JSON.stringify(value)}`)
	}
	const list = cached(`unit|${opts.locale}`, () => new ListFormat(opts.locale, {type: 'unit'})).format(parts)
	return opts.t(`${KEY}.custom`, {parts: list})
}

export function describeRepeatRule(repeat: TaskRepeat | null | undefined, opts: DescribeOptions): string {
	const dueDate = opts.dueDate ?? null
	const kind = classifyRepeat(repeat, dueDate)
	if (!repeat || kind === 'none') {
		return ''
	}
	// The server's projection of a read-only rule can omit parts (BYEASTER), so describing it would mislead.
	if (kind === 'locked') {
		return opts.t(`${KEY}.locked`)
	}
	if (kind === 'editable') {
		const form = repeatToForm(repeat, dueDate, opts.zone)
		const count = form.end === 'count' ? form.count : undefined
		const until = form.end === 'until' ? form.untilInstant : undefined
		return describeEditable(form, opts) + ending(count, until, opts)
	}
	const canonical = canonicalRepeat(repeat)
	const extended = unknownKeys(repeat).length === 0 ? describeExtended(canonical, repeat, opts) : null
	return extended ?? describeGeneric(canonical, repeat, opts)
}
