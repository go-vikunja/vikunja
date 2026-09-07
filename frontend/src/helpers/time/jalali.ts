import dayjs from 'dayjs'
import timezone from 'dayjs/plugin/timezone'
import utc from 'dayjs/plugin/utc'
import {
	d2j,
	isValidJalaaliDate,
	j2d,
	jalaaliMonthLength,
	toGregorian,
	toJalaali,
} from 'jalaali-js'

import {parseDateOrNull} from '@/helpers/parseDateOrNull'

dayjs.extend(utc)
dayjs.extend(timezone)

export const JALALI_LOCALE = 'fa-IR'
export const JALALI_CALENDAR_LOCALE = 'fa-IR-u-ca-persian'
export const JALALI_CALENDAR_LATIN_DIGITS_LOCALE = 'fa-IR-u-ca-persian-nu-latn'

const PERSIAN_DIGITS = '۰۱۲۳۴۵۶۷۸۹'
const ARABIC_INDIC_DIGITS = '٠١٢٣٤٥٦٧٨٩'
const DIGIT_MAP: Record<string, string> = {}

for (let i = 0; i < 10; i++) {
	DIGIT_MAP[PERSIAN_DIGITS[i]] = String(i)
	DIGIT_MAP[ARABIC_INDIC_DIGITS[i]] = String(i)
}

const NON_ASCII_DIGIT_PATTERN = new RegExp(`[${PERSIAN_DIGITS}${ARABIC_INDIC_DIGITS}]`, 'g')

/**
 * Whether the given locale should use the Jalali calendar for presentation.
 * Jalali is a presentation/input concern only: the internal representation
 * stays a Gregorian instant serialized via toISOString().
 */
export function isJalaliLocale(locale: string | null | undefined): boolean {
	if (typeof locale !== 'string') {
		return false
	}

	const normalized = locale.trim().toLowerCase()
	return normalized === 'fa-ir' || normalized === 'fa'
}

/**
 * Converts Persian (U+06F0-U+06F9) and Arabic-Indic (U+0660-U+0669) digits
 * to ASCII digits. All other characters are left unchanged.
 */
export function normalizePersianDigits(value: string): string {
	return value.replace(NON_ASCII_DIGIT_PATTERN, (digit) => DIGIT_MAP[digit])
}

/**
 * Converts ASCII digits to Persian (U+06F0-U+06F9) digits for display.
 * All other characters are left unchanged.
 */
export function toPersianDigits(value: string | number): string {
	return value.toString().replace(/[0-9]/g, (digit) => PERSIAN_DIGITS[Number(digit)])
}

/**
 * Resolves the timezone display code should use: the user's configured
 * timezone when set, otherwise the browser's local timezone. Never throws:
 * an unusable configured value falls back instead of breaking rendering.
 */
export function resolveUserTimezone(configuredTimezone?: string | null): string {
	const fallback = safeBrowserTimezone()

	if (typeof configuredTimezone !== 'string' || configuredTimezone.trim() === '') {
		return fallback
	}

	return isValidTimezone(configuredTimezone) ? configuredTimezone : fallback
}

function safeBrowserTimezone(): string {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone ?? 'UTC'
	} catch {
		return 'UTC'
	}
}

function isValidTimezone(timezone: string): boolean {
	try {
		new Intl.DateTimeFormat(undefined, {timeZone: timezone})
		return true
	} catch {
		return false
	}
}

export interface JalaliDate {
	// 1-based: Farvardin = 1, Esfand = 12.
	year: number
	month: number
	day: number
}

export interface JalaliDateTime extends JalaliDate {
	hours: number
	minutes: number
}

/**
 * Gregorian (1-based month) to Jalali. Pure calendar math via jalaali-js,
 * no timezone involved: a calendar date maps to exactly one Jalali date.
 */
export function gregorianToJalali(year: number, month: number, day: number): JalaliDate {
	const {jy, jm, jd} = toJalaali(year, month, day)
	return {year: jy, month: jm, day: jd}
}

/**
 * Jalali to Gregorian (1-based month). Returns null for impossible dates
 * such as month 13 or Esfand 30 in a common year.
 */
export function jalaliToGregorian(date: JalaliDate): {year: number, month: number, day: number} | null {
	if (
		!Number.isInteger(date.year) || !Number.isInteger(date.month) || !Number.isInteger(date.day) ||
		!isValidJalaaliDate(date.year, date.month, date.day)
	) {
		return null
	}

	const {gy, gm, gd} = toGregorian(date.year, date.month, date.day)
	return {year: gy, month: gm, day: gd}
}

/**
 * Days in a Jalali month (month is 1-based). Esfand correctly yields 29 or
 * 30 depending on the leap year.
 */
export function getJalaliMonthLength(year: number, month: number): number {
	return jalaaliMonthLength(year, month)
}

/**
 * Weekday of a Jalali date in JavaScript convention (0 = Sunday, 6 = Saturday).
 * Zone-independent: a Jalali date is exactly one Gregorian date.
 * Returns null for impossible dates.
 */
export function jalaliWeekday(date: JalaliDate): number | null {
	const gregorian = jalaliToGregorian(date)
	if (gregorian === null) {
		return null
	}

	return new Date(Date.UTC(gregorian.year, gregorian.month - 1, gregorian.day)).getUTCDay()
}

/**
 * Adds (or subtracts) days to a Jalali date via Julian Day Numbers, so month
 * and year boundaries - including Esfand in leap years - are always correct.
 */
export function addJalaliDays(date: JalaliDate, days: number): JalaliDate {
	const {jy, jm, jd} = d2j(j2d(date.year, date.month, date.day) + days)
	return {year: jy, month: jm, day: jd}
}

/**
 * Weekday of an instant in the given timezone, JavaScript convention
 * (0 = Sunday, 6 = Saturday). Used to anchor shortcut math to the user's
 * wall-clock day rather than the machine zone.
 */
export function weekdayInTimezone(date: Date, timeZone: string): number {
	return dayjs(date).tz(resolveUserTimezone(timeZone)).day()
}

/**
 * Splits an instant into Jalali date plus wall-clock time in the given
 * timezone. Returns null for unusable dates. This is the exact inverse of
 * jalaliToInstant: the same instant formatted by Phase 1 display code in the
 * same timezone shows this Jalali date.
 */
export function instantToJalali(date: Date | string | null | undefined, timeZone: string): JalaliDateTime | null {
	const parsed = parseDateOrNull(date)
	if (parsed === null) {
		return null
	}

	const zoned = dayjs(parsed).tz(resolveUserTimezone(timeZone))
	const {jy, jm, jd} = toJalaali(zoned.year(), zoned.month() + 1, zoned.date())
	return {year: jy, month: jm, day: jd, hours: zoned.hour(), minutes: zoned.minute()}
}

function pad2(value: number): string {
	return value.toString().padStart(2, '0')
}

/**
 * Builds the instant for a Jalali date plus wall-clock time in the given
 * timezone - e.g. 1405/01/01 00:00 in Asia/Tehran. Returns null for
 * impossible dates or out-of-range times. The timezone is sanitized, so this
 * never throws on bad input.
 */
export function jalaliToInstant(date: JalaliDateTime, timeZone: string): Date | null {
	const gregorian = jalaliToGregorian(date)
	if (
		gregorian === null ||
		!Number.isInteger(date.hours) || !Number.isInteger(date.minutes) ||
		date.hours < 0 || date.hours > 23 || date.minutes < 0 || date.minutes > 59
	) {
		return null
	}

	const wallClock = `${gregorian.year}-${pad2(gregorian.month)}-${pad2(gregorian.day)} ${pad2(date.hours)}:${pad2(date.minutes)}`
	const instant = dayjs.tz(wallClock, resolveUserTimezone(timeZone))
	return instant.isValid() ? instant.toDate() : null
}

function toValidDate(date: Date | string | null | undefined): Date | null {
	return parseDateOrNull(date)
}

/**
 * Formats a Gregorian instant using the Jalali calendar via the native
 * Intl API. Returns an empty string for anything that isn't a usable date
 * (null, undefined, invalid Date, api zero time) since Intl throws on those.
 *
 * When called without options, year/month/day default to numeric. When
 * options are given, only timeZone defaults (to UTC) and the caller fully
 * controls the components - so `{weekday: 'short'}` really yields just the
 * weekday instead of weekday plus a date.
 *
 * Pass an explicit timeZone (e.g. the user's timezone, 'UTC' in tests) for
 * deterministic output.
 */
export function formatJalaliDate(
	date: Date | string | null | undefined,
	options?: Intl.DateTimeFormatOptions,
	locale: string = JALALI_CALENDAR_LOCALE,
): string {
	const parsed = toValidDate(date)
	if (parsed === null) {
		return ''
	}

	const {timeZone = 'UTC', ...rest} = options ?? {year: 'numeric', month: '2-digit', day: '2-digit'}

	return new Intl.DateTimeFormat(locale, {timeZone, ...rest}).format(parsed)
}

/**
 * Same guarantees as formatJalaliDate, but returns the Intl parts so future
 * code can build custom strings (e.g. jYYYY/jMM/jDD) without re-parsing.
 * Returns an empty array for unusable dates.
 */
export function formatJalaliParts(
	date: Date | string | null | undefined,
	options?: Intl.DateTimeFormatOptions,
	locale: string = JALALI_CALENDAR_LOCALE,
): Intl.DateTimeFormatPart[] {
	const parsed = toValidDate(date)
	if (parsed === null) {
		return []
	}

	const {timeZone = 'UTC', ...rest} = options ?? {year: 'numeric', month: '2-digit', day: '2-digit'}

	return new Intl.DateTimeFormat(locale, {timeZone, ...rest}).formatToParts(parsed)
}
