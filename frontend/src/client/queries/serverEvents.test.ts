import {
	expect,
	it,
	vi,
} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	parseServerCacheEvent,
	serverCacheEventMutationOptions,
} from './serverEvents'
import {
	timeEntryKeys,
	normalizeTimeEntry,
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
	client.setQueryData(commentKeys.page(1, 'asc', 1), {items: []})
	client.setQueryData(commentKeys.page(2, 'asc', 1), {items: []})
	client.setQueryData(taskKeys.detail(1), normalizeTask({id: 1}))
	const event = parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: 1}},
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryState(commentKeys.page(1, 'asc', 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(commentKeys.page(2, 'asc', 1))?.isInvalidated).toBe(false)
	expect(client.getQueryState(taskKeys.detail(1))?.isInvalidated).toBe(true)
})

it('stales task collections only for a task they hold', async () => {
	const client = new QueryClient()
	const listKey = taskKeys.allList({project: 2})
	client.setQueryData(listKey, [normalizeTask({id: 5})])
	const commentOn = (taskId: number) => parseServerCacheEvent('notification.created', {
		name: 'task.comment',
		notification: {task: {id: taskId}},
	}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(commentOn(1))
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(false)
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(commentOn(5))
	expect(client.getQueryState(listKey)?.isInvalidated).toBe(true)
})

it('reconciles timer events without inserting into an unrelated filtered list', async () => {
	const client = new QueryClient()
	const entry = normalizeTimeEntry({
		id: 4,
		user_id: 7,
		start_time: '2026-09-19T08:00:00Z',
		end_time: null,
	})
	client.setQueryData(timeEntryKeys.active(7), null)
	client.setQueryData(timeEntryKeys.list('task_id = 99', 'UTC'), [])
	const event = parseServerCacheEvent('timer.created', entry, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryData(timeEntryKeys.active(7))).toEqual(entry)
	expect(client.getQueryData(timeEntryKeys.list('task_id = 99', 'UTC'))).toEqual([])
	expect(client.getQueryState(timeEntryKeys.list('task_id = 99', 'UTC'))?.isInvalidated).toBe(true)
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute({
		kind: 'timer.deleted',
		entry,
	})
	expect(client.getQueryData(timeEntryKeys.active(7))).toBeNull()
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
