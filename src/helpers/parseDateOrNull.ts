/**
 * Make date objects from timestamps
 */
export function parseDateOrNull(date) {
	if (date instanceof Date) {
		return date
	}

	if ((typeof date === 'string' || date instanceof String) && !date.startsWith('0001')) {
		return new Date(date)
	}

	return null
}
