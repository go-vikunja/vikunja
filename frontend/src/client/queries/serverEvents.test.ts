import {
	expect,
	it,
	vi,
} from 'vitest'
import {
	QueryClient,
	QueryObserver,
} from '@tanstack/vue-query'
import {
	parseServerCacheEvent,
	serverCacheEventMutationOptions,
} from './serverEvents'
import {
	timeEntryKeys,
	normalizeTimeEntry,
	type TimeEntryPage,
	type TimeEntryResponse,
} from './timeEntries'
import {commentKeys} from './comments'
import {
	taskKeys,
	normalizeTask,
} from './tasks'
import {kanbanKeys} from './kanban'
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

const RUNNING_ENTRY = normalizeTimeEntry({
	id: 4,
	user_id: 7,
	start_time: '2026-09-19T08:00:00Z',
	end_time: null,
})

const TASK_ENTRY = normalizeTimeEntry({
	id: 9,
	user_id: 7,
	task_id: 3,
	start_time: '2026-09-19T08:00:00Z',
	end_time: '2026-09-19T09:00:00Z',
})

function timeEntryPage(items: TimeEntryResponse[]): TimeEntryPage {
	return {
		items,
		page: 1,
		per_page: 50,
		total: items.length,
		total_pages: 1,
	}
}

it('ignores malformed payloads and unrelated notifications', () => {
	expect(parseServerCacheEvent('timer.created', undefined, 7)).toBeNull()
	expect(parseServerCacheEvent('notification.created', {name: 'team.member.added'}, 7)).toBeNull()
})

it('rejects timer entries that do not belong to the current user', () => {
	const entry = {
		id: 4,
		user_id: 8,
		task_id: 1,
	}
	expect(parseServerCacheEvent('timer.created', entry, 7)).toBeNull()
	expect(parseServerCacheEvent('timer.created', entry, undefined)).toBeNull()
	expect(parseServerCacheEvent('timer.created', {
		id: 4,
		task_id: 1,
	}, undefined)).toBeNull()
	expect(parseServerCacheEvent('timer.created', {
		...entry,
		user_id: 7,
	}, 7)).not.toBeNull()
})

it('rejects timer entries with an unusable id or task id', () => {
	expect(parseServerCacheEvent('timer.created', {
		id: 0,
		user_id: 7,
	}, 7)).toBeNull()
	expect(parseServerCacheEvent('timer.created', {
		id: 4,
		user_id: 7,
		task_id: '1',
	}, 7)).toBeNull()
})

it('rejects comment notifications without a usable task id', () => {
	expect(parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {},
	}, 7)).toBeNull()
	expect(parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: 0}},
	}, 7)).toBeNull()
})

it('invalidates only the notified task comments and detail', async () => {
	const client = new QueryClient()
	client.setQueryData(commentKeys.page(1, 'asc', 1, 50), {items: []})
	client.setQueryData(commentKeys.page(2, 'asc', 1, 50), {items: []})
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}))
	const event = parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: 1}},
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(commentKeys.page(1, 'asc', 1, 50))?.isInvalidated).toBe(true)
	expect(client.getQueryState(commentKeys.page(2, 'asc', 1, 50))?.isInvalidated).toBe(false)
	expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(true)
})

it('stales task collections only for a task they hold', async () => {
	const client = new QueryClient()
	const listKey = taskKeys.allList({project: 2})
	const queryFn = vi.fn(() => [normalizeTask({id: 5})])
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	client.setQueryData(taskKeys.detail(5), normalizeTask({id: 5}))
	const commentOn = (taskId: number) => parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: taskId}},
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(commentOn(1))
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(commentOn(5))
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(true)
	expect(client.getQueryState(listKey)?.fetchStatus).toBe('idle')
	expect(queryFn).toHaveBeenCalledTimes(1)
	expect(client.getQueryState(taskKeys.detail(5))?.isInvalidated).toBe(true)
	unsubscribe()
})

it('does not stale task collections for an open task none of them hold', async () => {
	const client = new QueryClient()
	const listKey = taskKeys.allList({project: 2})
	client.setQueryData(listKey, [normalizeTask({id: 5})])
	client.setQueryData(taskKeys.detail(9), normalizeTask({id: 9}))
	const event = parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: 9}},
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	expect(client.getQueryState(taskKeys.detail(9))?.isInvalidated).toBe(true)
})

it('refetches an open list for a created timer without inserting it or refetching the active key', async () => {
	const client = new QueryClient()
	const listKey = timeEntryKeys.list('task_id = 99', 'UTC', 1, 50)
	const queryFn = vi.fn(() => timeEntryPage([]))
	const activeFn = vi.fn(() => null)
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	const unsubscribeActive = new QueryObserver(client, {
		queryKey: timeEntryKeys.active(7),
		queryFn: activeFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	await vi.waitFor(() => expect(activeFn).toHaveBeenCalledTimes(1))
	const event = parseServerCacheEvent('timer.created', RUNNING_ENTRY, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryData(timeEntryKeys.active(7))).toEqual(RUNNING_ENTRY)
	expect(client.getQueryData<TimeEntryPage>(listKey)?.items).toEqual([])
	expect(queryFn).toHaveBeenCalledTimes(2)
	expect(activeFn).toHaveBeenCalledTimes(1)
	unsubscribe()
	unsubscribeActive()
})

it('patches an updated timer into an open list without refetching it', async () => {
	const client = new QueryClient()
	const listKey = timeEntryKeys.list('task_id = 99', 'UTC', 1, 50)
	const queryFn = vi.fn(() => timeEntryPage([RUNNING_ENTRY]))
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	client.setQueryData(timeEntryKeys.active(7), RUNNING_ENTRY)
	const stopped = normalizeTimeEntry({
		...RUNNING_ENTRY,
		end_time: '2026-09-19T09:00:00Z',
	})
	const event = parseServerCacheEvent('timer.updated', stopped, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryData<TimeEntryPage>(listKey)?.items).toEqual([stopped])
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	expect(queryFn).toHaveBeenCalledTimes(1)
	unsubscribe()
})

it('invalidates both tasks when an updated timer moves to another task', async () => {
	const client = new QueryClient()
	client.setQueryData(timeEntryKeys.list('task_id = 3', 'UTC', 1, 50), timeEntryPage([TASK_ENTRY]))
	client.setQueryData(taskKeys.detail(3), normalizeTask({id: 3}))
	client.setQueryData(taskKeys.detail(4), normalizeTask({id: 4}))
	const event = parseServerCacheEvent('timer.updated', {
		...TASK_ENTRY,
		task_id: 4,
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(taskKeys.detail(3))?.isInvalidated).toBe(true)
	expect(client.getQueryState(taskKeys.detail(4))?.isInvalidated).toBe(true)
})

it('invalidates only the timer task when an update keeps it there', async () => {
	const client = new QueryClient()
	client.setQueryData(timeEntryKeys.list('task_id = 3', 'UTC', 1, 50), timeEntryPage([TASK_ENTRY]))
	client.setQueryData(taskKeys.detail(3), normalizeTask({id: 3}))
	client.setQueryData(taskKeys.detail(4), normalizeTask({id: 4}))
	const event = parseServerCacheEvent('timer.updated', {
		...TASK_ENTRY,
		comment: 'changed',
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(taskKeys.detail(3))?.isInvalidated).toBe(true)
	expect(client.getQueryState(taskKeys.detail(4))?.isInvalidated).toBe(false)
})

it('invalidates only the new task when the updated timer is not cached', async () => {
	const client = new QueryClient()
	client.setQueryData(taskKeys.detail(3), normalizeTask({id: 3}))
	client.setQueryData(taskKeys.detail(4), normalizeTask({id: 4}))
	const event = parseServerCacheEvent('timer.updated', {
		...TASK_ENTRY,
		task_id: 4,
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(taskKeys.detail(3))?.isInvalidated).toBe(false)
	expect(client.getQueryState(taskKeys.detail(4))?.isInvalidated).toBe(true)
})

it('removes a deleted timer from an open list without refetching it', async () => {
	const client = new QueryClient()
	const listKey = timeEntryKeys.list('task_id = 99', 'UTC', 1, 50)
	const queryFn = vi.fn(() => timeEntryPage([RUNNING_ENTRY]))
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	client.setQueryData(timeEntryKeys.active(7), RUNNING_ENTRY)
	const event = parseServerCacheEvent('timer.deleted', RUNNING_ENTRY, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryData<TimeEntryPage>(listKey)?.items).toEqual([])
	expect(client.getQueryData<TimeEntryPage>(listKey)?.total).toBe(0)
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	expect(queryFn).toHaveBeenCalledTimes(1)
	unsubscribe()
})

it('refetches only cache families that loaded before the first subscription', async () => {
	const client = new QueryClient()
	const listKey = timeEntryKeys.list('task_id = 99', 'UTC', 1, 50)
	const queryFn = vi.fn(() => timeEntryPage([]))
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	const since = Date.now() + 1000
	const commentKey = commentKeys.page(1, 'asc', 1, 50)
	const boardKey = kanbanKeys.board(1, 1)
	client.setQueryData(commentKey, {items: []}, {updatedAt: since - 1})
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}), {updatedAt: since + 1})
	client.setQueryData(boardKey, {buckets: []}, {updatedAt: since - 1})
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute({
		kind: 'subscribed',
		since,
	})
	expect(queryFn).toHaveBeenCalledTimes(2)
	expect(client.getQueryState(commentKey)?.isInvalidated).toBe(true)
	expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(false)
	expect(client.getQueryState(boardKey)?.isInvalidated).toBe(false)
	unsubscribe()
})

it('leaves a query still loading at the first subscription untouched', async () => {
	const client = new QueryClient()
	const listKey = timeEntryKeys.list('task_id = 1', 'UTC', 1, 50)
	const queryFn = vi.fn(() => new Promise<TimeEntryPage>(() => {}))
	const unsubscribe = new QueryObserver(client, {
		queryKey: listKey,
		queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(queryFn).toHaveBeenCalledTimes(1))
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute({
		kind: 'subscribed',
		since: Date.now() + 1000,
	})
	expect(queryFn).toHaveBeenCalledTimes(1)
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	unsubscribe()
})

it('refreshes task details and marks board membership stale after reconnecting', async () => {
	const client = new QueryClient()
	client.setQueryData(taskKeys.detail(1), normalizeTask({
		id: 1,
		comment_count: 0,
	}))
	const boardKey = kanbanKeys.board(1, 1)
	client.setQueryData(boardKey, {buckets: []})
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute({kind: 'reconnect'})
	expect(client.getQueryState(boardKey)?.isInvalidated).toBe(true)
	expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(true)
})
