/**
 * Make date objects from timestamps
 */
export function parseDateOrNull(date: string | Date) {
	if (date instanceof Date) {
		return date
	}

	if ((typeof date === 'string') && !date.startsWith('0001')) {
		return new Date(date)
	}

	return null
}
