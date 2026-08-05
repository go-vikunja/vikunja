import {describe, expect, it} from 'vitest'

import {runWrites} from './runWrites'

describe('runWrites', () => {
	function deferredWrite() {
		const inFlight: string[] = []
		let maxConcurrent = 0
		const completed: string[] = []
		const write = async (item: string) => {
			inFlight.push(item)
			maxConcurrent = Math.max(maxConcurrent, inFlight.length)
			await Promise.resolve()
			inFlight.splice(inFlight.indexOf(item), 1)
			completed.push(item)
		}
		return {write, completed, getMaxConcurrent: () => maxConcurrent}
	}

	it('runs all writes in parallel when concurrent', async () => {
		const {write, completed, getMaxConcurrent} = deferredWrite()
		await runWrites(['a', 'b', 'c'], write, true)
		expect(completed).toHaveLength(3)
		expect(getMaxConcurrent()).toBeGreaterThan(1)
	})

	it('runs writes one at a time when not concurrent', async () => {
		const {write, completed, getMaxConcurrent} = deferredWrite()
		await runWrites(['a', 'b', 'c'], write, false)
		expect(completed).toEqual(['a', 'b', 'c'])
		expect(getMaxConcurrent()).toBe(1)
	})

	it('does nothing for an empty list', async () => {
		const {write, completed} = deferredWrite()
		await runWrites([], write, false)
		expect(completed).toHaveLength(0)
	})
})
