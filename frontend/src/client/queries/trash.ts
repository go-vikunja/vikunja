import {
	keepPreviousData,
	queryOptions,
	useMutation,
	type QueryClient,
} from '@tanstack/vue-query'
import {
	projectsDelete,
	projectsDeletePermanently,
	projectsRestore,
	projectsTrashEmpty,
	projectsTrashList,
	tasksDeletePermanently,
	tasksRestore,
	tasksTrashEmpty,
	tasksTrashList,
	type TrashedSubProject,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {removeFromPages, toPaginated, type Paginated} from './pagination'
import {
	normalizeTask,
	taskKeys,
	type TaskResponse,
} from './tasks'
import {invalidateTaskMembership} from './taskCache'
import {timeEntryKeys} from './timeEntries'
import {
	normalizeProject,
	projectKeys,
	removeProjectSubtree,
	type ProjectResponse,
} from './projects'
import {trashKeys} from './trashShared'
import {i18n} from '@/i18n'
import {removeProjectFromHistory} from '@/modules/projectHistory'

export {
	projectInTrashFromError,
	taskInTrashFromError,
	type ProjectInTrash,
	type TaskInTrash,
	trashKeys,
} from './trashShared'

// Matches the server's hard-coded retention.
export const TRASH_RETENTION_DAYS = 30

export type TrashedTaskPage = Paginated<TaskResponse>
export type TrashedProjectPage = Paginated<ProjectResponse>

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

// The API lists them shallowest first; show each below its parent instead.
function orderTrashedSubProjects(projectId: number, subProjects: TrashedSubProject[]): TrashedSubProject[] {
	const children = new Map<number, TrashedSubProject[]>()
	for (const sub of subProjects) {
		const siblings = children.get(sub.parent_project_id ?? 0) ?? []
		siblings.push(sub)
		children.set(sub.parent_project_id ?? 0, siblings)
	}
	const ordered: TrashedSubProject[] = []
	const visit = (parentId: number) => {
		for (const sub of children.get(parentId) ?? []) {
			ordered.push(sub)
			visit(sub.id ?? 0)
		}
	}
	visit(projectId)
	return ordered
}

export function trashedProjectsQuery(page: number) {
	return queryOptions({
		queryKey: trashKeys.projectList(page),
		queryFn: async ({signal}): Promise<TrashedProjectPage> => {
			const {data} = await projectsTrashList({query: {page}, signal})
			const paginated = toPaginated(data, page)
			return {
				...paginated,
				items: paginated.items.map(raw => {
					const project = normalizeProject(raw)
					return {...project, trashed_sub_projects: orderTrashedSubProjects(project.id, project.trashed_sub_projects ?? [])}
				}),
			}
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

export function deleteProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => {
			await projectsDelete({path: {id}})
		},
		optimistic: {
			queryKeys: () => [projectKeys.list()],
			update: (id, client) => removeProjectSubtree(client, id),
		},
		onSuccess: (_data, id, client, ids) => {
			ids.forEach(projectId => client.removeQueries({queryKey: projectKeys.detail(projectId)}))
			removeProjectFromHistory({id})
		},
		// The view toasts with an undo action
		onSettled: (_input, client) => Promise.all([
			// Already pruned optimistically; a refetch reloads every page.
			client.invalidateQueries({queryKey: projectKeys.list(), refetchType: 'none'}),
			client.invalidateQueries({queryKey: taskKeys.details}),
			invalidateAfterTrashChange(client),
		]),
		// Failures reach the global Vue error handler, which toasts them.
		toastError: () => false,
	})
}

export function restoreProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => (await projectsRestore({path: {id}})).data,
		onSuccess: (_project, id, client) => removeFromPages<ProjectResponse>(client, trashKeys.projects, project => project.id === id),
		onSettled: (_id, client) => invalidateAfterProjectTrashChange(client),
		successMessage: () => i18n.global.t('trash.restoreProjectSuccess'),
	})
}

export function deleteProjectPermanentlyMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => { await projectsDeletePermanently({path: {id}}) },
		onSuccess: (_data, id, client) => removeFromPages<ProjectResponse>(client, trashKeys.projects, project => project.id === id),
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

export function useDeleteProjectMutation() {
	return useMutation(deleteProjectMutationOptions())
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
