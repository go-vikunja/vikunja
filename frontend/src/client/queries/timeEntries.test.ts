import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	activeTimerQuery,
	timeEntriesQuery,
	timeEntryKeys,
	normalizeTimeEntry,
	stopTimerMutationOptions,
	deleteTimeEntryMutationOptions,
} from './timeEntries'
const sdk = vi.hoisted(() => ({
	timeEntriesList: vi.fn(),
	timeEntriesTimerStop: vi.fn(),
	timeEntriesDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))
let client: QueryClient
const running = {
	id: 4,
	user_id: 7,
	task_id: 1,
	start_time: '2026-09-19T09:00:00Z',
	end_time: null,
}
beforeEach(() => { vi.clearAllMocks(); client = new QueryClient() })

it('hydrates only the current user running timer and keeps ISO dates', async () => {
	sdk.timeEntriesList.mockResolvedValue({data: {items: [running]}})
	const result = await client.fetchQuery(activeTimerQuery(7))
	expect(result?.start_time).toBe(running.start_time)
	expect(sdk.timeEntriesList).toHaveBeenCalledWith({
		query: {
			filter: 'user_id = 7 && end_time = null',
			per_page: 1,
		},
		signal: expect.any(AbortSignal),
	})
})

it('passes browse filters and timezone to the generated client', async () => {
	sdk.timeEntriesList.mockResolvedValue({data: {
		items: [running],
		total_pages: 1,
	}})
	await client.fetchQuery(timeEntriesQuery('task_id = 1', 'Europe/Berlin'))
	expect(sdk.timeEntriesList).toHaveBeenCalledWith({
		query: {
			filter: 'task_id = 1',
			filter_timezone: 'Europe/Berlin',
			per_page: 250,
			page: 1,
		},
		signal: expect.any(AbortSignal),
	})
})

it('patches a stop in loaded lists without inserting into unrelated filters', async () => {
	const list = timeEntryKeys.list('task_id = 1', 'UTC')
	const unrelated = timeEntryKeys.list('task_id = 2', 'UTC')
	client.setQueryData(list, [normalizeTimeEntry(running)])
	client.setQueryData(unrelated, [normalizeTimeEntry({
		...running,
		id: 5,
		task_id: 2,
	})])
	client.setQueryData(timeEntryKeys.active(7), normalizeTimeEntry(running))
	sdk.timeEntriesTimerStop.mockResolvedValue({data: {
		...running,
		end_time: '2026-09-19T10:00:00Z',
	}})
	await client.getMutationCache().build(client, stopTimerMutationOptions()).execute(undefined)
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
	expect(client.getQueryData(list)).toMatchObject([{
		id: 4,
		end_time: '2026-09-19T10:00:00Z',
	}])
	expect(client.getQueryData(unrelated)).toMatchObject([{id: 5}])
})

it('deletes the matching active timer without affecting another user', async () => {
	client.setQueryData(timeEntryKeys.active(7), normalizeTimeEntry(running))
	client.setQueryData(timeEntryKeys.active(8), normalizeTimeEntry({
		...running,
		id: 5,
		user_id: 8,
	}))
	sdk.timeEntriesDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, deleteTimeEntryMutationOptions()).execute({
		id: 4,
		taskId: 1,
	})
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
	expect(client.getQueryData(timeEntryKeys.active(8))).toMatchObject({id: 5})
})
