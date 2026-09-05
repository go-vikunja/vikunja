import {describe, it, expect, beforeAll, beforeEach, vi} from 'vitest'

import {canReloadForChunkLoadError, handleChunkLoadErrors, markChunkLoadErrorReload} from './handleChunkLoadErrors'

describe('canReloadForChunkLoadError', () => {
	beforeEach(() => {
		sessionStorage.clear()
	})

	it('allows a reload when nothing was reloaded yet', () => {
		expect(canReloadForChunkLoadError()).toBe(true)
	})

	it('blocks a reload right after one happened', () => {
		markChunkLoadErrorReload(1_000)

		expect(canReloadForChunkLoadError(2_000)).toBe(false)
	})

	it('allows a reload again after the cooldown passed', () => {
		markChunkLoadErrorReload(1_000)

		expect(canReloadForChunkLoadError(1_000 + 60_000)).toBe(true)
	})

	it('allows a reload when the stored timestamp is garbage', () => {
		sessionStorage.setItem('chunkLoadErrorReloadedAt', 'not a number')

		expect(canReloadForChunkLoadError()).toBe(true)
	})
})

describe('handleChunkLoadErrors', () => {
	const reload = vi.fn()

	beforeAll(() => {
		Object.defineProperty(window, 'location', {
			configurable: true,
			value: {...window.location, reload},
		})
		handleChunkLoadErrors()
	})

	beforeEach(() => {
		sessionStorage.clear()
		reload.mockClear()
	})

	function rejectWith(reason: unknown) {
		const event = new Event('unhandledrejection', {cancelable: true})
		Object.assign(event, {reason})
		window.dispatchEvent(event)
		return event
	}

	it('reloads on a rejected stale chunk import', () => {
		const event = rejectWith(new Error('\'text/html\' is not a valid JavaScript MIME type.'))

		expect(reload).toHaveBeenCalledOnce()
		expect(event.defaultPrevented).toBe(true)
	})

	it('reloads when vite reports a preload error', () => {
		window.dispatchEvent(new Event('vite:preloadError', {cancelable: true}))

		expect(reload).toHaveBeenCalledOnce()
	})

	it('does not reload for an unrelated rejection', () => {
		rejectWith(new Error('something actually broke'))

		expect(reload).not.toHaveBeenCalled()
	})

	it('does not reload for a rejection without a message', () => {
		rejectWith(undefined)

		expect(reload).not.toHaveBeenCalled()
	})

	it('does not reload twice within the cooldown', () => {
		markChunkLoadErrorReload()
		rejectWith(new Error('Importing a module script failed.'))

		expect(reload).not.toHaveBeenCalled()
	})
})
