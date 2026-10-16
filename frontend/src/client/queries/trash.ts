import {
	keepPreviousData,
	queryOptions,
	useMutation,
	type QueryClient,
} from '@tanstack/vue-query'
import {
	tasksDeletePermanently,
	tasksRestore,
	tasksTrashEmpty,
	tasksTrashList,
	type User,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {removeFromPages, toPaginated, type Paginated} from './pagination'
import {normalizeTask, type TaskResponse} from './tasks'
import {invalidateTaskMembership} from './taskCache'
import {timeEntryKeys} from './timeEntries'
import {i18n} from '@/i18n'

// Matches the server's hard-coded retention.
export const TRASH_RETENTION_DAYS = 30

export const ERROR_CODE_TASK_IS_IN_TRASH = 4036

export type TrashedTaskPage = Paginated<TaskResponse>

export const trashKeys = {
	all: ['trash'] as const,
	tasks: ['trash', 'tasks'] as const,
	taskList: (page: number) => ['trash', 'tasks', page] as const,
}

export type TaskInTrash = {
	project_id: number
	deleted_at: string
	deleted_by?: User
}

export function taskInTrashFromError(cause: unknown): TaskInTrash | null {
	if (typeof cause !== 'object' || cause === null) return null
	const {code, details} = cause as {code?: number, details?: unknown}
	if (code !== ERROR_CODE_TASK_IS_IN_TRASH || typeof details !== 'object' || details === null) return null
	return details as TaskInTrash
}

export function trashDaysLeft(deletedAt: string | Date, now = new Date()): number {
	const purgeAt = new Date(deletedAt).getTime() + TRASH_RETENTION_DAYS * 24 * 60 * 60 * 1000
	return Math.max(0, Math.ceil((purgeAt - now.getTime()) / (24 * 60 * 60 * 1000)))
}

export function trashedTasksQuery(page: number) {
	return queryOptions({
		queryKey: trashKeys.taskList(page),
		queryFn: async ({signal}): Promise<TrashedTaskPage> => {
			const {data} = await tasksTrashList({query: {page}, signal})
			const paginated = toPaginated(data, page)
			return {...paginated, items: paginated.items.map(normalizeTask)}
		},
		placeholderData: keepPreviousData,
		staleTime: 0,
	})
}

// Deleting or restoring a task stops timers and flips in_trash on its time entries.
export function invalidateAfterTrashChange(client: QueryClient, taskId?: number) {
	return Promise.all([
		client.invalidateQueries({queryKey: trashKeys.all}),
		client.invalidateQueries({queryKey: timeEntryKeys.all}),
		invalidateTaskMembership(client, taskId, 'active'),
	])
}

export function restoreTaskMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => normalizeTask((await tasksRestore({path: {task: id}})).data),
		onSuccess: (_task, id, client) => removeFromPages<TaskResponse>(client, trashKeys.tasks, task => task.id === id),
		onSettled: (id, client) => invalidateAfterTrashChange(client, id),
		successMessage: () => i18n.global.t('trash.restoreSuccess'),
	})
}

export function deleteTaskPermanentlyMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await tasksDeletePermanently({path: {task: id}}) },
		onSuccess: (_data, id, client) => removeFromPages<TaskResponse>(client, trashKeys.tasks, task => task.id === id),
		onSettled: (_id, client) => Promise.all([
			client.invalidateQueries({queryKey: trashKeys.all}),
			client.invalidateQueries({queryKey: timeEntryKeys.all}),
		]),
		successMessage: () => i18n.global.t('trash.deleteForeverSuccess'),
	})
}

export function emptyTaskTrashMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => { await tasksTrashEmpty() },
		onSettled: (_input, client) => Promise.all([
			client.invalidateQueries({queryKey: trashKeys.all}),
			client.invalidateQueries({queryKey: timeEntryKeys.all}),
		]),
		successMessage: () => i18n.global.t('trash.emptySuccess'),
	})
}

export function useRestoreTaskMutation() {
	return useMutation(restoreTaskMutationOptions())
}

export function useDeleteTaskPermanentlyMutation() {
	return useMutation(deleteTaskPermanentlyMutationOptions())
}

export function useEmptyTaskTrashMutation() {
	return useMutation(emptyTaskTrashMutationOptions())
}
