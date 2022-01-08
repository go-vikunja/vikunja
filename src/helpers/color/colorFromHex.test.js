import {test, expect} from 'vitest'

import {colorFromHex} from './colorFromHex'

test('hex', () => {
	const color = '#ffffff'
	expect(colorFromHex(color)).toBe('ffffff')
})

test('no hex', () => {
	const color = 'ffffff'
	expect(colorFromHex(color)).toBe('ffffff')
})
