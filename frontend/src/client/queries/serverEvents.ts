import {
	useMutation,
	type QueryClient,
	type QueryKey,
} from '@tanstack/vue-query'
import type {TimeEntry} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
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

const isTaskDetailKey = (key: QueryKey) => taskKeys.details.every((part, index) => key[index] === part)

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
				if (event.kind === 'timer.deleted') removeTimeEntry(client, event.entry.id)
				else patchTimeEntry(client, event.entry)
				return Promise.all([
					client.invalidateQueries({queryKey: timeEntryKeys.all}),
					invalidateCachedTask(client, event.entry.task_id),
				])
			}
			if (event.kind === 'comments') {
				return Promise.all([
					client.invalidateQueries({queryKey: commentKeys.task(event.taskId)}),
					invalidateCachedTask(client, event.taskId),
				])
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
