import {test, expect} from 'vitest'

import {colorIsDark} from './colorIsDark'

test('dark color', () => {
	const color = '#111111'
	expect(colorIsDark(color)).toBe(false)
})

test('light color', () => {
	const color = '#DDDDDD'
	expect(colorIsDark(color)).toBe(true)
})

test('default dark', () => {
	const color = ''
	expect(colorIsDark(color)).toBe(true)
})
