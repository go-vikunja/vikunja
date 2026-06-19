import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {filterEmojis, __resetEmojiCacheForTest, loadEmojis} from './emojiData'

const fixture = [
	{shortcodes: ['grinning', 'grinning_face'], annotation: 'grinning face', tags: ['face', 'grin'], emoji: '😀'},
	{shortcodes: ['eyes'], annotation: 'eyes', tags: ['look'], emoji: '👀'},
	{shortcodes: ['eyeglasses'], annotation: 'glasses', tags: ['eye'], emoji: '👓'},
	{shortcodes: ['smile'], annotation: 'grinning face with smiling eyes', tags: ['eye', 'smile'], emoji: '😄'},
]

describe('emojiData', () => {
	beforeEach(() => {
		__resetEmojiCacheForTest()
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => fixture,
		}))
	})

	afterEach(() => {
		vi.unstubAllGlobals()
	})

	it('flattens multi-shortcode entries and sorts alphabetically', async () => {
		const idx = await loadEmojis()
		const codes = idx.map(e => e.shortcode)
		expect(codes).toEqual(['eyeglasses', 'eyes', 'grinning', 'grinning_face', 'smile'])
	})

	it('returns [] for empty query', () => {
		expect(filterEmojis([{shortcode: 'eyes', emoji: '👀', annotation: '', tags: []}], '')).toEqual([])
	})

	it('prefers startsWith matches over substring matches', () => {
		const loaded = [
			{shortcode: 'eyeglasses', emoji: '👓', annotation: 'glasses', tags: ['eye']},
			{shortcode: 'eyes', emoji: '👀', annotation: 'eyes', tags: []},
			{shortcode: 'smile', emoji: '😄', annotation: 'grinning face with smiling eyes', tags: ['eye']},
		]
		const result = filterEmojis(loaded, 'eye')
		expect(result[0].shortcode).toBe('eyeglasses')
		expect(result[1].shortcode).toBe('eyes')
		expect(result[2].shortcode).toBe('smile')
	})

	it('limits results to 15', () => {
		const big = Array.from({length: 100}, (_, i) => ({
			shortcode: `foo_${String(i).padStart(3, '0')}`, emoji: '✨', annotation: '', tags: [],
		}))
		expect(filterEmojis(big, 'foo')).toHaveLength(15)
	})

	it('caches the fetch promise across calls', async () => {
		await loadEmojis()
		await loadEmojis()
		expect((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls).toHaveLength(1)
	})
})
