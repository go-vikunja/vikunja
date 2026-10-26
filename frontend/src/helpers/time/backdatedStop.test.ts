import {describe, it, expect} from 'vitest'

import {backdatedStop} from './backdatedStop'

describe('backdatedStop', () => {
	const runningStart = new Date('2026-06-07T09:00:00Z')
	const now = new Date('2026-06-07T10:15:00Z')

	it('stops the running timer at a start between its start and now', () => {
		const start = new Date('2026-06-07T10:00:00Z')
		expect(backdatedStop(runningStart, start, now)).toEqual(start)
	})

	it('returns null for a start before the running timer started', () => {
		expect(backdatedStop(runningStart, new Date('2026-06-07T08:00:00Z'), now)).toBeNull()
	})

	it('returns null for a start equal to the running timer start', () => {
		expect(backdatedStop(runningStart, runningStart, now)).toBeNull()
	})

	it('returns null for a start in the future', () => {
		expect(backdatedStop(runningStart, new Date('2026-06-07T11:00:00Z'), now)).toBeNull()
	})
})
