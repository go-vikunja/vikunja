import {effectScope, ref} from 'vue'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {useDelayedLoading} from './useDelayedLoading'
import {LOADING_TIMEOUT} from '@/stores/helper'

describe('useDelayedLoading', () => {
	beforeEach(() => vi.useFakeTimers())
	afterEach(() => vi.useRealTimers())

	it('stays false when loading ends before the timeout', () => {
		const source = ref(false)
		const scope = effectScope()
		const loading = scope.run(() => useDelayedLoading(source))!

		source.value = true
		vi.advanceTimersByTime(LOADING_TIMEOUT - 1)
		source.value = false
		vi.advanceTimersByTime(LOADING_TIMEOUT)

		expect(loading.value).toBe(false)
		scope.stop()
	})

	it('turns on after the timeout and off immediately', () => {
		const source = ref(false)
		const scope = effectScope()
		const loading = scope.run(() => useDelayedLoading(source))!

		source.value = true
		vi.advanceTimersByTime(LOADING_TIMEOUT)
		expect(loading.value).toBe(true)

		source.value = false
		expect(loading.value).toBe(false)
		scope.stop()
	})

	it('delays a source that starts out loading', () => {
		const scope = effectScope()
		const loading = scope.run(() => useDelayedLoading(() => true))!

		expect(loading.value).toBe(false)
		vi.advanceTimersByTime(LOADING_TIMEOUT)
		expect(loading.value).toBe(true)
		scope.stop()
	})

	it('drops the pending timer when the scope is disposed', () => {
		const source = ref(true)
		const scope = effectScope()
		const loading = scope.run(() => useDelayedLoading(source))!

		scope.stop()
		vi.advanceTimersByTime(LOADING_TIMEOUT)

		expect(loading.value).toBe(false)
	})
})
