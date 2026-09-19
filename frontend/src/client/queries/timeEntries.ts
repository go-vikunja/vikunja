import {
	queryOptions,
	useMutation,
	type QueryClient,
} from '@tanstack/vue-query'
import {
	timeEntriesCreate,
	timeEntriesDelete,
	timeEntriesList,
	timeEntriesTimerStop,
	timeEntriesUpdate,
} from '@/client/generated'
import type {
	TimeEntry,
	TimeEntryWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'

export type TimeEntryResponse = Omit<TimeEntry, 'id' | 'user_id' | 'task_id' | 'project_id' | 'comment'> &
	Required<Pick<TimeEntry, 'id' | 'user_id' | 'task_id' | 'project_id' | 'comment'>>

export function normalizeTimeEntry(entry: TimeEntry): TimeEntryResponse {
	return {
		...entry,
		id: entry.id ?? 0,
		user_id: entry.user_id ?? 0,
		task_id: entry.task_id ?? 0,
		project_id: entry.project_id ?? 0,
		comment: entry.comment ?? '',
	}
}

export const timeEntryKeys = {
	all: ['time-entries'] as const,
	lists: ['time-entries', 'list'] as const,
	list: (filter: string, timezone: string) => ['time-entries', 'list', filter, timezone] as const,
	activeTimers: ['time-entries', 'active'] as const,
	active: (userId: number) => ['time-entries', 'active', userId] as const,
}

export function timeEntriesQuery(filter: string, timezone: string) {
	return queryOptions({
		queryKey: timeEntryKeys.list(filter, timezone),
		queryFn: async ({signal}) => (await fetchAllPages(page => timeEntriesList({
			query: {
				filter,
				filter_timezone: timezone,
				per_page: 250,
				page,
			},
			signal,
		}).then(({data}) => data))).map(normalizeTimeEntry),
	})
}

export function activeTimerQuery(userId: number) {
	return queryOptions({
		queryKey: timeEntryKeys.active(userId),
		enabled: userId > 0,
		queryFn: async ({signal}) => {
			const {data} = await timeEntriesList({
				query: {
					filter: `user_id = ${userId} && end_time = null`,
					per_page: 1,
				},
				signal,
			})
			return data.items?.[0] ? normalizeTimeEntry(data.items[0]) : null
		},
	})
}

export function patchTimeEntry(client: QueryClient, entry: TimeEntry) {
	const normalized = normalizeTimeEntry(entry)
	client.setQueriesData<TimeEntryResponse[]>({queryKey: timeEntryKeys.lists}, current =>
		current?.some(item => item.id === entry.id)
			? current.map(item => item.id === entry.id ? normalized : item) : undefined,
	)
	client.setQueryData<TimeEntryResponse | null>(timeEntryKeys.active(normalized.user_id), current => {
		if (current === undefined) return current
		if (!entry.end_time) return normalized
		return current?.id === entry.id ? null : current
	})
}

export function removeTimeEntry(client: QueryClient, id: number) {
	client.setQueriesData<TimeEntryResponse[]>({queryKey: timeEntryKeys.lists}, current =>
		current?.some(entry => entry.id === id) ? current.filter(entry => entry.id !== id) : undefined,
	)
	client.setQueriesData<TimeEntryResponse | null>({queryKey: timeEntryKeys.activeTimers}, current =>
		current?.id === id ? null : undefined,
	)
}

function settle(client: QueryClient, taskId?: number) {
	return Promise.all([
		client.invalidateQueries({queryKey: timeEntryKeys.all}),
		invalidateTaskMembership(client, taskId === 0 ? undefined : taskId),
	])
}

export function createTimeEntryMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (body: TimeEntryWritable) => (await timeEntriesCreate({body})).data,
		onSuccess: (entry, _input, client) => {
			patchTimeEntry(client, entry)
			if (entry.task_id) mapTaskEverywhere(client, entry.task_id, task => ({
				...task,
				time_entries_count: task.time_entries_count === undefined ? undefined : task.time_entries_count + 1,
			}))
		},
		onSettled: (input, client) => settle(client, input.task_id),
	})
}

export function updateTimeEntryMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, ...body}: TimeEntryWritable & Required<Pick<TimeEntry, 'id'>>) =>
			(await timeEntriesUpdate({
				path: {id},
				body,
			})).data,
		onSuccess: (entry, _input, client) => patchTimeEntry(client, entry),
		onSettled: (input, client) => settle(client, input.task_id),
	})
}

export function stopTimerMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await timeEntriesTimerStop()).data,
		onSuccess: (entry, _input, client) => patchTimeEntry(client, entry),
		onSettled: (_input, client) => settle(client),
	})
}

// task_id is 0 for entries booked straight onto a project; those have no task count to adjust.
export type DeleteTimeEntryInput = {
	id: number
	taskId: number
}

export function deleteTimeEntryMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id}: DeleteTimeEntryInput) => { await timeEntriesDelete({path: {id}}) },
		onSuccess: (_data, {id, taskId}, client) => {
			removeTimeEntry(client, id)
			if (taskId > 0) mapTaskEverywhere(client, taskId, task => ({
				...task,
				time_entries_count: task.time_entries_count === undefined ? undefined : Math.max(0, task.time_entries_count - 1),
			}))
		},
		onSettled: ({taskId}, client) => settle(client, taskId),
	})
}

export function useCreateTimeEntryMutation() { return useMutation(createTimeEntryMutationOptions()) }
export function useUpdateTimeEntryMutation() { return useMutation(updateTimeEntryMutationOptions()) }
export function useStopTimerMutation() { return useMutation(stopTimerMutationOptions()) }
export function useDeleteTimeEntryMutation() { return useMutation(deleteTimeEntryMutationOptions()) }
