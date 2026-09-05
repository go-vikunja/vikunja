/**
 * Returns a new date from any format in a way that all browsers, especially safari, can understand.
 *
 * @see https://kolaente.dev/vikunja/frontend/issues/207
 *
 * @param dateString
 * @returns {Date}
 */
export function createDateFromString(dateString: string | Date) {
	if (dateString instanceof Date) {
		return dateString
	}

	// Safari won't parse `YYYY-MM-DD HH:mm` but does understand slashes. Full iso strings
	// (the ones with a `T`) parse everywhere and would break when mangled.
	if (!dateString.includes('T') && dateString.includes('-')) {
		dateString = dateString.replace(/-/g, '/')
	}

	return new Date(dateString)
}
