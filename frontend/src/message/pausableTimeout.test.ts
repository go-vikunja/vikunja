import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'

import {pausableTimeout} from './pausableTimeout'

describe('pausableTimeout', () => {
	beforeEach(() => {
		vi.useFakeTimers()
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('expires after its duration', () => {
		const onExpire = vi.fn()
		pausableTimeout(1000, onExpire)

		vi.advanceTimersByTime(999)
		expect(onExpire).not.toHaveBeenCalled()
		vi.advanceTimersByTime(1)
		expect(onExpire).toHaveBeenCalledOnce()
	})

	it('pauses while held and resumes with the time that was left', () => {
		const onExpire = vi.fn()
		const timeout = pausableTimeout(1000, onExpire)

		vi.advanceTimersByTime(400)
		timeout.hold('hover')
		vi.advanceTimersByTime(5000)
		expect(onExpire).not.toHaveBeenCalled()

		timeout.release('hover')
		vi.advanceTimersByTime(599)
		expect(onExpire).not.toHaveBeenCalled()
		vi.advanceTimersByTime(1)
		expect(onExpire).toHaveBeenCalledOnce()
	})

	it('stays paused until every hold is released', () => {
		const onExpire = vi.fn()
		const timeout = pausableTimeout(1000, onExpire)

		timeout.hold('hover')
		timeout.hold('focus')
		timeout.hold('focus')
		timeout.release('hover')
		vi.advanceTimersByTime(5000)
		expect(onExpire).not.toHaveBeenCalled()

		timeout.release('focus')
		vi.advanceTimersByTime(1000)
		expect(onExpire).toHaveBeenCalledOnce()
	})

	it('ignores a release without a hold', () => {
		const onExpire = vi.fn()
		const timeout = pausableTimeout(1000, onExpire)

		timeout.release('hover')
		vi.advanceTimersByTime(1000)
		expect(onExpire).toHaveBeenCalledOnce()
	})

	it('never expires once cancelled', () => {
		const onExpire = vi.fn()
		const timeout = pausableTimeout(1000, onExpire)

		timeout.hold('focus')
		timeout.cancel()
		timeout.release('focus')
		vi.advanceTimersByTime(5000)
		expect(onExpire).not.toHaveBeenCalled()
	})
})
