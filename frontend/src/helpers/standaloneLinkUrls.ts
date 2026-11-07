export function safePreviewUrl(value: string): URL | null {
	try {
		const url = new URL(value)
		return ['http:', 'https:'].includes(url.protocol) && !url.username && !url.password ? url : null
	} catch {
		return null
	}
}

export function standaloneLinkUrls(html: string, internalBases: string[]): string[] {
	const document = new DOMParser().parseFromString(html, 'text/html')
	const bases = internalBases.map(safePreviewUrl).filter((base): base is URL => base !== null)
	const urls = new Set<string>()
	for (const paragraph of document.querySelectorAll('p')) {
		if (paragraph.closest('pre, code')) continue
		const lines = document.createElement('div')
		for (const node of paragraph.childNodes) {
			if (node.nodeName === 'BR') {
				addLine(lines)
				lines.replaceChildren()
			} else {
				lines.appendChild(node.cloneNode(true))
			}
		}
		addLine(lines)
		if (urls.size >= 10) break
	}
	return [...urls].slice(0, 10)

	function addLine(line: HTMLElement) {
		const anchors = line.querySelectorAll('a[href]')
		if (anchors.length !== 1 || line.querySelector('code, img')) return
		const href = anchors[0].getAttribute('href') ?? ''
		if (line.textContent?.trim() !== href || anchors[0].textContent?.trim() !== href) return
		const url = safePreviewUrl(href)
		if (!url) return
		if (bases.some(base => url.origin === base.origin && (
			url.pathname === base.pathname.replace(/\/$/, '') ||
			url.pathname.startsWith(base.pathname.replace(/\/$/, '') + '/')
		))) return
		urls.add(url.href)
	}
}
