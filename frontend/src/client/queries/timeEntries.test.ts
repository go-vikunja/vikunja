import {
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	activeTimerQuery,
	createTimeEntryMutationOptions,
	timeEntriesQuery,
	timeEntryKeys,
	normalizeTimeEntry,
	stopTimerMutationOptions,
	updateTimeEntryMutationOptions,
	deleteTimeEntryMutationOptions,
	type TimeEntryPage,
	type TimeEntryResponse,
} from './timeEntries'
import type {TimeEntry} from '@/client/generated'
import {
	normalizeTask,
	taskKeys,
	type TaskResponse,
} from './tasks'
const sdk = vi.hoisted(() => ({
	timeEntriesList: vi.fn(),
	timeEntriesCreate: vi.fn(),
	timeEntriesTimerStop: vi.fn(),
	timeEntriesUpdate: vi.fn(),
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
function page(items: TimeEntry[], overrides: Partial<TimeEntryPage> = {}): TimeEntryPage {
	return {
		items: items.map(normalizeTimeEntry),
		page: 1,
		per_page: 50,
		total: items.length,
		total_pages: 1,
		...overrides,
	}
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

it('resolves the active timer to null without a request when there is no user id', async () => {
	const result = await client.fetchQuery(activeTimerQuery(0))
	expect(result).toBeNull()
	expect(sdk.timeEntriesList).not.toHaveBeenCalled()
})

it('requests exactly the asked-for page with the browse filter and timezone', async () => {
	sdk.timeEntriesList.mockResolvedValue({data: {
		items: [running],
		page: 2,
		per_page: 50,
		total: 60,
		total_pages: 2,
	}})
	const page = await client.fetchQuery(timeEntriesQuery('task_id = 1', 'Europe/Berlin', 2, 50))
	expect(sdk.timeEntriesList).toHaveBeenCalledTimes(1)
	expect(sdk.timeEntriesList).toHaveBeenCalledWith({
		query: {
			filter: 'task_id = 1',
			filter_timezone: 'Europe/Berlin',
			page: 2,
			per_page: 50,
		},
		signal: expect.any(AbortSignal),
	})
	expect(page).toEqual({
		items: [{
			...running,
			project_id: 0,
			comment: '',
		}],
		page: 2,
		per_page: 50,
		total: 60,
		total_pages: 2,
	})
})

it('fills in the page envelope with the requested size when the server omits it', async () => {
	sdk.timeEntriesList.mockResolvedValue({data: {items: null}})
	const page = await client.fetchQuery(timeEntriesQuery('', 'UTC', 3, 25))
	expect(page).toEqual({
		items: [],
		page: 3,
		per_page: 25,
		total: 0,
		total_pages: 0,
	})
})

it('clamps per_page to the v2 maximum', async () => {
	sdk.timeEntriesList.mockResolvedValue({data: {
		items: [],
		page: 1,
		per_page: 1000,
		total: 0,
		total_pages: 0,
	}})
	await client.fetchQuery(timeEntriesQuery('', 'UTC', 1, 5000))
	expect(sdk.timeEntriesList).toHaveBeenCalledWith({
		query: {
			filter: '',
			filter_timezone: 'UTC',
			page: 1,
			per_page: 1000,
		},
		signal: expect.any(AbortSignal),
	})
})

it('patches a stop in loaded lists without inserting into unrelated filters', async () => {
	const list = timeEntryKeys.list('task_id = 1', 'UTC', 1, 50)
	const unrelated = timeEntryKeys.list('task_id = 2', 'UTC', 1, 50)
	const taskList = taskKeys.list({project: 1})
	client.setQueryData(list, page([running]))
	client.setQueryData(unrelated, page([{
		...running,
		id: 5,
		task_id: 2,
	}]))
	client.setQueryData(timeEntryKeys.active(7), normalizeTimeEntry(running))
	client.setQueryData(taskList, {items: [], total: 0, total_pages: 0})
	sdk.timeEntriesTimerStop.mockResolvedValue({data: {
		...running,
		end_time: '2026-09-19T10:00:00Z',
	}})
	await client.getMutationCache().build(client, stopTimerMutationOptions()).execute(undefined)
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
	expect(client.getQueryData<TimeEntryPage>(list)?.items).toMatchObject([{
		id: 4,
		end_time: '2026-09-19T10:00:00Z',
	}])
	expect(client.getQueryData<TimeEntryPage>(unrelated)?.items).toMatchObject([{id: 5}])
	expect(client.getQueryState(taskList)?.isInvalidated).toBe(false)
	expect(client.getQueryState(list)?.isInvalidated).toBe(true)
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

it('drops the deleted entry from a loaded list and decrements the task entry count', async () => {
	const list = timeEntryKeys.list('task_id = 1', 'UTC', 1, 50)
	const detail = taskKeys.detail(1)
	client.setQueryData(list, page([
		running,
		{
			...running,
			id: 6,
		},
	], {
		total: 60,
		total_pages: 2,
	}))
	client.setQueryData(detail, normalizeTask({
		id: 1,
		title: 'task',
		project_id: 1,
		time_entries_count: 2,
	}))
	sdk.timeEntriesDelete.mockResolvedValue({data: undefined})
	await client.getMutationCache().build(client, deleteTimeEntryMutationOptions()).execute({
		id: 4,
		taskId: 1,
	})
	expect(client.getQueryData<TimeEntryPage>(list)).toMatchObject({
		items: [{id: 6}],
		total: 59,
		total_pages: 2,
	})
	expect(client.getQueryData<TaskResponse>(detail)?.time_entries_count).toBe(1)
})

it('leaves a loaded list untouched on create and only bumps the task entry count', async () => {
	const list = timeEntryKeys.list('task_id = 1', 'UTC', 1, 50)
	const detail = taskKeys.detail(1)
	client.setQueryData(list, page([running], {total: 7}))
	client.setQueryData(detail, normalizeTask({
		id: 1,
		title: 'task',
		project_id: 1,
		time_entries_count: 1,
	}))
	sdk.timeEntriesCreate.mockResolvedValue({data: {
		...running,
		id: 9,
		end_time: '2026-09-19T10:00:00Z',
	}})
	await client.getMutationCache().build(client, createTimeEntryMutationOptions()).execute({
		task_id: 1,
		start_time: running.start_time,
		end_time: '2026-09-19T10:00:00Z',
	})
	expect(client.getQueryData<TimeEntryPage>(list)).toMatchObject({
		items: [{id: 4}],
		total: 7,
	})
	expect(client.getQueryData<TaskResponse>(detail)?.time_entries_count).toBe(2)
	expect(client.getQueryState(list)?.isInvalidated).toBe(true)
})

function seedTaskMove() {
	client.setQueryData(taskKeys.detail(1), normalizeTask({
		id: 1,
		title: 'from',
		project_id: 1,
		time_entries_count: 2,
	}))
	client.setQueryData(taskKeys.detail(2), normalizeTask({
		id: 2,
		title: 'to',
		project_id: 1,
		time_entries_count: 5,
	}))
	sdk.timeEntriesUpdate.mockResolvedValue({data: {
		...running,
		task_id: 2,
		end_time: '2026-09-19T10:00:00Z',
	}})
}

const moveToTaskTwo = {
	id: 4,
	task_id: 2,
	start_time: running.start_time,
}

it('moves the entry count between both tasks using the task id cached for the entry', async () => {
	client.setQueryData(timeEntryKeys.list('task_id = 1', 'UTC', 1, 50), page([running]))
	seedTaskMove()
	await client.getMutationCache().build(client, updateTimeEntryMutationOptions()).execute(moveToTaskTwo)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(1))?.time_entries_count).toBe(1)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(2))?.time_entries_count).toBe(6)
	expect(client.getQueryState(taskKeys.detail(2))?.isInvalidated).toBe(true)
})

it('moves nothing when a second update repeats a task move the cache already recorded', async () => {
	client.setQueryData(timeEntryKeys.list('task_id = 1', 'UTC', 1, 50), page([running]))
	seedTaskMove()
	const options = updateTimeEntryMutationOptions()
	await client.getMutationCache().build(client, options).execute(moveToTaskTwo)
	await client.getMutationCache().build(client, options).execute(moveToTaskTwo)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(1))?.time_entries_count).toBe(1)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(2))?.time_entries_count).toBe(6)
})

it('moves nothing when no cached copy of the updated entry exists', async () => {
	seedTaskMove()
	await client.getMutationCache().build(client, updateTimeEntryMutationOptions()).execute(moveToTaskTwo)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(1))?.time_entries_count).toBe(2)
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(2))?.time_entries_count).toBe(5)
})

it('leaves the entry count alone when an update keeps the same task', async () => {
	client.setQueryData(timeEntryKeys.list('task_id = 1', 'UTC', 1, 50), page([running]))
	client.setQueryData(taskKeys.detail(1), normalizeTask({
		id: 1,
		title: 'same',
		project_id: 1,
		time_entries_count: 0,
	}))
	sdk.timeEntriesUpdate.mockResolvedValue({data: {
		...running,
		end_time: '2026-09-19T10:00:00Z',
	}})
	await client.getMutationCache().build(client, updateTimeEntryMutationOptions()).execute({
		id: 4,
		task_id: 1,
		start_time: running.start_time,
	})
	// Math.max(0, …) clamp makes an unguarded -1/+1 net to +1 from a 0 seed, unlike the 2 -> 1 -> 2 no-op a nonzero seed would hide.
	expect(client.getQueryData<TaskResponse>(taskKeys.detail(1))?.time_entries_count).toBe(0)
})

it('keeps the active timer cached when a still-running entry is updated', async () => {
	client.setQueryData(timeEntryKeys.active(7), normalizeTimeEntry(running))
	sdk.timeEntriesUpdate.mockResolvedValue({data: {
		...running,
		comment: 'renamed',
	}})
	await client.getMutationCache().build(client, updateTimeEntryMutationOptions()).execute({
		id: 4,
		task_id: 1,
		comment: 'renamed',
		start_time: running.start_time,
	})
	expect(client.getQueryData<TimeEntryResponse | null>(timeEntryKeys.active(7))).toMatchObject({
		id: 4,
		comment: 'renamed',
		end_time: null,
	})
})
