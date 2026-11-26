import {test, expect} from 'vitest'
import {encode} from 'blurhash'

import {blurHashIsLight} from './blurHashIsLight'

function hashOf(top: number, bottom: number) {
	const size = 16
	const pixels = new Uint8ClampedArray(size * size * 4)
	for (let i = 0; i < size * size; i++) {
		pixels.fill(i < size * size / 2 ? top : bottom, i * 4, i * 4 + 3)
		pixels[i * 4 + 3] = 255
	}
	return encode(pixels, size, size, 4, 4)
}

test('light image', () => {
	expect(blurHashIsLight(hashOf(240, 240))).toBe(true)
})

test('dark image', () => {
	expect(blurHashIsLight(hashOf(20, 20))).toBe(false)
})

test('only the bottom half counts', () => {
	expect(blurHashIsLight(hashOf(20, 240))).toBe(true)
	expect(blurHashIsLight(hashOf(240, 20))).toBe(false)
})

test('invalid hash falls back to dark', () => {
	expect(blurHashIsLight('nope')).toBe(false)
})
