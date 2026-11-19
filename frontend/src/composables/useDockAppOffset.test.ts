import {describe, expect, it} from 'vitest'

import {appTopInset, applyDockAppOffset, isMacDockApp} from './useDockAppOffset'

function nav(standalone: boolean | undefined, maxTouchPoints: number) {
	return {
		standalone,
		maxTouchPoints,
	} as unknown as Navigator
}

describe('isMacDockApp', () => {
	it.each([
		['Safari Dock app on a Mac', true, 0, true],
		['normal Safari tab', false, 0, false],
		['installed app on iPhone or iPad', true, 5, false],
		['Chrome or Firefox, installed or not', undefined, 0, false],
	])('%s', (_name, standalone, touchPoints, expected) => {
		expect(isMacDockApp(nav(standalone, touchPoints))).toBe(expected)
	})
})

describe('applyDockAppOffset', () => {
	it('sets the inset in a Mac Dock app', () => {
		const root = document.createElement('div')
		applyDockAppOffset(nav(true, 0), root)
		expect(root.style.getPropertyValue('--app-top-inset')).toBe('40px')
	})

	it('leaves the property unset elsewhere', () => {
		const root = document.createElement('div')
		applyDockAppOffset(nav(false, 0), root)
		expect(root.style.getPropertyValue('--app-top-inset')).toBe('')
		expect(appTopInset(root)).toBe(0)
	})
})
