export function parseDateOrString(rawValue: string | undefined | null, fallback: unknown): (unknown | string | Date) {
	if (rawValue === null || typeof rawValue === 'undefined') {
		return fallback
	}

	if (rawValue.toLowerCase().includes('now') || rawValue.toLowerCase().includes('||')) {
		return rawValue
	}

	const d = new Date(rawValue)

	return !isNaN(+d)
		? d
		: rawValue
}
