// Only used for short previews, so bound the parsing cost of huge descriptions.
const MAX_INPUT_LENGTH = 4096

export function htmlToPreviewText(html: string): string {
	if (!html) {
		return ''
	}
	const doc = new DOMParser().parseFromString(html.slice(0, MAX_INPUT_LENGTH), 'text/html')
	return doc.body.textContent ?? ''
}
