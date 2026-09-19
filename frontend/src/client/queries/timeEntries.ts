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
import {
	pageSizeFor,
	totalPagesFor,
	type Paginated,
} from './pagination'
import {
	invalidateTaskMembership,
	mapTaskEverywhere,
} from './taskCache'

export type TimeEntryResponse = Omit<TimeEntry, 'id' | 'user_id' | 'task_id' | 'project_id' | 'comment'> &
	Required<Pick<TimeEntry, 'id' | 'user_id' | 'task_id' | 'project_id' | 'comment'>>

export type TimeEntryPage = Paginated<TimeEntryResponse>

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
	list: (filter: string, timezone: string, page: number, perPage: number) =>
		['time-entries', 'list', filter, timezone, page, perPage] as const,
	activeTimers: ['time-entries', 'active'] as const,
	active: (userId: number) => ['time-entries', 'active', userId] as const,
}

export function timeEntriesQuery(filter: string, timezone: string, page: number, perPage: number) {
	const clampedPerPage = pageSizeFor(perPage)
	return queryOptions({
		queryKey: timeEntryKeys.list(filter, timezone, page, clampedPerPage),
		queryFn: async ({signal}): Promise<TimeEntryPage> => {
			const {data} = await timeEntriesList({
				query: {
					filter,
					filter_timezone: timezone,
					page,
					per_page: clampedPerPage,
				},
				signal,
			})
			return {
				...data,
				items: (data.items ?? []).map(normalizeTimeEntry),
				page: data.page ?? page,
				per_page: data.per_page ?? clampedPerPage,
				total: data.total ?? 0,
				total_pages: data.total_pages ?? 0,
			}
		},
	})
}

export function activeTimerQuery(userId: number) {
	return queryOptions({
		queryKey: timeEntryKeys.active(userId),
		queryFn: async ({signal}) => {
			if (userId <= 0) return null
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
	client.setQueriesData<TimeEntryPage>({queryKey: timeEntryKeys.lists}, current =>
		current?.items.some(item => item.id === entry.id)
			? {
				...current,
				items: current.items.map(item => item.id === entry.id ? normalized : item),
			}
			: undefined,
	)
	client.setQueryData<TimeEntryResponse | null>(timeEntryKeys.active(normalized.user_id), current => {
		if (current === undefined) return current
		if (!entry.end_time) return normalized
		return current?.id === entry.id ? null : current
	})
}

export function removeTimeEntry(client: QueryClient, id: number) {
	client.setQueriesData<TimeEntryPage>({queryKey: timeEntryKeys.lists}, current => {
		if (!current?.items.some(entry => entry.id === id)) return undefined
		const total = Math.max(0, current.total - 1)
		return {
			...current,
			items: current.items.filter(entry => entry.id !== id),
			total,
			total_pages: totalPagesFor(current, total),
		}
	})
	client.setQueriesData<TimeEntryResponse | null>({queryKey: timeEntryKeys.activeTimers}, current =>
		current?.id === id ? null : undefined,
	)
}

function bumpTaskEntryCount(client: QueryClient, taskId: number, delta: number) {
	if (taskId <= 0) return
	mapTaskEverywhere(client, taskId, task => ({
		...task,
		time_entries_count: task.time_entries_count === undefined
			? undefined
			: Math.max(0, task.time_entries_count + delta),
	}))
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
			bumpTaskEntryCount(client, entry.task_id ?? 0, 1)
		},
		onSettled: (input, client) => settle(client, input.task_id),
	})
}

export type UpdateTimeEntryInput = TimeEntryWritable & Required<Pick<TimeEntry, 'id'>>

export function cachedTaskIdOf(client: QueryClient, id: number): number | undefined {
	for (const [, page] of client.getQueriesData<TimeEntryPage>({queryKey: timeEntryKeys.lists})) {
		const cached = page?.items.find(entry => entry.id === id)
		if (cached) return cached.task_id
	}
	for (const [, timer] of client.getQueriesData<TimeEntryResponse | null>({queryKey: timeEntryKeys.activeTimers})) {
		if (timer?.id === id) return timer.task_id
	}
	return undefined
}

export function updateTimeEntryMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({
			id,
			...body
		}: UpdateTimeEntryInput) =>
			(await timeEntriesUpdate({
				path: {id},
				body,
			})).data,
		onSuccess: (entry, {id}, client) => {
			const previousTaskId = cachedTaskIdOf(client, id)
			patchTimeEntry(client, entry)
			const taskId = entry.task_id ?? 0
			if (previousTaskId === undefined || previousTaskId === taskId) return
			bumpTaskEntryCount(client, previousTaskId, -1)
			bumpTaskEntryCount(client, taskId, 1)
		},
		onSettled: (input, client) => settle(client, input.task_id),
	})
}

export function stopTimerMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => (await timeEntriesTimerStop()).data,
		onSuccess: (entry, _input, client) => patchTimeEntry(client, entry),
		// Stopping changes no task-derived data (the entry was already counted at create); settle() would also mark every task list/board stale.
		onSettled: (_input, client) => client.invalidateQueries({queryKey: timeEntryKeys.all}),
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
			bumpTaskEntryCount(client, taskId, -1)
		},
		onSettled: ({taskId}, client) => settle(client, taskId),
	})
}

export function useCreateTimeEntryMutation() {
	return useMutation(createTimeEntryMutationOptions())
}

export function useUpdateTimeEntryMutation() {
	return useMutation(updateTimeEntryMutationOptions())
}

export function useStopTimerMutation() {
	return useMutation(stopTimerMutationOptions())
}

export function useDeleteTimeEntryMutation() {
	return useMutation(deleteTimeEntryMutationOptions())
}
