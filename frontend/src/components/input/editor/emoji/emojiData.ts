export interface EmojiEntry {
	emoji: string
	shortcode: string
	annotation: string
	tags: string[]
}

interface RawEmoji {
	shortcodes: string[]
	annotation: string
	tags?: string[]
	emoji: string
}

const MAX_RESULTS = 15

let cache: Promise<EmojiEntry[]> | null = null

export function __resetEmojiCacheForTest() {
	cache = null
}

export function loadEmojis(): Promise<EmojiEntry[]> {
	if (cache) return cache
	cache = fetch('/emojis.json')
		.then(res => {
			if (!res.ok) throw new Error(`emojis.json HTTP ${res.status}`)
			return res.json() as Promise<RawEmoji[]>
		})
		.then(raw => {
			const flat: EmojiEntry[] = []
			for (const entry of raw) {
				for (const shortcode of entry.shortcodes) {
					flat.push({
						emoji: entry.emoji,
						shortcode,
						annotation: entry.annotation,
						tags: entry.tags ?? [],
					})
				}
			}
			flat.sort((a, b) => a.shortcode.localeCompare(b.shortcode))
			return flat
		})
		.catch(err => {
			cache = null
			throw err
		})
	return cache
}

export function filterEmojis(index: EmojiEntry[], rawQuery: string): EmojiEntry[] {
	const query = rawQuery.toLowerCase()
	if (query === '') return []

	const starts: EmojiEntry[] = []
	const contains: EmojiEntry[] = []

	for (const entry of index) {
		if (entry.shortcode.startsWith(query)) {
			starts.push(entry)
			continue
		}
		if (
			entry.shortcode.includes(query) ||
			entry.annotation.toLowerCase().includes(query) ||
			entry.tags.some(t => t.toLowerCase().includes(query))
		) {
			contains.push(entry)
		}
		if (starts.length >= MAX_RESULTS) break
	}

	return [...starts, ...contains].slice(0, MAX_RESULTS)
}
