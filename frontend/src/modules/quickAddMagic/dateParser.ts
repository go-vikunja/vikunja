import {calculateDayInterval} from '@/helpers/time/calculateDayInterval'
import {calculateNearestHours} from '@/helpers/time/calculateNearestHours'
import {replaceAll} from '@/helpers/replaceAll'
import {REPEAT_TYPES} from '@/types/IRepeatAfter'
import {getDefaultLocales, type QuickAddMagicLocale} from './locales'
import {findByForm, toPattern} from './locales/pattern'
import type {DatePhrase} from './locales/types'

export interface dateParseResult {
	newText: string,
	date: Date | null,
}

interface dateFoundResult {
	foundText: string | null,
	date: Date | null,
}

/**
 * Matches a date regex against text, rejecting matches that appear in the middle
 * of text with non-date content on both sides. This prevents false positives like
 * "The 9/11 Report" while still allowing "meeting 9/11 at 10:00".
 *
 * Matches at the start or end of text are always accepted. Middle matches are
 * only accepted when followed by a time expression (at/@ prefix).
 */
function matchDateAtBoundary(text: string, pattern: string): RegExpExecArray | null {
	const regex = new RegExp(`(^| )${pattern}($| )`, 'gi')
	let result: RegExpExecArray | null
	while ((result = regex.exec(text)) !== null) {
		const matchEnd = result.index + result[0].length
		const isAtStart = result.index === 0
		const isAtEnd = matchEnd >= text.length

		if (isAtStart || isAtEnd) return result

		// Allow middle-of-text matches when followed by a time expression
		const afterMatch = text.substring(matchEnd)
		if (/^(at |@ )/i.test(afterMatch)) return result
	}

	return null
}

/**
 * Returns the matched phrase text, or null. With strictBoundary, a space or
 * end must directly follow the phrase so words merely starting with it
 * (e.g. "завтрак") don't match.
 */
function matchDatePhrase(text: string, form: string, strictBoundary: boolean): string | null {
	const tail = strictBoundary ? '($| )' : ''
	const result = new RegExp(`(^| )${toPattern(form)}${tail}`, 'i').exec(text)
	return result === null ? null : result[0].trim()
}

const monthFormsGroup = (locale: QuickAddMagicLocale): string =>
	`(${locale.months.flatMap(m => m.forms).map(toPattern).join('|')})`

const timePrefixesPattern = (locales: QuickAddMagicLocale[]): string =>
	[...new Set(locales.flatMap(l => l.timePrefixes))].map(toPattern).join('|')

export const parseDate = (text: string, now: Date = new Date(), locales: QuickAddMagicLocale[] = getDefaultLocales()): dateParseResult => {
	for (const locale of locales) {
		for (const [phrase, forms] of Object.entries(locale.phrases)) {
			for (const form of forms) {
				const matched = matchDatePhrase(text, form, locale.phraseStrictBoundary ?? false)
				if (matched !== null) {
					return applyDatePhrase(text, phrase as DatePhrase, matched, locales)
				}
			}
		}
	}

	for (const locale of locales) {
		const parsed = getDateFromWeekday(text, now, locale)
		if (parsed.date !== null) {
			return addTimeToDate(text, parsed.date, parsed.foundText, locales)
		}
	}

	for (const locale of locales) {
		const parsed = getDayFromText(text, now, locale)
		if (parsed.date !== null) {
			const month = getMonthFromText(text, parsed.date as Date, locale)
			return addTimeToDate(month.newText, month.date, parsed.foundText, locales)
		}
	}

	const parsedIn = getDateFromTextIn(text, now, locales)
	if (parsedIn.date !== null) {
		return addTimeToDate(text, parsedIn.date, parsedIn.foundText, locales)
	}

	const parsed = getDateFromText(text, now, locales)

	if (parsed.date === null) {
		const time = addTimeToDate(text, new Date(now), parsed.foundText, locales)

		if (time.date !== null && +now !== +time.date) {
			return time
		}

		return {
			newText: replaceAll(text, parsed.foundText, ''),
			date: parsed.date,
		}
	}

	return addTimeToDate(text, parsed.date, parsed.foundText, locales)
}

const applyDatePhrase = (text: string, phrase: DatePhrase, matched: string, locales: QuickAddMagicLocale[]): dateParseResult => {
	switch (phrase) {
		case 'today':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('today')), matched, locales)
		case 'tonight': {
			const taskDate = getDateFromInterval(calculateDayInterval('today'))
			taskDate.setHours(21)
			return addTimeToDate(text, taskDate, matched, locales)
		}
		case 'tomorrow':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('tomorrow')), matched, locales)
		case 'nextMonday':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('nextMonday')), matched, locales)
		case 'thisWeekend':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('thisWeekend')), matched, locales)
		case 'laterThisWeek':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('laterThisWeek')), matched, locales)
		case 'laterNextWeek':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('laterNextWeek')), matched, locales)
		case 'nextWeek':
			return addTimeToDate(text, getDateFromInterval(calculateDayInterval('nextWeek')), matched, locales)
		case 'nextMonth': {
			const date: Date = new Date()
			date.setDate(1)
			date.setMonth(date.getMonth() + 1)
			date.setHours(calculateNearestHours(date))
			date.setMinutes(0)
			date.setSeconds(0)

			return addTimeToDate(text, date, matched, locales)
		}
		case 'endOfMonth': {
			const curDate: Date = new Date()
			const date: Date = new Date(curDate.getFullYear(), curDate.getMonth() + 1, 0)
			date.setHours(calculateNearestHours(date))
			date.setMinutes(0)
			date.setSeconds(0)

			return addTimeToDate(text, date, matched, locales)
		}
	}
}

const addTimeToDate = (text: string, date: Date, previousMatch: string | null, locales: QuickAddMagicLocale[]): dateParseResult => {
	previousMatch = previousMatch?.trim() || ''
	text = replaceAll(text, previousMatch, '')
	if (previousMatch === null) {
		return {
			newText: text,
			date: null,
		}
	}

	const timeRegex = ` (${timePrefixesPattern(locales)}) ([0-9][0-9]?(:[0-9][0-9])?( ?(a|p)m)?)`
	const matcher = new RegExp(timeRegex, 'ig')
	const results = matcher.exec(text)

	if (results !== null) {
		const time = results[2]
		const parts = time.split(':')
		let hours = parseInt(parts[0])
		let minutes = 0
		if (time.toLowerCase().endsWith('pm')) {
			if (hours !== 12) {
				hours += 12
			}
		} else if (time.toLowerCase().endsWith('am') && hours === 12) {
			hours = 0
		}
		if (parts.length > 1) {
			minutes = parseInt(parts[1])
		}

		date.setHours(hours)
		date.setMinutes(minutes)
		date.setSeconds(0)
		date.setMilliseconds(0)
	}

	const replace = results !== null ? results[0] : previousMatch
	return {
		newText: replaceAll(text, replace, '').trim(),
		date,
	}
}

export const getDateFromText = (text: string, now: Date = new Date(), locales: QuickAddMagicLocale[] = getDefaultLocales()) => {
	const datePatterns: string[] = [
		'(?<found>(?<month>[0-9][0-9]?)\\/(?<day>[0-9][0-9]?)(\\/(?<year>[0-9][0-9]([0-9][0-9])?))?)',
		'(?<found>(?<year>[0-9][0-9][0-9][0-9]?)\\/(?<month>[0-9][0-9]?)\\/(?<day>[0-9][0-9]))',
		'(?<found>(?<year>[0-9][0-9][0-9][0-9]?)-(?<month>[0-9][0-9]?)-(?<day>[0-9][0-9]))',
		'(?<found>(?<day>[0-9][0-9]?)\\.(?<month>[0-9][0-9]?)(\\.(?<year>[0-9][0-9]([0-9][0-9])?))?)',
	]

	let result: string | null = null
	let results: RegExpExecArray | null
	let foundText: string | null = ''
	let containsYear = true

	// 1. Try parsing the text as a "usual" date, like 2021-06-24 or "06/24/2021" or "27/01" or "01/27"
	for (const datePattern of datePatterns) {
		results = matchDateAtBoundary(text, datePattern)
		if (results !== null) {
			const {day, month, year, found} = {...results.groups}
			let tmp_year = year

			if (tmp_year === undefined) {
				tmp_year = year ?? now.getFullYear()
				containsYear = false
			}

			result = `${month}/${day}/${tmp_year}`
			result = !isNaN(new Date(result).getTime()) ? result : `${day}/${month}/${tmp_year}`
			result = !isNaN(new Date(result).getTime()) ? result : null

			if(result !== null){
				foundText = found
				break
			}
		}
	}

	if (result !== null) {
		const date = new Date(result)
		if (isNaN(date.getTime())) {
			return {
				foundText,
				date: null,
			}
		}

		if (!containsYear && date < now) {
			date.setFullYear(date.getFullYear() + 1)
		}

		return {
			foundText,
			date,
		}
	}

	// 2. Try parsing the date as something like "jan 21" or "21 jan" (or "21 серпня")
	for (const locale of locales) {
		const monthGroup = monthFormsGroup(locale)
		const monthRegex = new RegExp(`(^| )(${monthGroup} [0-9][0-9]?|[0-9][0-9]? ${monthGroup})`, 'i')
		results = monthRegex.exec(text)
		if (results === null) {
			continue
		}

		const dayMatch = /[0-9][0-9]?/.exec(results[0])
		const monthForm = results[0].replace(dayMatch?.[0] ?? '', '').trim()
		const monthDef = findByForm(locale.months, monthForm)
		if (dayMatch === null || monthDef === undefined) {
			continue
		}

		const day = parseInt(dayMatch[0])
		const date = new Date(now.getFullYear(), monthDef.month, day)
		if (date.getDate() !== day) {
			// day overflowed into the next month — reject like "mar 32"
			continue
		}

		if (date < now) {
			date.setFullYear(date.getFullYear() + 1)
		}

		return {
			foundText: results[0].trim(),
			date,
		}
	}

	return {
		foundText,
		date: null,
	}
}

export const getDateFromTextIn = (text: string, now: Date = new Date(), locales: QuickAddMagicLocale[] = getDefaultLocales()) => {
	for (const locale of locales) {
		const inPrefixes = locale.inPrefixes.map(toPattern).join('|')
		// years intentionally excluded — "in 3 years" was never supported
		const units = locale.repeatUnits
			.filter(u => u.type !== REPEAT_TYPES.Years)
			.flatMap(u => u.forms)
			.map(toPattern)
			.join('|')
		const regex = new RegExp(`((${inPrefixes}) [0-9]+ (${units}))`, 'ig')
		const results = regex.exec(text)
		if (results === null) {
			continue
		}

		const foundText: string = results[0]
		const date = new Date(now)
		const parts = foundText.split(' ')
		const amount = parseInt(parts[1])
		const unit = findByForm(locale.repeatUnits, parts[2])
		if (unit === undefined) {
			continue
		}

		switch (unit.type) {
			case REPEAT_TYPES.Hours:
				date.setHours(date.getHours() + amount)
				break
			case REPEAT_TYPES.Days:
				date.setDate(date.getDate() + amount)
				break
			case REPEAT_TYPES.Weeks:
				date.setDate(date.getDate() + amount * 7)
				break
			case REPEAT_TYPES.Months:
				date.setMonth(date.getMonth() + amount)
				break
		}

		return {
			foundText,
			date,
		}
	}

	return {
		foundText: '',
		date: null,
	}
}

const getDateFromWeekday = (text: string, date: Date, locale: QuickAddMagicLocale): dateFoundResult => {
	const forms = locale.weekdays.flatMap(w => w.forms).map(toPattern).join('|')
	const prefixes = locale.weekdayPrefixes.map(toPattern).join('|')
	const prefixPart = prefixes === '' ? '' : `(?:(?:${prefixes}) )?`
	const matcher = new RegExp(`(^| )${prefixPart}(${forms})($| )`, 'g')
	const results: RegExpExecArray | null = matcher.exec(text.toLowerCase()) // The i modifier does not seem to work.
	if (results === null) {
		return {
			foundText: null,
			date: null,
		}
	}

	const weekday = findByForm(locale.weekdays, results[2])
	if (weekday === undefined) {
		return {
			foundText: null,
			date: null,
		}
	}

	const distance: number = (weekday.day + 7 - date.getDay()) % 7
	date.setDate(date.getDate() + distance)

	// This a space at the end of the found text to not break parsing suffix strings like "at 14:00" in cases where the
	// matched string comes with a space at the end (last part of the regex).
	let foundText = results[0]
	if (foundText.endsWith(' ')) {
		foundText = foundText.slice(0, foundText.length - 1)
	}

	return {
		foundText: foundText,
		date: date,
	}
}

const getDayFromText = (text: string, now: Date, locale: QuickAddMagicLocale) => {
	// Only match ordinals when followed by end-of-string, time expressions, or month names
	// This prevents matching "2nd Floor" or "13th floor" as dates
	const suffixes = locale.ordinalSuffixes.map(toPattern).join('|')
	const timeLookahead = locale.timePrefixes.map(p => ` ${toPattern(p)} `).join('|')
	const matcher = new RegExp(`(^| )(([1-2][0-9])|(3[01])|(0?[1-9]))(${suffixes})(?=$|${timeLookahead}| ${monthFormsGroup(locale)})`, 'ig')
	const results = matcher.exec(text)
	if (results === null) {
		return {
			foundText: null,
			date: null,
		}
	}

	const date = new Date(now)
	const day = parseInt(results[0])
	date.setDate(day)

	// If the parsed day is the 31st (or 29+ and the next month is february) but the next month only has 30 days,
	// setting the day to 31 will "overflow" the date to the next month, but the first.
	// This would look like a very weired bug. Now, to prevent that, we check if the day is the same as parsed after
	// setting it for the first time and set it again if it isn't - that would mean the month overflowed.
	while (date < now) {
		date.setMonth(date.getMonth() + 1)
	}

	if (date.getDate() !== day) {
		date.setDate(day)
	}

	return {
		foundText: results[0],
		date: date,
	}
}

const getMonthFromText = (text: string, date: Date, locale: QuickAddMagicLocale) => {
	const group = monthFormsGroup(locale)
	// \b only works for Latin scripts; Cyrillic needs explicit space boundaries
	const matcher = locale.monthWordBoundary
		? new RegExp(`\\b${group}\\b`, 'ig')
		: new RegExp(`(^| )${group}($| )`, 'ig')
	const results = matcher.exec(text)

	if (results === null) {
		return {
			newText: text,
			date,
		}
	}

	const monthDef = findByForm(locale.months, results[0].trim())
	if (monthDef !== undefined) {
		date.setMonth(monthDef.month)
	}

	return {
		newText: replaceAll(text, results[0].trim(), ''),
		date,
	}
}

const getDateFromInterval = (interval: number): Date => {
	const newDate = new Date()
	newDate.setDate(newDate.getDate() + interval)
	newDate.setHours(calculateNearestHours(newDate), 0, 0)

	return newDate
}
