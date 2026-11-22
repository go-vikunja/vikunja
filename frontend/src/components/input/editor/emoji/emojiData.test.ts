import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {filterEmojis, __resetEmojiCacheForTest, loadEmojis, loadEmojiCatalog} from './emojiData'
import {maxSupportedEmojiVersion} from './emojiSupport'

vi.mock('./emojiSupport', () => ({
	maxSupportedEmojiVersion: vi.fn(() => Infinity),
}))

const fixture = [
	{
		shortcodes: ['grinning', 'grinning_face'],
		annotation: 'grinning face',
		tags: ['face', 'grin'],
		emoji: '😀',
		version: 1,
	},
	{
		shortcodes: ['eyes'],
		annotation: 'eyes',
		tags: ['look'],
		emoji: '👀',
		version: 1,
	},
	{
		shortcodes: ['eyeglasses'],
		annotation: 'glasses',
		tags: ['eye'],
		emoji: '👓',
		version: 1,
	},
	{
		shortcodes: ['smile'],
		annotation: 'grinning face with smiling eyes',
		tags: ['eye', 'smile'],
		emoji: '😄',
		version: 1,
	},
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

	it('limits results to 15 by default', () => {
		const big = Array.from({length: 100}, (_, i) => ({
			shortcode: `foo_${String(i).padStart(3, '0')}`, emoji: '✨', annotation: '', tags: [],
		}))
		expect(filterEmojis(big, 'foo')).toHaveLength(15)
		expect(filterEmojis(big, 'foo', Infinity)).toHaveLength(100)
	})

	it('caches the fetch promise across calls', async () => {
		await loadEmojis()
		await loadEmojis()
		expect((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls).toHaveLength(1)
	})

	it('shares one request between catalog and autocomplete', async () => {
		await Promise.all([loadEmojiCatalog(), loadEmojis()])
		expect((globalThis.fetch as ReturnType<typeof vi.fn>).mock.calls).toHaveLength(1)
	})

	it('drops emoji and skin variants newer than the OS renders', async () => {
		vi.mocked(maxSupportedEmojiVersion).mockReturnValueOnce(15)
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue({
			ok: true,
			json: async () => [
				{
					shortcodes: ['leftwards_pushing_hand'],
					annotation: 'leftwards pushing hand',
					emoji: '🫷',
					group: 1,
					version: 15,
					skins: [
						{
							emoji: '🫷🏻',
							tone: 1,
							version: 15,
						},
						{
							emoji: '🫷🏼',
							tone: 2,
							version: 16,
						},
					],
				},
				{
					shortcodes: ['face_with_bags_under_eyes'],
					annotation: 'face with bags under eyes',
					emoji: '🫩',
					group: 0,
					version: 16,
				},
			],
		}))

		expect(await loadEmojiCatalog()).toEqual([
			{
				shortcodes: ['leftwards_pushing_hand'],
				annotation: 'leftwards pushing hand',
				emoji: '🫷',
				group: 1,
				version: 15,
				skins: [
					{
						emoji: '🫷🏻',
						tone: 1,
						version: 15,
					},
				],
			},
		])
		expect((await loadEmojis()).map(e => e.shortcode)).toEqual(['leftwards_pushing_hand'])
	})
})
