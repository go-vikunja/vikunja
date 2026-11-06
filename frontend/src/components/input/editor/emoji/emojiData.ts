export interface EmojiEntry {
	emoji: string
	shortcode: string
	annotation: string
	tags: string[]
}

export interface EmojiCatalogEntry {
	shortcodes: string[]
	annotation: string
	tags?: string[]
	emoji: string
	group: number
	skins?: {
		emoji: string
		tone: number | number[]
	}[]
}

const MAX_RESULTS = 15

let catalogCache: Promise<EmojiCatalogEntry[]> | null = null
let cache: Promise<EmojiEntry[]> | null = null

export function __resetEmojiCacheForTest() {
	catalogCache = null
	cache = null
}

export function loadEmojiCatalog(): Promise<EmojiCatalogEntry[]> {
	if (catalogCache) return catalogCache
	catalogCache = fetch('/emojis.json')
		.then(res => {
			if (!res.ok) throw new Error(`emojis.json HTTP ${res.status}`)
			return res.json() as Promise<EmojiCatalogEntry[]>
		})
		.catch(err => {
			catalogCache = null
			throw err
		})
	return catalogCache
}

export function loadEmojis(): Promise<EmojiEntry[]> {
	if (cache) return cache
	cache = loadEmojiCatalog()
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

export function filterEmojis(index: EmojiEntry[], rawQuery: string, limit = MAX_RESULTS): EmojiEntry[] {
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
		if (starts.length >= limit) break
	}

	return [...starts, ...contains].slice(0, limit)
}
