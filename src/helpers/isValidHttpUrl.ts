export function isValidHttpUrl(urlToCheck: string): boolean {
	let url

	try {
		url = new URL(urlToCheck)
	} catch (_) {
		return false
	}

	return url.protocol === 'http:' || url.protocol === 'https:'
}
