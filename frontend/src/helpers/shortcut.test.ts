import {describe, it, expect, vi} from 'vitest'

import * as appleDevice from '@/helpers/isAppleDevice'
import {parseKey, matchesKey, eventToShortcutString, isFormField} from './shortcut'

// Helper to create a partial KeyboardEvent with sensible defaults
function makeEvent(overrides: Partial<KeyboardEvent> = {}): KeyboardEvent {
	return {
		key: '',
		code: '',
		ctrlKey: false,
		altKey: false,
		shiftKey: false,
		metaKey: false,
		...overrides,
	} as KeyboardEvent
}

// --- parseKey ---

describe('parseKey', () => {
	it('should parse a simple key code', () => {
		const result = parseKey('KeyT')
		expect(result).toEqual({
			code: 'KeyT',
			ctrl: false,
			alt: false,
			shift: false,
			meta: false,
			mod: false,
		})
	})

	it('should parse Escape', () => {
		const result = parseKey('Escape')
		expect(result.code).toBe('Escape')
		expect(result.ctrl).toBe(false)
	})

	it('should parse Control modifier', () => {
		const result = parseKey('Control+KeyS')
		expect(result.code).toBe('KeyS')
		expect(result.ctrl).toBe(true)
		expect(result.alt).toBe(false)
		expect(result.shift).toBe(false)
		expect(result.meta).toBe(false)
		expect(result.mod).toBe(false)
	})

	it('should parse Meta modifier', () => {
		const result = parseKey('Meta+KeyK')
		expect(result.code).toBe('KeyK')
		expect(result.meta).toBe(true)
		expect(result.ctrl).toBe(false)
	})

	it('should parse Mod modifier', () => {
		const result = parseKey('Mod+KeyE')
		expect(result.code).toBe('KeyE')
		expect(result.mod).toBe(true)
		expect(result.ctrl).toBe(false)
		expect(result.meta).toBe(false)
	})

	it('should parse Shift modifier', () => {
		const result = parseKey('Shift+Delete')
		expect(result.code).toBe('Delete')
		expect(result.shift).toBe(true)
	})

	it('should parse Alt modifier', () => {
		const result = parseKey('Alt+KeyR')
		expect(result.code).toBe('KeyR')
		expect(result.alt).toBe(true)
	})

	it('should parse multiple modifiers', () => {
		const result = parseKey('Control+Shift+KeyA')
		expect(result.code).toBe('KeyA')
		expect(result.ctrl).toBe(true)
		expect(result.shift).toBe(true)
		expect(result.alt).toBe(false)
		expect(result.meta).toBe(false)
	})

	it('should be case-insensitive for modifier names', () => {
		const result = parseKey('control+KeyS')
		expect(result.ctrl).toBe(true)
		expect(result.code).toBe('KeyS')
	})

	it('should parse Shift+Slash', () => {
		const result = parseKey('Shift+Slash')
		expect(result.code).toBe('Slash')
		expect(result.shift).toBe(true)
	})

	it('should parse Period', () => {
		const result = parseKey('Period')
		expect(result.code).toBe('Period')
		expect(result.ctrl).toBe(false)
	})

	it('should parse Control+Period', () => {
		const result = parseKey('Control+Period')
		expect(result.code).toBe('Period')
		expect(result.ctrl).toBe(true)
	})

	it('should handle empty string', () => {
		const result = parseKey('')
		expect(result.code).toBe('')
	})
})

// --- matchesKey ---

describe('matchesKey', () => {
	it('should match a simple key', () => {
		const parsed = parseKey('KeyT')
		const event = makeEvent({code: 'KeyT'})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should not match when code differs', () => {
		const parsed = parseKey('KeyT')
		const event = makeEvent({code: 'KeyS'})
		expect(matchesKey(event, parsed)).toBe(false)
	})

	it('should not match when modifier is pressed but not expected', () => {
		const parsed = parseKey('KeyT')
		const event = makeEvent({code: 'KeyT', ctrlKey: true})
		expect(matchesKey(event, parsed)).toBe(false)
	})

	it('should not match when modifier is expected but not pressed', () => {
		const parsed = parseKey('Control+KeyS')
		const event = makeEvent({code: 'KeyS'})
		expect(matchesKey(event, parsed)).toBe(false)
	})

	it('should match Control modifier', () => {
		const parsed = parseKey('Control+KeyS')
		const event = makeEvent({code: 'KeyS', ctrlKey: true})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should match Meta modifier', () => {
		const parsed = parseKey('Meta+KeyK')
		const event = makeEvent({code: 'KeyK', metaKey: true})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should match Shift modifier', () => {
		const parsed = parseKey('Shift+Delete')
		const event = makeEvent({code: 'Delete', shiftKey: true})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should match Alt modifier', () => {
		const parsed = parseKey('Alt+KeyR')
		const event = makeEvent({code: 'KeyR', altKey: true})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should match multiple modifiers', () => {
		const parsed = parseKey('Control+Shift+KeyA')
		const event = makeEvent({code: 'KeyA', ctrlKey: true, shiftKey: true})
		expect(matchesKey(event, parsed)).toBe(true)
	})

	it('should not match if extra modifier is pressed', () => {
		const parsed = parseKey('Control+KeyS')
		const event = makeEvent({code: 'KeyS', ctrlKey: true, shiftKey: true})
		expect(matchesKey(event, parsed)).toBe(false)
	})

	describe('Mod modifier (platform-adaptive)', () => {
		it('should match Mod as Control on non-Apple devices', () => {
			const spy = vi.spyOn(appleDevice, 'isAppleDevice').mockReturnValue(false)

			const parsed = parseKey('Mod+KeyE')
			const event = makeEvent({code: 'KeyE', ctrlKey: true})
			expect(matchesKey(event, parsed)).toBe(true)

			const eventMeta = makeEvent({code: 'KeyE', metaKey: true})
			expect(matchesKey(eventMeta, parsed)).toBe(false)

			spy.mockRestore()
		})

		it('should match Mod as Meta on Apple devices', () => {
			const spy = vi.spyOn(appleDevice, 'isAppleDevice').mockReturnValue(true)

			const parsed = parseKey('Mod+KeyE')
			const eventMeta = makeEvent({code: 'KeyE', metaKey: true})
			expect(matchesKey(eventMeta, parsed)).toBe(true)

			const eventCtrl = makeEvent({code: 'KeyE', ctrlKey: true})
			expect(matchesKey(eventCtrl, parsed)).toBe(false)

			spy.mockRestore()
		})
	})
})

// --- eventToShortcutString ---

describe('eventToShortcutString', () => {
	it('should return simple key code for plain keypress', () => {
		const event = makeEvent({key: 't', code: 'KeyT'})
		expect(eventToShortcutString(event)).toBe('KeyT')
	})

	it('should include Control modifier', () => {
		const event = makeEvent({key: 'k', code: 'KeyK', ctrlKey: true})
		expect(eventToShortcutString(event)).toBe('Control+KeyK')
	})

	it('should include Meta modifier', () => {
		const event = makeEvent({key: 'k', code: 'KeyK', metaKey: true})
		expect(eventToShortcutString(event)).toBe('Meta+KeyK')
	})

	it('should include Shift modifier', () => {
		const event = makeEvent({key: 'Delete', code: 'Delete', shiftKey: true})
		expect(eventToShortcutString(event)).toBe('Shift+Delete')
	})

	it('should include Alt modifier', () => {
		const event = makeEvent({key: 'r', code: 'KeyR', altKey: true})
		expect(eventToShortcutString(event)).toBe('Alt+KeyR')
	})

	it('should include multiple modifiers in order', () => {
		const event = makeEvent({key: 'a', code: 'KeyA', ctrlKey: true, shiftKey: true})
		expect(eventToShortcutString(event)).toBe('Control+Shift+KeyA')
	})

	it('should return empty string for modifier-only keys', () => {
		expect(eventToShortcutString(makeEvent({key: 'Control', code: 'ControlLeft', ctrlKey: true}))).toBe('')
		expect(eventToShortcutString(makeEvent({key: 'Shift', code: 'ShiftLeft', shiftKey: true}))).toBe('')
		expect(eventToShortcutString(makeEvent({key: 'Alt', code: 'AltLeft', altKey: true}))).toBe('')
		expect(eventToShortcutString(makeEvent({key: 'Meta', code: 'MetaLeft', metaKey: true}))).toBe('')
	})

	it('should handle Escape', () => {
		const event = makeEvent({key: 'Escape', code: 'Escape'})
		expect(eventToShortcutString(event)).toBe('Escape')
	})

	it('should handle Enter', () => {
		const event = makeEvent({key: 'Enter', code: 'Enter'})
		expect(eventToShortcutString(event)).toBe('Enter')
	})

	it('should handle Period', () => {
		const event = makeEvent({key: '.', code: 'Period'})
		expect(eventToShortcutString(event)).toBe('Period')
	})

	it('should handle Control+Period', () => {
		const event = makeEvent({key: '.', code: 'Period', ctrlKey: true})
		expect(eventToShortcutString(event)).toBe('Control+Period')
	})

	it('should handle Slash with Shift (question mark)', () => {
		const event = makeEvent({key: '?', code: 'Slash', shiftKey: true})
		expect(eventToShortcutString(event)).toBe('Shift+Slash')
	})

	it('should produce correct code on non-Latin layouts', () => {
		// On a Russian keyboard, pressing the physical K key produces 'л'
		// but event.code is still 'KeyK'
		const event = makeEvent({key: 'л', code: 'KeyK'})
		expect(eventToShortcutString(event)).toBe('KeyK')
	})

	it('should handle all modifier combinations', () => {
		const event = makeEvent({
			key: 'a',
			code: 'KeyA',
			ctrlKey: true,
			altKey: true,
			shiftKey: true,
			metaKey: true,
		})
		expect(eventToShortcutString(event)).toBe('Control+Alt+Shift+Meta+KeyA')
	})
})

// --- isFormField ---

describe('isFormField', () => {
	it('should return true for input elements', () => {
		const input = document.createElement('input')
		expect(isFormField(input)).toBe(true)
	})

	it('should return true for textarea elements', () => {
		const textarea = document.createElement('textarea')
		expect(isFormField(textarea)).toBe(true)
	})

	it('should return true for select elements', () => {
		const select = document.createElement('select')
		expect(isFormField(select)).toBe(true)
	})

	it('should return true for contentEditable elements', () => {
		const div = document.createElement('div')
		div.contentEditable = 'true'
		expect(isFormField(div)).toBe(true)
	})

	it('should return false for regular div', () => {
		const div = document.createElement('div')
		expect(isFormField(div)).toBe(false)
	})

	it('should return false for button', () => {
		const button = document.createElement('button')
		expect(isFormField(button)).toBe(false)
	})

	it('should return false for null', () => {
		expect(isFormField(null)).toBe(false)
	})
})
