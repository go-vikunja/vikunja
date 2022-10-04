export function parseDateOrString(rawValue: string | undefined, fallback: unknown) {
	if (typeof rawValue === 'undefined') {
		return fallback
	}

	const d = new Date(rawValue)

	return !isNaN(+d)
		? d
		: rawValue
}
