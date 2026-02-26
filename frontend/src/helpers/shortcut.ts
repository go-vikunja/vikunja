import {isAppleDevice} from '@/helpers/isAppleDevice'

// --- Types ---

interface ParsedKey {
	code: string
	ctrl: boolean
	alt: boolean
	shift: boolean
	meta: boolean
	mod: boolean
}

// --- Core functions ---

function parseKey(keyStr: string): ParsedKey {
	const parts = keyStr.split('+')
	const code = parts.pop() || ''
	const modifiers = new Set(parts.map(m => m.toLowerCase()))

	return {
		code,
		ctrl: modifiers.has('control'),
		alt: modifiers.has('alt'),
		shift: modifiers.has('shift'),
		meta: modifiers.has('meta'),
		mod: modifiers.has('mod'),
	}
}

function matchesKey(event: KeyboardEvent, parsed: ParsedKey): boolean {
	if (event.code !== parsed.code) return false

	const isMac = isAppleDevice()

	const wantCtrl = parsed.ctrl || (!isMac && parsed.mod)
	const wantMeta = parsed.meta || (isMac && parsed.mod)

	if (event.ctrlKey !== wantCtrl) return false
	if (event.altKey !== parsed.alt) return false
	if (event.shiftKey !== parsed.shift) return false
	if (event.metaKey !== wantMeta) return false

	return true
}

/**
 * Convert a KeyboardEvent to a normalized shortcut string (event.code-based).
 * Replacement for eventToHotkeyString from @github/hotkey.
 *
 * Examples:
 *   Ctrl+K press -> 'Control+KeyK'
 *   Cmd+K press  -> 'Meta+KeyK'
 *   plain T      -> 'KeyT'
 *   Shift+Delete -> 'Shift+Delete'
 */
export function eventToShortcutString(event: KeyboardEvent): string {
	// Skip modifier-only keys
	if (['Control', 'Alt', 'Shift', 'Meta'].includes(event.key)) {
		return ''
	}

	const parts: string[] = []

	if (event.ctrlKey) parts.push('Control')
	if (event.altKey) parts.push('Alt')
	if (event.shiftKey) parts.push('Shift')
	if (event.metaKey) parts.push('Meta')

	parts.push(event.code)

	return parts.join('+')
}

// --- Form field detection ---

function isFormField(target: EventTarget | null): boolean {
	if (!(target instanceof HTMLElement)) return false

	const tagName = target.tagName.toLowerCase()
	if (tagName === 'input' || tagName === 'textarea' || tagName === 'select') return true
	if (target.contentEditable === 'true') return true

	return false
}

// --- Install / Uninstall ---

const SEQUENCE_TIMEOUT = 1500

interface Binding {
	keys: ParsedKey[][]
	el: HTMLElement
}

const bindings = new Set<Binding>()
let sequenceBuffer: string[] = []
let sequenceTimer: ReturnType<typeof setTimeout> | null = null

function resetSequence() {
	sequenceBuffer = []
	if (sequenceTimer !== null) {
		clearTimeout(sequenceTimer)
		sequenceTimer = null
	}
}

function globalKeydownHandler(event: KeyboardEvent) {
	if (event.defaultPrevented) return
	if (event.isComposing) return
	if (event.repeat) return

	const target = (event as any).explicitOriginalTarget || event.target
	if (target?.shadowRoot) return
	if (isFormField(target)) return

	for (const binding of bindings) {
		for (const sequence of binding.keys) {
			if (sequence.length === 1) {
				// Single-key shortcut
				if (matchesKey(event, sequence[0])) {
					event.preventDefault()
					binding.el.click()
					resetSequence()
					return
				}
			} else {
				// Sequence shortcut (e.g. 'KeyG KeyO')
				const stepIndex = sequenceBuffer.length
				if (stepIndex < sequence.length && matchesKey(event, sequence[stepIndex])) {
					sequenceBuffer.push(event.code)

					if (sequenceTimer !== null) {
						clearTimeout(sequenceTimer)
					}
					sequenceTimer = setTimeout(resetSequence, SEQUENCE_TIMEOUT)

					if (sequenceBuffer.length === sequence.length) {
						event.preventDefault()
						binding.el.click()
						resetSequence()
						return
					}

					// Partial match — consume the event
					event.preventDefault()
					return
				}
			}
		}
	}

	// No match for any sequence step — reset
	if (sequenceBuffer.length > 0) {
		resetSequence()
	}
}

let listenerInstalled = false

function ensureListener() {
	if (!listenerInstalled) {
		document.addEventListener('keydown', globalKeydownHandler)
		listenerInstalled = true
	}
}

function maybeRemoveListener() {
	if (bindings.size === 0 && listenerInstalled) {
		document.removeEventListener('keydown', globalKeydownHandler)
		listenerInstalled = false
	}
}

/**
 * Install a shortcut on an element -- clicking it when shortcut fires.
 * Handles sequences (space-separated keys like 'KeyG KeyO').
 */
export function install(el: HTMLElement, shortcut: string): void {
	const sequences = shortcut.split(' ').reduce<string[][]>((acc, part) => {
		// Each space-separated token is a step in the sequence
		if (acc.length === 0) acc.push([])
		acc[0].push(part)
		return acc
	}, [])

	const keys = sequences.map(seq => seq.map(parseKey))

	const binding: Binding = {keys, el}
	bindings.add(binding)
	;(el as any).__shortcutBinding = binding

	ensureListener()
}

/**
 * Remove an element's shortcut binding.
 */
export function uninstall(el: HTMLElement): void {
	const binding = (el as any).__shortcutBinding as Binding | undefined
	if (binding) {
		bindings.delete(binding)
		delete (el as any).__shortcutBinding
	}

	maybeRemoveListener()
}
