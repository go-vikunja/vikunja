import {
	keepPreviousData,
	queryOptions,
	useMutation,
	type QueryClient,
} from '@tanstack/vue-query'
import {
	projectsDeletePermanently,
	projectsRestore,
	projectsTrashEmpty,
	projectsTrashList,
	tasksDeletePermanently,
	tasksRestore,
	tasksTrashEmpty,
	tasksTrashList,
	type Project,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {removeFromPages, toPaginated, type Paginated} from './pagination'
import {normalizeTask, type TaskResponse} from './tasks'
import {invalidateTaskMembership} from './taskCache'
import {timeEntryKeys} from './timeEntries'
import {projectKeys} from './projects'
import {trashKeys} from './trashShared'
import {i18n} from '@/i18n'

export {
	ERROR_CODE_PROJECT_IS_IN_TRASH,
	ERROR_CODE_TASK_IS_IN_TRASH,
	isInTrashError,
	projectInTrashFromError,
	taskInTrashFromError,
	type ProjectInTrash,
	type TaskInTrash,
	trashKeys,
} from './trashShared'

// Matches the server's hard-coded retention.
export const TRASH_RETENTION_DAYS = 30

export type TrashedTaskPage = Paginated<TaskResponse>
export type TrashedProjectPage = Paginated<Project>


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

export function trashedProjectsQuery(page: number) {
	return queryOptions({
		queryKey: trashKeys.projectList(page),
		queryFn: async ({signal}): Promise<TrashedProjectPage> => {
			const {data} = await projectsTrashList({query: {page}, signal})
			return toPaginated(data, page)
		},
		placeholderData: keepPreviousData,
		staleTime: 0,
	})
}

// A project carries its tasks and their time entries with it.
function invalidateAfterProjectTrashChange(client: QueryClient) {
	return Promise.all([
		client.invalidateQueries({queryKey: projectKeys.all}),
		invalidateAfterTrashChange(client),
	])
}

export function restoreProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await projectsRestore({path: {id}})).data,
		onSuccess: (_project, id, client) => removeFromPages<Project>(client, trashKeys.projects, project => project.id === id),
		onSettled: (_id, client) => invalidateAfterProjectTrashChange(client),
		successMessage: () => i18n.global.t('trash.restoreProjectSuccess'),
	})
}

export function deleteProjectPermanentlyMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await projectsDeletePermanently({path: {id}}) },
		onSuccess: (_data, id, client) => removeFromPages<Project>(client, trashKeys.projects, project => project.id === id),
		onSettled: (_id, client) => Promise.all([
			client.invalidateQueries({queryKey: trashKeys.all}),
			client.invalidateQueries({queryKey: timeEntryKeys.all}),
		]),
		successMessage: () => i18n.global.t('trash.deleteProjectForeverSuccess'),
	})
}

export function emptyProjectTrashMutationOptions() {
	return contextMutationOptions({
		mutationFn: async () => { await projectsTrashEmpty() },
		onSettled: (_input, client) => Promise.all([
			client.invalidateQueries({queryKey: trashKeys.all}),
			client.invalidateQueries({queryKey: timeEntryKeys.all}),
		]),
		successMessage: () => i18n.global.t('trash.emptySuccess'),
	})
}

export function useRestoreProjectMutation() {
	return useMutation(restoreProjectMutationOptions())
}

export function useDeleteProjectPermanentlyMutation() {
	return useMutation(deleteProjectPermanentlyMutationOptions())
}

export function useEmptyProjectTrashMutation() {
	return useMutation(emptyProjectTrashMutationOptions())
}
