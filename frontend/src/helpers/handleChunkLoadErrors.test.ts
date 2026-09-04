import {describe, it, expect, beforeEach} from 'vitest'

import {canReloadForChunkLoadError, markChunkLoadErrorReload} from './handleChunkLoadErrors'

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
