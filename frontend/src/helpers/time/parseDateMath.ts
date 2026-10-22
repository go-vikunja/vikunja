// Mirrors github.com/jszwedko/go-datemath as the API evaluates it: Monday week start, fiscal year starting
// Jan 1, no round-up. Evaluated in the browser's timezone.

const UNIT = 'fy|fQ|[yQMwdbhHms]'
const MATH = new RegExp(`^(?:([+-])(\\d*)(${UNIT})|/(${UNIT}))`)
const MILLI_TIMESTAMP = /^\d{5,}/
const DATE = /^(\d{4})(?:-(\d{2})(?:-(\d{2}))?)?/
const TIME = /^(\d{2})(?::(\d{2})(?::(\d{2})(?:\.(\d{1,3}))?)?)?/
const TIMEZONE = /^(?:Z|([+-])(\d{2}):(\d{2}))/

const HOUR = 60 * 60 * 1000
const MINUTE = 60 * 1000
const SECOND = 1000

interface WallTime {
	year: number
	month: number
	day: number
	hours: number
	minutes: number
	seconds: number
	milliseconds: number
}

// Date's constructors map years 0-99 to 1900-1999, the setters don't.
function localDate(t: WallTime): Date {
	const date = new Date(0)
	date.setFullYear(t.year, t.month - 1, t.day)
	date.setHours(t.hours, t.minutes, t.seconds, t.milliseconds)
	return date
}

function zonedDate(t: WallTime, offsetMinutes: number): Date {
	const date = new Date(0)
	date.setUTCFullYear(t.year, t.month - 1, t.day)
	date.setUTCHours(t.hours, t.minutes - offsetMinutes, t.seconds, t.milliseconds)
	return date
}

function parseTime(input: string): {time: Pick<WallTime, 'hours' | 'minutes' | 'seconds' | 'milliseconds'>, length: number} | null {
	const match = input.match(TIME)
	if (match === null) {
		return null
	}

	const [all, hours, minutes = '0', seconds = '0', fraction = '0'] = match
	const time = {
		hours: Number(hours),
		minutes: Number(minutes),
		seconds: Number(seconds),
		milliseconds: Number(fraction.padEnd(3, '0')),
	}
	if (time.hours > 23 || time.minutes > 59 || time.seconds > 59) {
		return null
	}

	return {time, length: all.length}
}

function parseAbsolute(input: string): {date: Date, rest: string} | null {
	const timestamp = input.match(MILLI_TIMESTAMP)
	if (timestamp !== null) {
		return {date: new Date(Number(timestamp[0])), rest: input.slice(timestamp[0].length)}
	}

	const dateMatch = input.match(DATE)
	if (dateMatch === null) {
		const time = parseTime(input)
		return time && {
			date: localDate({year: 1970, month: 1, day: 1, ...time.time}),
			rest: input.slice(time.length),
		}
	}

	const [dateString, year, month = '1', day = '1'] = dateMatch
	const wallTime: WallTime = {
		year: Number(year),
		month: Number(month),
		day: Number(day),
		hours: 0,
		minutes: 0,
		seconds: 0,
		milliseconds: 0,
	}
	if (wallTime.month > 12 || wallTime.day > new Date(wallTime.year, wallTime.month, 0).getDate()) {
		return null
	}

	let rest = input.slice(dateString.length)
	if (!rest.startsWith('T')) {
		return {date: localDate(wallTime), rest}
	}

	const time = parseTime(rest.slice(1))
	if (time === null) {
		return null
	}
	Object.assign(wallTime, time.time)
	rest = rest.slice(1 + time.length)

	const timezone = rest.match(TIMEZONE)
	if (timezone === null) {
		return {date: localDate(wallTime), rest}
	}

	const [zone, sign, offsetHours, offsetMinutes] = timezone
	const offset = zone === 'Z' ? 0 : (sign === '-' ? -1 : 1) * (Number(offsetHours) * 60 + Number(offsetMinutes))
	return {date: zonedDate(wallTime, offset), rest: rest.slice(zone.length)}
}

function isBusinessDay(date: Date) {
	return date.getDay() !== 0 && date.getDay() !== 6
}

function add(date: Date, amount: number, unit: string): Date {
	const result = new Date(date)
	switch (unit) {
		case 'y':
		case 'fy':
			result.setFullYear(result.getFullYear() + amount)
			return result
		case 'Q':
		case 'fQ':
			result.setMonth(result.getMonth() + 3 * amount)
			return result
		case 'M':
			result.setMonth(result.getMonth() + amount)
			return result
		case 'w':
			result.setDate(result.getDate() + 7 * amount)
			return result
		case 'd':
			result.setDate(result.getDate() + amount)
			return result
		case 'b': {
			const step = amount < 0 ? -1 : 1
			for (let left = Math.abs(amount); left > 0; left--) {
				do {
					result.setDate(result.getDate() + step)
				} while (!isBusinessDay(result))
			}
			return result
		}
		// Absolute durations, not wall clock: differs across DST changes.
		case 'h':
		case 'H':
			return new Date(date.getTime() + amount * HOUR)
		case 'm':
			return new Date(date.getTime() + amount * MINUTE)
		default:
			return new Date(date.getTime() + amount * SECOND)
	}
}

function roundDown(date: Date, unit: string): Date | null {
	const result = new Date(date)
	switch (unit) {
		case 'y':
		case 'fy':
			result.setMonth(0, 1)
			break
		case 'Q':
		case 'fQ':
			result.setMonth(Math.floor(result.getMonth() / 3) * 3, 1)
			break
		case 'M':
			result.setDate(1)
			break
		case 'w':
			result.setDate(result.getDate() - (result.getDay() + 6) % 7)
			break
		case 'd':
			break
		// go-datemath truncates these on absolute time, so in a zone like +05:30 `/h` lands on a half hour.
		case 'h':
		case 'H':
			return new Date(Math.floor(date.getTime() / HOUR) * HOUR)
		case 'm':
			return new Date(Math.floor(date.getTime() / MINUTE) * MINUTE)
		case 's':
			return new Date(Math.floor(date.getTime() / SECOND) * SECOND)
		// `/b` panics in go-datemath.
		default:
			return null
	}
	result.setHours(0, 0, 0, 0)
	return result
}

/**
 * Evaluates datemath like `now/w+1w` or `2024-01-01||+1M`. Returns null when the input isn't valid datemath.
 */
export function parseDateMath(expression: string, now: Date = new Date()): Date | null {
	const input = expression.trim()

	let date: Date | null
	let rest: string
	if (input.startsWith('now')) {
		date = new Date(now)
		rest = input.slice(3)
	} else {
		const absolute = parseAbsolute(input)
		if (absolute === null) {
			return null
		}
		date = absolute.date
		rest = absolute.rest
		if (rest === '') {
			return date
		}
		if (!rest.startsWith('||')) {
			return null
		}
		rest = rest.slice(2)
	}

	while (rest !== '') {
		const match = rest.match(MATH)
		if (match === null) {
			return null
		}

		const [operation, sign, amount, addUnit, roundUnit] = match
		date = roundUnit
			? roundDown(date, roundUnit)
			: add(date, (amount === '' ? 1 : Number(amount)) * (sign === '-' ? -1 : 1), addUnit)
		if (date === null) {
			return null
		}
		rest = rest.slice(operation.length)
	}

	return date
}
