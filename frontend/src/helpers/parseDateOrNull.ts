// The api sends this year instead of `null` for "no date set".
const ZERO_TIME_YEAR = 1

/**
 * Single boundary parser for api and user supplied dates. Returns null for everything
 * that isn't a usable date so that no invalid `Date` ever reaches formatting or
 * serialization - `toISOString` and `Intl.DateTimeFormat` throw a RangeError on those.
 */
export function parseDateOrNull(date: Date | string | null | undefined): Date | null {
	const parsed = date instanceof Date
		? date
		: (typeof date === 'string' && date !== '' ? new Date(date) : null)

	if (
		parsed === null ||
		Number.isNaN(parsed.getTime()) ||
		parsed.getUTCFullYear() <= ZERO_TIME_YEAR
	) {
		return null
	}

	return parsed
}
