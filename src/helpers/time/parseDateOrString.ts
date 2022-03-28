export function parseDateOrString(rawValue: string | undefined, fallback: any): string | Date {
	if (typeof rawValue === 'undefined') {
		return fallback
	}

	const d = new Date(rawValue)

	// @ts-ignore if rawValue is an invalid date, isNan will return false.
	return !isNaN(d)
		? d
		: rawValue
}
