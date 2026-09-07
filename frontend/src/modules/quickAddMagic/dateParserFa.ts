import {useAuthStore} from '@/stores/auth'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {getDefaultTimeParts} from '@/helpers/time/getDateWithTime'
import {replaceAll} from '@/helpers/replaceAll'
import {
	addJalaliDays,
	getJalaliMonthLength,
	instantToJalali,
	jalaliToGregorian,
	jalaliToInstant,
	normalizePersianDigits,
	resolveUserTimezone,
	weekdayInTimezone,
} from '@/helpers/time/jalali'
import {REPEAT_TYPES} from '@/types/IRepeatAfter'
import type {repeatParsedResult} from './types'
import type {dateParseResult} from './dateParser'

export interface PersianDateOptions {
	locale?: string,
	timeZone?: string,
}

function resolveFaTimeZone(provided?: string | null): string {
	if (typeof provided === 'string' && provided.trim() !== '') {
		return resolveUserTimezone(provided)
	}
	try {
		return resolveUserTimezone(useAuthStore().settings.timezone)
	} catch {
		return resolveUserTimezone()
	}
}

function getFaDefaultTime(now: Date): {hours: number, minutes: number} {
	try {
		return getDefaultTimeParts(now)
	} catch {
		return {hours: calculateNearestHours(now), minutes: 0}
	}
}

function findFaKeyword(working: string, original: string, corePattern: string): {matchedNormalized: string, matchedOriginal: string} | null {
	const re = new RegExp(`(^|[\\s\u200c])(${corePattern})(?=$|[\\s\u200c](?!ها))`)
	const m = re.exec(working)
	if (m === null) {
		return null
	}
	const leading = m[1]
	const core = m[2]
	const coreStart = m.index + leading.length
	return {
		matchedNormalized: core,
		matchedOriginal: original.slice(coreStart, coreStart + core.length),
	}
}

function normalizeWeekdayKey(value: string): string {
	return value
		.replace(/[\s\u200c]/g, '')
		.replace(/ي/g, 'ی')
		.replace(/ك/g, 'ک')
}

function weekdayToJsDay(matched: string): number | null {
	const key = normalizeWeekdayKey(matched)
	switch (key) {
		case 'شنبه':
			return 6
		case 'یکشنبه':
			return 0
		case 'دوشنبه':
			return 1
		case 'سهشنبه':
			return 2
		case 'چهارشنبه':
			return 3
		case 'پنجشنبه':
			return 4
		case 'جمعه':
			return 5
		default:
			return null
	}
}

function normalizeMonthKey(value: string): string {
	return value
		.replace(/ي/g, 'ی')
		.replace(/ك/g, 'ک')
		.replace(/آ/g, 'ا')
}

function monthNameToNumber(matched: string): number | null {
	const key = normalizeMonthKey(matched)
	switch (key) {
		case 'فروردین':
			return 1
		case 'اردیبهشت':
			return 2
		case 'خرداد':
			return 3
		case 'تیر':
			return 4
		case 'مرداد':
			return 5
		case 'شهریور':
			return 6
		case 'مهر':
			return 7
		case 'ابان':
			return 8
		case 'اذر':
			return 9
		case 'دی':
			return 10
		case 'بهمن':
			return 11
		case 'اسفند':
			return 12
		default:
			return null
	}
}

const FA_MONTHS = 'فرورد[یی]ن|ارد[یی]بهشت|خرداد|ت[یی]ر|مرداد|شهر[یی]ور|مهر|[آا]بان|[آا]ذر|د[یی]|بهمن|اسفند'

function extractFaTime(workingWithout: string, originalWithout: string): {hours: number, minutes: number, matchedOriginal: string} | null {
	const faRe = new RegExp('(^|[\\s‌])(ساعت[\\s‌]+(\\d{1,2}):(\\d{2}))')
	const faMatch = faRe.exec(workingWithout)
	if (faMatch !== null) {
		const hours = Number(faMatch[3])
		const minutes = Number(faMatch[4])
		if (hours >= 0 && hours <= 23 && minutes >= 0 && minutes <= 59) {
			const leading = faMatch[1]
			const core = faMatch[2]
			const coreStart = faMatch.index + leading.length
			return {
				hours,
				minutes,
				matchedOriginal: originalWithout.slice(coreStart, coreStart + core.length),
			}
		}
	}

	const enRe = new RegExp('(^|[\\s‌])((at|@)[\\s‌]+(\\d{1,2})(:(\\d{2}))?(\\s*(am|pm))?)', 'i')
	const enMatch = enRe.exec(workingWithout)
	if (enMatch !== null) {
		let hours = parseInt(enMatch[4], 10)
		let minutes = 0
		if (Number.isNaN(hours)) {
			return null
		}
		const meridiem = enMatch[7]?.toLowerCase()
		if (meridiem === 'pm' && hours !== 12) {
			hours += 12
		} else if (meridiem === 'am' && hours === 12) {
			hours = 0
		}
		if (typeof enMatch[6] !== 'undefined') {
			minutes = parseInt(enMatch[6], 10)
		}
		if (Number.isNaN(minutes) || hours < 0 || hours > 23 || minutes < 0 || minutes > 59) {
			return null
		}
		const leading = enMatch[1]
		const core = enMatch[2]
		const coreStart = enMatch.index + leading.length
		return {
			hours,
			minutes,
			matchedOriginal: originalWithout.slice(coreStart, coreStart + core.length),
		}
	}

	return null
}

function applyFaTimeAndConvert(
	originalText: string,
	matchedOriginalDate: string,
	year: number,
	month: number,
	day: number,
	timeZone: string,
	defaults: {hours: number, minutes: number},
): dateParseResult | null {
	const withoutDate = replaceAll(originalText, matchedOriginalDate, '')
	const workingWithout = normalizePersianDigits(withoutDate)
	const time = extractFaTime(workingWithout, withoutDate)

	let hours = defaults.hours
	let minutes = defaults.minutes
	let newText = withoutDate.trim()
	if (time !== null) {
		hours = time.hours
		minutes = time.minutes
		newText = replaceAll(withoutDate, time.matchedOriginal, '').trim()
	}

	const date = jalaliToInstant({year, month, day, hours, minutes}, timeZone)
	if (date === null) {
		return null
	}
	return {newText, date}
}

function findFaNumeric(working: string, original: string): {matchedOriginal: string, year: number, month: number, day: number} | null {
	const re = new RegExp('(^|[\\s‌])(\\d{4}[\\s‌]*[/.-][\\s‌]*(\\d{1,2})[\\s‌]*[/.-][\\s‌]*(\\d{1,2}))(?=$|[\\s‌](?!ها))', 'g')
	let m: RegExpExecArray | null
	while ((m = re.exec(working)) !== null) {
		const leading = m[1]
		const core = m[2]
		const parts = core.split(/[/.-]/)
		const year = Number(parts[0])
		const month = Number(parts[1])
		const day = Number(parts[2])
		if (year < 1300 || year > 1500) {
			continue
		}
		if (jalaliToGregorian({year, month, day}) === null) {
			continue
		}
		const matchEnd = m.index + m[0].length
		const isAtStart = m.index === 0
		const isAtEnd = matchEnd >= working.length
		if (!isAtStart && !isAtEnd) {
			const after = working.slice(matchEnd)
			if (!/^[\s\u200c]+(ساعت[\s\u200c]+\d|at[\s\u200c]+|@[\s\u200c]+)/i.test(after)) {
				continue
			}
		}
		const coreStart = m.index + leading.length
		return {
			matchedOriginal: original.slice(coreStart, coreStart + core.length),
			year,
			month,
			day,
		}
	}
	return null
}

function findFaDayMonth(working: string, original: string): {matchedOriginal: string, day: number, month: number} | null {
	const re = new RegExp(`(^|[\\s\u200c])(\\d{1,2})[\\s\u200c]+(${FA_MONTHS})(?=$|[\\s\u200c](?!ها))`)
	const m = re.exec(working)
	if (m === null) {
		return null
	}
	const leading = m[1]
	const core = m[0].slice(leading.length)
	const day = Number(m[2])
	const month = monthNameToNumber(m[3])
	if (month === null) {
		return null
	}
	const coreStart = m.index + leading.length
	return {
		matchedOriginal: original.slice(coreStart, coreStart + core.length),
		day,
		month,
	}
}

export const parsePersianDate = (
	text: string,
	now: Date = new Date(),
	options?: PersianDateOptions,
): dateParseResult | null => {
	const timeZone = resolveFaTimeZone(options?.timeZone)
	const cur = instantToJalali(now, timeZone)
	if (cur === null) {
		return null
	}
	const curWeekday = weekdayInTimezone(now, timeZone)
	const defaults = getFaDefaultTime(now)
	const working = normalizePersianDigits(text)

	const afterTomorrow = findFaKeyword(working, text, 'پس[\\s\u200c]?فردا')
	if (afterTomorrow !== null) {
		const target = addJalaliDays(cur, 2)
		return applyFaTimeAndConvert(text, afterTomorrow.matchedOriginal, target.year, target.month, target.day, timeZone, defaults)
	}

	const today = findFaKeyword(working, text, 'امروز')
	if (today !== null) {
		const target = addJalaliDays(cur, 0)
		return applyFaTimeAndConvert(text, today.matchedOriginal, target.year, target.month, target.day, timeZone, defaults)
	}

	const tomorrow = findFaKeyword(working, text, 'فردا')
	if (tomorrow !== null) {
		const target = addJalaliDays(cur, 1)
		return applyFaTimeAndConvert(text, tomorrow.matchedOriginal, target.year, target.month, target.day, timeZone, defaults)
	}

	const weekdayRe = new RegExp('(^|[\\s‌])([یی][کك][\\s‌]?شنبه|دو[\\s‌]?شنبه|سه[\\s‌]?شنبه|چهار[\\s‌]?شنبه|پنج[\\s‌]?شنبه|شنبه|جمعه)(?=$|[\\s‌](?!ها))')
	const weekdayMatch = weekdayRe.exec(working)
	if (weekdayMatch !== null) {
		const targetDay = weekdayToJsDay(weekdayMatch[2])
		if (targetDay !== null) {
			const distance = (targetDay + 7 - curWeekday) % 7
			const target = addJalaliDays(cur, distance)
			const leading = weekdayMatch[1]
			const core = weekdayMatch[2]
			const coreStart = weekdayMatch.index + leading.length
			const matchedOriginal = text.slice(coreStart, coreStart + core.length)
			return applyFaTimeAndConvert(text, matchedOriginal, target.year, target.month, target.day, timeZone, defaults)
		}
	}

	const nextMonthStart = findFaKeyword(working, text, 'اول[\\s\u200c]+ماه[\\s\u200c]+بعد')
	if (nextMonthStart !== null) {
		const year = cur.month === 12 ? cur.year + 1 : cur.year
		const month = cur.month === 12 ? 1 : cur.month + 1
		return applyFaTimeAndConvert(text, nextMonthStart.matchedOriginal, year, month, 1, timeZone, defaults)
	}

	const endOfMonth = findFaKeyword(working, text, '[آا]خر[\\s\u200c]+ماه')
	if (endOfMonth !== null) {
		const day = getJalaliMonthLength(cur.year, cur.month)
		return applyFaTimeAndConvert(text, endOfMonth.matchedOriginal, cur.year, cur.month, day, timeZone, defaults)
	}

	const numeric = findFaNumeric(working, text)
	if (numeric !== null) {
		return applyFaTimeAndConvert(text, numeric.matchedOriginal, numeric.year, numeric.month, numeric.day, timeZone, defaults)
	}

	const dayMonth = findFaDayMonth(working, text)
	if (dayMonth !== null) {
		let year = cur.year
		if (dayMonth.month < cur.month || (dayMonth.month === cur.month && dayMonth.day < cur.day)) {
			year = cur.year + 1
		}
		if (jalaliToGregorian({year, month: dayMonth.month, day: dayMonth.day}) === null) {
			return null
		}
		return applyFaTimeAndConvert(text, dayMonth.matchedOriginal, year, dayMonth.month, dayMonth.day, timeZone, defaults)
	}

	const timeOnlyRe = new RegExp('(^|[\\s‌])(ساعت[\\s‌]+(\\d{1,2}):(\\d{2}))')
	const timeOnlyMatch = timeOnlyRe.exec(working)
	if (timeOnlyMatch !== null) {
		const hours = Number(timeOnlyMatch[3])
		const minutes = Number(timeOnlyMatch[4])
		if (hours >= 0 && hours <= 23 && minutes >= 0 && minutes <= 59) {
			const leading = timeOnlyMatch[1]
			const core = timeOnlyMatch[2]
			const coreStart = timeOnlyMatch.index + leading.length
			const matchedOriginal = text.slice(coreStart, coreStart + core.length)
			const newText = replaceAll(text, matchedOriginal, '').trim()
			const date = jalaliToInstant({year: cur.year, month: cur.month, day: cur.day, hours, minutes}, timeZone)
			if (date !== null) {
				return {newText, date}
			}
		}
	}

	return null
}

const FA_REPEAT_CORE = 'هر[\\s\u200c]+ماه|ماهانه|هر[\\s\u200c]+سال|سالانه|هر[\\s\u200c]+هفته|هفتگی|هر[\\s\u200c]+روز|روزانه'

function repeatTypeFor(matched: string): typeof REPEAT_TYPES.Months | typeof REPEAT_TYPES.Years | typeof REPEAT_TYPES.Weeks | typeof REPEAT_TYPES.Days | null {
	const key = matched.replace(/[\s‌]/g, '').replace(/ي/g, 'ی')
	switch (key) {
		case 'هرماه':
		case 'ماهانه':
			return REPEAT_TYPES.Months
		case 'هرسال':
		case 'سالانه':
			return REPEAT_TYPES.Years
		case 'هرهفته':
		case 'هفتگی':
			return REPEAT_TYPES.Weeks
		case 'هرروز':
		case 'روزانه':
			return REPEAT_TYPES.Days
		default:
			return null
	}
}

export const getPersianRepeats = (text: string): repeatParsedResult => {
	const re = new RegExp(`(^|[\\s‌])(${FA_REPEAT_CORE})(?=$|[\\s‌](?!ها))`)
	const m = re.exec(normalizePersianDigits(text))
	if (m === null) {
		return {
			textWithoutMatched: text,
			repeats: null,
		}
	}
	const core = m[2]
	const type = repeatTypeFor(core)
	if (type === null) {
		return {
			textWithoutMatched: text,
			repeats: null,
		}
	}
	let fullOriginal = text.slice(m.index, m.index + m[0].length)
	if (fullOriginal.endsWith(' ') || fullOriginal.endsWith('‌')) {
		fullOriginal = fullOriginal.substring(0, fullOriginal.length - 1)
	}
	return {
		textWithoutMatched: text.replace(fullOriginal, ''),
		repeats: {
			amount: 1,
			type,
		},
	}
}
