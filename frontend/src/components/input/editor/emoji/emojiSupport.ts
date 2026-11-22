// Newest first. The .1 releases are mostly ZWJ sequences that degrade to their parts, so the .0 emoji stands in.
const TEST_EMOJI: [string, number][] = [
	['🫪', 17],
	['🫩', 16],
	['🫨', 15.1],
	['🫠', 14],
	['🥲', 13.1],
	['🥻', 12.1],
	['🥰', 11],
	['🤩', 5],
]

let maxVersion: number | undefined

function pixel(emoji: string, color: string) {
	const ctx = document.createElement('canvas').getContext('2d', {willReadFrequently: true})!
	ctx.textBaseline = 'top'
	ctx.font = '100px "Apple Color Emoji", "Segoe UI Emoji", "Noto Color Emoji", sans-serif'
	ctx.fillStyle = color
	ctx.scale(0.01, 0.01)
	ctx.fillText(emoji, 0, 0)
	return ctx.getImageData(0, 0, 1, 1).data.join()
}

// A color glyph ignores fillStyle; a missing one draws nothing or a monochrome fallback.
function rendersInColor(emoji: string) {
	const black = pixel(emoji, '#000')
	return black === pixel(emoji, '#fff') && !black.startsWith('0,0,0,')
}

// Blocked canvas or no color emoji font: show all, grey boxes beat an empty picker
function detect() {
	try {
		const index = TEST_EMOJI.findIndex(([emoji]) => rendersInColor(emoji))
		// Newest row renders: data may outpace this table, so no limit
		return index > 0 ? TEST_EMOJI[index][1] : Infinity
	} catch {
		return Infinity
	}
}

export function maxSupportedEmojiVersion() {
	maxVersion ??= detect()
	return maxVersion
}
