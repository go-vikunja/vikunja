/**
 * Make date objects from timestamps
*/
export function parseDateOrNull(date) {
	if (date instanceof Date) {
		return date
	}

	if (date && !date.startsWith('0001')) {
		return new Date(date)
	}
	return null
}
