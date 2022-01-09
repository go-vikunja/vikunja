export function parseDateOrString(rawValue: string, fallback: any) {
	if (typeof rawValue === 'undefined') {
		return fallback
	}

	const d = new Date(rawValue)

	// @ts-ignore if rawValue is an invalid date, isNan will return false.
	return !isNaN(d)
		? d
		: rawValue
}
