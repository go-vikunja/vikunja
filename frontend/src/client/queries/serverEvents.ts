import {
	useMutation,
	type QueryClient,
	type QueryKey,
} from '@tanstack/vue-query'
import type {TimeEntry} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
	cachedTaskIdOf,
	normalizeTimeEntry,
	patchTimeEntry,
	removeTimeEntry,
	timeEntryKeys,
	type TimeEntryResponse,
} from './timeEntries'
import {commentKeys} from './comments'
import {taskKeys} from './tasks'
import {
	invalidateTaskMembership,
	taskQueryKeys,
} from './taskCache'

export const CACHE_EVENTS = [
	'timer.created',
	'timer.updated',
	'timer.deleted',
	'notification.created',
] as const

export type ServerCacheEventName = typeof CACHE_EVENTS[number]

type TimerEvent = 'timer.created' | 'timer.updated' | 'timer.deleted'
export type ServerCacheEvent =
	| {
		kind: TimerEvent,
		entry: TimeEntryResponse,
	}
	| {
		kind: 'comments',
		taskId: number,
	}
	| {
		kind: 'subscribed',
		since: number,
	}
	| {kind: 'reconnect'}

function isRecord(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null
}

export function parseServerCacheEvent(
	event: ServerCacheEventName,
	data: unknown,
	currentUserId: number | undefined,
): ServerCacheEvent | null {
	if (!isRecord(data)) return null
	if (event === 'timer.created' || event === 'timer.updated' || event === 'timer.deleted') {
		if (currentUserId === undefined || data.user_id !== currentUserId) return null
		if (typeof data.id !== 'number' || data.id <= 0) return null
		if (data.task_id !== undefined && data.task_id !== null && typeof data.task_id !== 'number') return null
		return {
			kind: event,
			entry: normalizeTimeEntry(data as TimeEntry),
		}
	}
	if (event === 'notification.created' && data.name === 'task.comment' && isRecord(data.notification)) {
		const task = data.notification.task
		if (isRecord(task) && typeof task.id === 'number' && task.id > 0) return {
			kind: 'comments',
			taskId: task.id,
		}
	}
	return null
}

const startsWith = (key: QueryKey, prefix: readonly unknown[]) => prefix.every((part, index) => key[index] === part)
const isTaskDetailKey = (key: QueryKey) => startsWith(key, taskKeys.details)

const SUBSCRIBE_SWEEP_KEYS = [
	timeEntryKeys.all,
	commentKeys.all,
	taskKeys.details,
]

// Only collections that already hold the task can have gone stale; its own detail key is invalidated anyway.
async function invalidateCachedTask(client: QueryClient, taskId: number | undefined): Promise<void> {
	if (!taskId) return
	const heldByCollection = taskQueryKeys(client, taskId).some(key => !isTaskDetailKey(key))
	await Promise.all([
		client.invalidateQueries({queryKey: taskKeys.detail(taskId)}),
		...(heldByCollection ? [invalidateTaskMembership(client)] : []),
	])
}

export function serverCacheEventMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (event: ServerCacheEvent) => event,
		onSettled: (event, client) => {
			if ('entry' in event) {
				// The payload carries only the new task, so a reassignment would leave the old one's count stale.
				const previousTaskId = event.kind === 'timer.updated' ? cachedTaskIdOf(client, event.entry.id) : undefined
				if (event.kind === 'timer.deleted') removeTimeEntry(client, event.entry.id)
				else patchTimeEntry(client, event.entry)
				return Promise.all([
					// Whether a new entry belongs to a server-filtered list cannot be decided from the payload.
					...(event.kind === 'timer.created' ? [client.invalidateQueries({queryKey: timeEntryKeys.lists})] : []),
					invalidateCachedTask(client, event.entry.task_id),
					...(previousTaskId === event.entry.task_id ? [] : [invalidateCachedTask(client, previousTaskId)]),
				])
			}
			if (event.kind === 'comments') {
				return Promise.all([
					client.invalidateQueries({queryKey: commentKeys.task(event.taskId)}),
					invalidateCachedTask(client, event.taskId),
				])
			}
			if (event.kind === 'subscribed') {
				// Data that arrived before the subscribe frames went out may have missed an event nobody saw.
				return client.invalidateQueries({
					predicate: query => query.state.dataUpdatedAt > 0
						&& query.state.dataUpdatedAt < event.since
						&& SUBSCRIBE_SWEEP_KEYS.some(prefix => startsWith(query.queryKey, prefix)),
				})
			}
			return Promise.all([
				client.invalidateQueries({queryKey: timeEntryKeys.all}),
				client.invalidateQueries({queryKey: commentKeys.all}),
				client.invalidateQueries({queryKey: taskKeys.details}),
				invalidateTaskMembership(client),
			])
		},
		toastError: () => false,
	})
}

export function useServerCacheEventMutation() {
	return useMutation(serverCacheEventMutationOptions())
}
