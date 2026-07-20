import {test, expect} from 'vitest'

import {getTextColor} from './getTextColor'

// Duplicated from colorIsDark on purpose: an independent oracle, reusing the production math would make the test tautological
function toLinear(c: number) {
	const v = c / 255
	return v <= 0.04045 ? v / 12.92 : Math.pow((v + 0.055) / 1.055, 2.4)
}

function relativeLuminance(hex: string) {
	// expand CSS shorthand (#fff) to full form, since LIGHT/DARK use it
	const full = hex.length === 4
		? `#${[...hex.slice(1)].map(c => c + c).join('')}`
		: hex
	const rgb = parseInt(full.substring(1, 7), 16)
	const r = (rgb >> 16) & 0xff
	const g = (rgb >> 8) & 0xff
	const b = (rgb >> 0) & 0xff
	return 0.2126 * toLinear(r) + 0.7152 * toLinear(g) + 0.0722 * toLinear(b)
}

function contrastRatio(hexA: string, hexB: string) {
	const lA = relativeLuminance(hexA)
	const lB = relativeLuminance(hexB)
	const lighter = Math.max(lA, lB)
	const darker = Math.min(lA, lB)
	return (lighter + 0.05) / (darker + 0.05)
}

const STEPS = [0, 51, 102, 153, 204, 255]

test('every sweep color pairs with a text color clearing 4.5:1', () => {
	const failures: {hex: string, ratio: number}[] = []

	STEPS.forEach(r => {
		STEPS.forEach(g => {
			STEPS.forEach(b => {
				const hex = `#${[r, g, b].map(c => c.toString(16).padStart(2, '0')).join('')}`
				const textColor = getTextColor(hex)
				const ratio = contrastRatio(hex, textColor)

				if (ratio < 4.5) {
					failures.push({hex, ratio})
				}
			})
		})
	})

	expect(failures).toEqual([])
})

test('mid-tone red gets dark text', () => {
	const color = '#e74c3c'
	const textColor = getTextColor(color)
	expect(textColor).toBe('#000')
	expect(contrastRatio(color, textColor)).toBeGreaterThanOrEqual(4.5)
})

test('mid-tone blue gets dark text', () => {
	const color = '#4287f5'
	const textColor = getTextColor(color)
	expect(textColor).toBe('#000')
	expect(contrastRatio(color, textColor)).toBeGreaterThanOrEqual(4.5)
})
