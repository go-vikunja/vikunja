import {describe, it, expect, beforeEach, vi} from 'vitest'
import {setActivePinia, createPinia} from 'pinia'

import {useTimeTrackingStore} from './timeTracking'
import type {ITimeEntry} from '@/modelTypes/ITimeEntry'

const {getAllMock, removeMock, authInfo} = vi.hoisted(() => ({
	getAllMock: vi.fn(),
	removeMock: vi.fn(),
	authInfo: {value: {id: 7} as {id: number} | null},
}))

vi.mock('@/services/timeEntry', async importOriginal => {
	const actual = await importOriginal<typeof import('@/services/timeEntry')>()
	return {
		...actual,
		useTimeEntryService: () => ({
			getAll: getAllMock,
			remove: removeMock,
		}),
	}
})

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		info: authInfo.value,
	}),
}))

function entry(id: number, endTime: Date | null): ITimeEntry {
	return {
		id,
		userId: 1,
		taskId: 1,
		projectId: 0,
		startTime: new Date(),
		endTime,
		comment: '',
		created: new Date(),
		updated: new Date(),
		maxPermission: null,
	}
}

describe('timeTracking store', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		getAllMock.mockReset()
		removeMock.mockReset()
		authInfo.value = {id: 7}
	})

	it('a running entry becomes the active timer', () => {
		const store = useTimeTrackingStore()
		store.applyTimerEvent(entry(4, null))
		expect(store.activeTimer?.id).toBe(4)
		expect(store.hasActiveTimer).toBe(true)
	})

	it('a stopped entry clears the matching active timer', () => {
		const store = useTimeTrackingStore()
		store.applyTimerEvent(entry(4, null))
		store.applyTimerEvent(entry(4, new Date()))
		expect(store.activeTimer).toBeNull()
	})

	it('a stop for a different timer leaves the active one alone', () => {
		const store = useTimeTrackingStore()
		store.applyTimerEvent(entry(4, null))
		store.applyTimerEvent(entry(5, new Date()))
		expect(store.activeTimer?.id).toBe(4)
	})

	it('patches a stopped entry in the loaded list', () => {
		const store = useTimeTrackingStore()
		store.browsedEntries = [entry(4, null), entry(5, null)]
		const stopped = entry(4, new Date('2026-01-01T10:00:00Z'))
		store.applyTimerEvent(stopped)
		expect(store.browsedEntries.find((e: ITimeEntry) => e.id === 4)?.endTime).toEqual(stopped.endTime)
		expect(store.browsedEntries).toHaveLength(2)
	})

	it('does not insert an unknown entry into the loaded list', () => {
		const store = useTimeTrackingStore()
		store.browsedEntries = [entry(4, null)]
		store.applyTimerEvent(entry(9, new Date()))
		expect(store.browsedEntries).toHaveLength(1)
		expect(store.browsedEntries.find((e: ITimeEntry) => e.id === 9)).toBeUndefined()
	})

	it('hydrates the active timer scoped to the current user', async () => {
		getAllMock.mockResolvedValue({items: [entry(4, null)]})

		const store = useTimeTrackingStore()
		await store.hydrateActiveTimer()

		expect(getAllMock).toHaveBeenCalledWith({
			filter: 'user_id = 7 && end_time = null',
			perPage: 1,
		})
		expect(store.activeTimer?.id).toBe(4)
	})

	it('clears the active timer when deleting the running entry', async () => {
		removeMock.mockResolvedValue(undefined)

		const store = useTimeTrackingStore()
		store.browsedEntries = [entry(4, null), entry(5, new Date())]
		store.applyTimerEvent(entry(4, null))

		await store.removeEntry(4)

		expect(removeMock).toHaveBeenCalledWith(4)
		expect(store.browsedEntries.map((e: ITimeEntry) => e.id)).toEqual([5])
		expect(store.activeTimer).toBeNull()
	})

	it('applyTimerDeletion drops the entry and clears the matching active timer', () => {
		const store = useTimeTrackingStore()
		store.browsedEntries = [entry(4, null), entry(5, new Date())]
		store.applyTimerEvent(entry(4, null))

		store.applyTimerDeletion(4)

		expect(store.browsedEntries.map((e: ITimeEntry) => e.id)).toEqual([5])
		expect(store.activeTimer).toBeNull()
	})

	it('applyTimerDeletion of another entry leaves the active timer alone', () => {
		const store = useTimeTrackingStore()
		store.browsedEntries = [entry(4, null), entry(5, new Date())]
		store.applyTimerEvent(entry(4, null))

		store.applyTimerDeletion(5)

		expect(store.browsedEntries.map((e: ITimeEntry) => e.id)).toEqual([4])
		expect(store.activeTimer?.id).toBe(4)
	})
})
