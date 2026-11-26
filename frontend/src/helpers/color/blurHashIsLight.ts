import {decode} from 'blurhash'

import {DARK_TEXT_MIN_LUMINANCE, relativeLuminance} from './colorIsDark'

const SIZE = 8

// Only the bottom half counts: that's where the project card puts its title.
export function blurHashIsLight(blurHash: string): boolean {
	let pixels: Uint8ClampedArray
	try {
		pixels = decode(blurHash, SIZE, SIZE)
	} catch {
		return false
	}

	let sum = 0
	const first = SIZE * SIZE / 2
	for (let i = first; i < SIZE * SIZE; i++) {
		sum += relativeLuminance(pixels[i * 4], pixels[i * 4 + 1], pixels[i * 4 + 2])
	}
	return sum / (SIZE * SIZE - first) > DARK_TEXT_MIN_LUMINANCE
}
