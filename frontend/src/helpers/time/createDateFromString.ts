/**
 * Returns a new date from any format in a way that all browsers, especially safari, can understand.
 *
 * @see https://code.vikunja.io/vikunja/issues/207
 *
 * @param dateString
 * @returns {Date}
 */
export function createDateFromString(dateString: string | Date) {
	if (dateString instanceof Date) {
		return dateString
	}

	if (dateString.includes('-')) {
		dateString = dateString.replace(/-/g, '/')
	}

	return new Date(dateString)
}
