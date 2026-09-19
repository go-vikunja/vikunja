import {useMutation} from '@tanstack/vue-query'
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
import {invalidateTaskMembership} from './taskCache'

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
	event: string,
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

export function serverCacheEventMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (event: ServerCacheEvent) => event,
		onSettled: (event, client) => {
			if ('entry' in event) {
				if (event.kind === 'timer.deleted') removeTimeEntry(client, event.entry.id)
				else patchTimeEntry(client, event.entry)
				return Promise.all([
					client.invalidateQueries({queryKey: timeEntryKeys.all}),
					invalidateTaskMembership(client, event.entry.task_id || undefined, 'active'),
				])
			}
			if (event.kind === 'comments') {
				return Promise.all([
					client.invalidateQueries({queryKey: commentKeys.task(event.taskId)}),
					invalidateTaskMembership(client, event.taskId, 'active'),
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
