import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'

import {SHORTCUTS} from '@/constants/shortcuts'
import * as appleDevice from '@/helpers/isAppleDevice'
import {parseKey, matchesKey, eventToShortcutString, isFormField, install, uninstall, shortcutBindingToDisplay} from './shortcut'

function makeEvent(overrides: Partial<KeyboardEvent> = {}): KeyboardEvent {
	const event = new KeyboardEvent('keydown')
	const props = {
		key: overrides.code?.startsWith('Key') ? overrides.code.slice(3).toLowerCase() : overrides.code ?? '',
		code: '',
		getModifierState: () => false,
		...overrides,
	}
	for (const [name, value] of Object.entries(props)) {
		Object.defineProperty(event, name, {value})
	}
	return event
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
		expect(eventToShortcutString(event)).toBe('?')
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

describe('keyboard layouts', () => {
	const cases: [string, KeyboardEventInit, string, string][] = [
		['AZERTY A', {
			key: 'a',
			code: 'KeyQ',
		}, 'KeyA', 'KeyQ'],
		['Dvorak T', {
			key: 't',
			code: 'KeyK',
		}, 'KeyT', 'KeyK'],
		['Dvorak Ctrl+K', {
			key: 'k',
			code: 'KeyV',
			ctrlKey: true,
		}, 'Control+KeyK', 'Control+KeyV'],
		['AZERTY Cmd+A', {
			key: 'a',
			code: 'KeyQ',
			metaKey: true,
		}, 'Meta+KeyA', 'Meta+KeyQ'],
		['Russian Ctrl+K', {
			key: 'л',
			code: 'KeyK',
			ctrlKey: true,
		}, 'Control+KeyK', 'KeyK'],
		['Greek T', {
			key: 'τ',
			code: 'KeyT',
		}, 'KeyT', 'KeyR'],
		['Arabic T', {
			key: 'ف',
			code: 'KeyT',
		}, 'KeyT', 'KeyR'],
		['Danish question mark', {
			key: '?',
			code: 'Minus',
			shiftKey: true,
		}, '?', 'Shift+Slash'],
		['US question mark', {
			key: '?',
			code: 'Slash',
			shiftKey: true,
		}, '?', 'Shift+Slash'],
		['unshifted question mark', {
			key: '?',
			code: 'KeyM',
		}, '?', 'KeyM'],
		['AZERTY period', {
			key: '.',
			code: 'Semicolon',
			shiftKey: true,
		}, 'Period', 'Semicolon'],
		['Dvorak Ctrl+period', {
			key: '.',
			code: 'KeyE',
			ctrlKey: true,
		}, 'Control+Period', 'Control+KeyE'],
		['Caps Lock', {
			key: 'T',
			code: 'KeyT',
		}, 'KeyT', 'Shift+KeyT'],
		['explicit Shift', {
			key: 'R',
			code: 'KeyO',
			shiftKey: true,
		}, 'Shift+KeyR', 'KeyR'],
	]

	it.each(cases)('%s matches the same shortcut through both entry points', (_, init, binding, wrongBinding) => {
		const event = new KeyboardEvent('keydown', init)
		expect(eventToShortcutString(event)).toBe(binding)
		expect(matchesKey(event, parseKey(binding))).toBe(true)
		expect(matchesKey(event, parseKey(wrongBinding))).toBe(false)
	})

	it.each(['å', 'é', 'ø', ';', '_'])('does not fall back to the physical letter key for %s', key => {
		const event = new KeyboardEvent('keydown', {
			key,
			code: 'KeyT',
		})
		expect(matchesKey(event, parseKey('KeyT'))).toBe(false)
		expect(eventToShortcutString(event)).not.toBe('KeyT')
	})

	it('does not map a non-Latin letter on a punctuation key to that punctuation', () => {
		const event = new KeyboardEvent('keydown', {
			key: 'ю',
			code: 'Period',
		})
		expect(matchesKey(event, parseKey('Period'))).toBe(false)
	})

	it.each(['Dead', 'Process', 'Unidentified', 'Compose', ''])('ignores %s events', key => {
		const event = new KeyboardEvent('keydown', {
			key,
			code: 'KeyT',
		})
		expect(eventToShortcutString(event)).toBe('')
		expect(matchesKey(event, parseKey('KeyT'))).toBe(false)
	})

	it('ignores IME composition in imperative handlers too', () => {
		const event = new KeyboardEvent('keydown', {
			key: 'k',
			code: 'KeyK',
			ctrlKey: true,
			isComposing: true,
		})
		expect(eventToShortcutString(event)).toBe('')
		expect(matchesKey(event, parseKey('Control+KeyK'))).toBe(false)
	})

	it('ignores legacy IME keydown events before isComposing is set', () => {
		const event = makeEvent({
			key: 'k',
			code: 'KeyK',
			ctrlKey: true,
			keyCode: 229,
		})
		expect(eventToShortcutString(event)).toBe('')
	})

	it('ignores keydown events that are not KeyboardEvents', () => {
		const event = new Event('keydown')
		expect(eventToShortcutString(event)).toBe('')
		expect(matchesKey(event, parseKey('Control+KeyK'))).toBe(false)
	})

	it('does not interpret AltGr text entry as a shortcut', () => {
		const event = makeEvent({
			key: 'л',
			code: 'KeyK',
			ctrlKey: true,
			altKey: true,
			getModifierState: modifier => modifier === 'AltGraph',
		})
		expect(eventToShortcutString(event)).toBe('')
		expect(matchesKey(event, parseKey('Control+Alt+KeyK'))).toBe(false)
	})

	it('only opens help for a question mark without extra command modifiers', () => {
		const binding = parseKey(SHORTCUTS.showKeyboardShortcuts)
		expect(matchesKey(new KeyboardEvent('keydown', {
			key: '?',
			code: 'Minus',
			shiftKey: true,
		}), binding)).toBe(true)
		for (const init of [
			{
				key: '_',
				code: 'Slash',
				shiftKey: true,
			},
			{
				key: '?',
				code: 'Minus',
				shiftKey: true,
				ctrlKey: true,
			},
			{
				key: '?',
				code: 'Minus',
				shiftKey: true,
				altKey: true,
			},
			{
				key: '?',
				code: 'Minus',
				shiftKey: true,
				metaKey: true,
			},
		]) {
			expect(matchesKey(new KeyboardEvent('keydown', init), binding)).toBe(false)
		}
	})
})


describe('shortcutBindingToDisplay', () => {
	it('should convert a simple key binding to display keys', () => {
		expect(shortcutBindingToDisplay('KeyT')).toEqual({
			keys: ['t'],
		})
	})

	it('should convert Mod bindings to ctrl on non-Apple devices', () => {
		const spy = vi.spyOn(appleDevice, 'isAppleDevice').mockReturnValue(false)

		expect(shortcutBindingToDisplay('Mod+KeyE')).toEqual({
			keys: ['ctrl', 'e'],
		})

		spy.mockRestore()
	})

	it('should convert Mod bindings to command on Apple devices', () => {
		const spy = vi.spyOn(appleDevice, 'isAppleDevice').mockReturnValue(true)

		expect(shortcutBindingToDisplay('Mod+KeyE')).toEqual({
			keys: ['⌘', 'e'],
		})

		spy.mockRestore()
	})

	it('should convert Shift+Slash to display keys', () => {
		expect(shortcutBindingToDisplay('Shift+Slash')).toEqual({
			keys: ['shift', '/'],
		})
	})

	it('should convert sequences to display keys with then combination', () => {
		expect(shortcutBindingToDisplay('KeyG KeyO')).toEqual({
			keys: ['g', 'o'],
			combination: 'then',
		})
	})

	it('should convert Alt bindings to display keys', () => {
		expect(shortcutBindingToDisplay('Alt+KeyR')).toEqual({
			keys: ['alt', 'r'],
		})
	})

	it('should convert Backspace to a display key', () => {
		expect(shortcutBindingToDisplay('Backspace')).toEqual({
			keys: ['backspace'],
		})
	})

	it('should convert Delete to a display key', () => {
		expect(shortcutBindingToDisplay('Delete')).toEqual({
			keys: ['delete'],
		})
	})

	it('should convert Period bindings to display keys', () => {
		expect(shortcutBindingToDisplay('Period')).toEqual({
			keys: ['.'],
		})
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

// --- install / uninstall integration tests ---

function fireKeydown(code: string, mods: Partial<Pick<KeyboardEvent, 'ctrlKey' | 'altKey' | 'shiftKey' | 'metaKey'>> = {}) {
	const event = new KeyboardEvent('keydown', {
		code,
		key: code.startsWith('Key') ? code.slice(3).toLowerCase() : code,
		bubbles: true,
		cancelable: true,
		...mods,
	})
	document.dispatchEvent(event)
	return event
}

describe('install / uninstall (sequence handling)', () => {
	let elA: HTMLElement
	let elB: HTMLElement
	let clickA: ReturnType<typeof vi.fn<(event: MouseEvent) => void>>
	let clickB: ReturnType<typeof vi.fn<(event: MouseEvent) => void>>

	beforeEach(() => {
		elA = document.createElement('button')
		elB = document.createElement('button')
		clickA = vi.fn()
		clickB = vi.fn()
		elA.addEventListener('click', clickA)
		elB.addEventListener('click', clickB)
	})

	afterEach(() => {
		uninstall(elA)
		uninstall(elB)
	})

	it.each([
		['Dvorak', 'g', 'KeyU', 'o', 'KeyS'],
		['Russian', 'п', 'KeyG', 'щ', 'KeyO'],
	])('completes navigation sequences on %s', (_, key1, code1, key2, code2) => {
		install(elA, 'KeyG KeyO')
		install(elB, 'KeyO')
		for (const [key, code] of [[key1, code1], [key2, code2]]) {
			document.dispatchEvent(new KeyboardEvent('keydown', {
				key,
				code,
			}))
		}
		expect(clickA).toHaveBeenCalledTimes(1)
		expect(clickB).not.toHaveBeenCalled()
	})

	it('should not fire single-key shortcut while a sequence is in progress', () => {
		// KeyA is standalone, KeyG KeyA is a sequence
		install(elA, 'KeyA')
		install(elB, 'KeyG KeyA')

		fireKeydown('KeyG') // starts sequence
		fireKeydown('KeyA') // should complete sequence, NOT fire standalone

		expect(clickB).toHaveBeenCalledTimes(1)
		expect(clickA).not.toHaveBeenCalled()
	})

	it('should fire single-key shortcut when no sequence is in progress', () => {
		install(elA, 'KeyA')
		install(elB, 'KeyG KeyO')

		fireKeydown('KeyA')

		expect(clickA).toHaveBeenCalledTimes(1)
		expect(clickB).not.toHaveBeenCalled()
	})

	it('should not misfire sequences with different prefixes', () => {
		// KeyG KeyO and KeyH KeyO — pressing KeyH then KeyO should not fire KeyG KeyO
		install(elA, 'KeyG KeyO')
		install(elB, 'KeyH KeyO')

		fireKeydown('KeyH')
		fireKeydown('KeyO')

		expect(clickB).toHaveBeenCalledTimes(1)
		expect(clickA).not.toHaveBeenCalled()
	})

	it('should reset and allow single-key shortcuts after sequence timeout', async () => {
		install(elA, 'KeyA')
		install(elB, 'KeyG KeyA')

		fireKeydown('KeyG') // starts sequence

		// Wait for timeout (1500ms)
		await new Promise(r => setTimeout(r, 1600))

		fireKeydown('KeyA') // sequence timed out, should fire standalone

		expect(clickA).toHaveBeenCalledTimes(1)
		expect(clickB).not.toHaveBeenCalled()
	}, 3000)

	it('should reset when a non-matching key is pressed during a sequence', () => {
		install(elA, 'KeyA')
		install(elB, 'KeyG KeyO')

		fireKeydown('KeyG') // starts sequence
		fireKeydown('KeyX') // doesn't match any active sequence — resets

		// Now KeyA should work as standalone
		fireKeydown('KeyA')

		expect(clickA).toHaveBeenCalledTimes(1)
		expect(clickB).not.toHaveBeenCalled()
	})
})
