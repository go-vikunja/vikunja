import {useMutation} from '@tanstack/vue-query'
import type {TimeEntry} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {
	patchTimeEntry,
	removeTimeEntry,
	timeEntryKeys,
} from './timeEntries'
import {commentKeys} from './comments'
import {taskKeys} from './tasks'
import {invalidateTaskMembership} from './taskCache'

type TimerEntry = TimeEntry & Required<Pick<TimeEntry, 'id' | 'user_id'>>
type TimerEvent = 'timer.created' | 'timer.updated' | 'timer.deleted'
export type ServerCacheEvent =
	| {
		kind: TimerEvent,
		entry: TimerEntry,
	}
	| {
		kind: 'comments',
		taskId: number,
	}
	| {kind: 'reconnect'}

function record(value: unknown): value is Record<string, unknown> {
	return typeof value === 'object' && value !== null
}

export function parseServerCacheEvent(event: string, data: unknown): ServerCacheEvent | null {
	if (!record(data)) return null
	if (event === 'timer.created' || event === 'timer.updated' || event === 'timer.deleted') {
		if (typeof data.id !== 'number' || data.id <= 0 || typeof data.user_id !== 'number') return null
		return {
			kind: event,
			entry: data as TimerEntry,
		}
	}
	if (event === 'notification.created' && data.name === 'task.comment' && record(data.notification)) {
		const task = data.notification.task
		if (record(task) && typeof task.id === 'number' && task.id > 0) return {
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
					invalidateTaskMembership(client, event.entry.task_id),
					...(event.entry.task_id ? [client.invalidateQueries({queryKey: taskKeys.detail(event.entry.task_id)})] : []),
				])
			}
			if (event.kind === 'comments') {
				return Promise.all([
					client.invalidateQueries({queryKey: commentKeys.task(event.taskId)}),
					client.invalidateQueries({queryKey: taskKeys.detail(event.taskId)}),
					invalidateTaskMembership(client, event.taskId),
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
