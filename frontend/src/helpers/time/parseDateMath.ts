const OPERATION = /^(?:([+-])(\d+)([yMwdhHms])|\/([yMwdhHms]))/

function add(date: Date, amount: number, unit: string) {
	switch (unit) {
		case 'y': date.setFullYear(date.getFullYear() + amount); break
		case 'M': date.setMonth(date.getMonth() + amount); break
		case 'w': date.setDate(date.getDate() + amount * 7); break
		case 'd': date.setDate(date.getDate() + amount); break
		case 'h':
		case 'H': date.setHours(date.getHours() + amount); break
		case 'm': date.setMinutes(date.getMinutes() + amount); break
		case 's': date.setSeconds(date.getSeconds() + amount); break
	}
}

function roundDown(date: Date, unit: string) {
	switch (unit) {
		case 'y': date.setMonth(0, 1); date.setHours(0, 0, 0, 0); break
		case 'M': date.setDate(1); date.setHours(0, 0, 0, 0); break
		// Monday, matching go-datemath's default on the backend.
		case 'w': date.setDate(date.getDate() - (date.getDay() + 6) % 7); date.setHours(0, 0, 0, 0); break
		case 'd': date.setHours(0, 0, 0, 0); break
		case 'h':
		case 'H': date.setMinutes(0, 0, 0); break
		case 'm': date.setSeconds(0, 0); break
		case 's': date.setMilliseconds(0); break
	}
}

/**
 * Evaluates datemath like `now/w+1w` or `2024-01-01||+1M`. Returns null when the input isn't datemath.
 */
export function parseDateMath(expression: string, now: Date = new Date()): Date | null {
	const trimmed = expression.trim()

	let date: Date
	let rest: string
	if (trimmed.startsWith('now')) {
		date = new Date(now)
		rest = trimmed.slice(3)
	} else {
		const separator = trimmed.indexOf('||')
		if (separator === -1) {
			return null
		}
		date = new Date(trimmed.slice(0, separator))
		if (isNaN(+date)) {
			return null
		}
		rest = trimmed.slice(separator + 2)
	}

	while (rest !== '') {
		const match = rest.match(OPERATION)
		if (match === null) {
			return null
		}

		const [operation, sign, amount, addUnit, roundUnit] = match
		if (roundUnit) {
			roundDown(date, roundUnit)
		} else {
			add(date, Number(amount) * (sign === '-' ? -1 : 1), addUnit)
		}
		rest = rest.slice(operation.length)
	}

	return date
}
