import {queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'

import {
	projectsCreate,
	projectsDelete,
	projectsDuplicate,
	projectsList,
	projectsRead,
	projectsUpdate,
	patchProjectsRead,
} from '@/client/generated'
import type {
	Project,
	ProjectView,
	ProjectWritable,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {PERMISSIONS} from '@/constants/permissions'
import {colorFromHex} from '@/helpers/color/colorFromHex'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {i18n} from '@/i18n'

import {contextMutationOptions} from './contextMutation'
import {fetchAllPages} from './fetchAllPages'
import {API_MAX_PER_PAGE} from './pagination'
import {setSubscription} from './subscriptionRequests'
import {taskKeys} from './tasks'

export type ProjectResponse = Omit<Project,
	'id' |
	'title' |
	'description' |
	'hex_color' |
	'identifier' |
	'is_archived' |
	'is_favorite' |
	'parent_project_id' |
	'position' |
	'views'
> & {
	id: number
	title: string
	description: string
	hex_color: string
	identifier: string
	is_archived: boolean
	is_favorite: boolean
	parent_project_id: number
	position: number
	views: ProjectView[]
}

export type ProjectListResult = {
	projects: ProjectResponse[]
	favoriteProject: ProjectResponse | null
	savedFilterProjects: ProjectResponse[]
}

type ProjectDraft = Required<Pick<ProjectWritable,
	'title' |
	'description' |
	'hex_color' |
	'identifier' |
	'is_archived' |
	'is_favorite' |
	'parent_project_id' |
	'position'
>>

type UpdateProjectInput = ProjectDraft & {id: number}

type DuplicateProjectInput = {
	projectId: number
	parentProjectId?: number
	duplicateShares?: boolean
}

export const projectKeys = {
	all: ['projects'] as const,
	list: () => ['projects', 'list'] as const,
	detail: (id: number) => ['projects', 'detail', id] as const,
}

export function createProjectDraft(project: Partial<ProjectWritable> = {}): ProjectDraft {
	return {
		title: '',
		description: '',
		hex_color: '',
		identifier: '',
		is_archived: false,
		is_favorite: false,
		parent_project_id: 0,
		position: 0,
		...project,
	}
}

export function normalizeProject(project: Project): ProjectResponse {
	if (typeof project.id !== 'number') {
		throw new Error('Project response is missing an id')
	}

	return {
		...project,
		id: project.id,
		title: project.title ?? '',
		description: project.description ?? '',
		hex_color: project.hex_color && !project.hex_color.startsWith('#')
			? `#${project.hex_color}`
			: project.hex_color ?? '',
		identifier: project.identifier ?? '',
		is_archived: project.is_archived ?? false,
		is_favorite: project.is_favorite ?? false,
		parent_project_id: project.parent_project_id ?? 0,
		position: project.position ?? 0,
		views: [...(project.views ?? [])].sort((a, b) => (a.position ?? 0) - (b.position ?? 0)),
	}
}

function sortProjects(projects: ProjectResponse[]): ProjectResponse[] {
	return [...projects].sort((a, b) => a.position - b.position)
}

function sortSavedFilters(projects: ProjectResponse[]): ProjectResponse[] {
	return [...projects].sort((a, b) => a.title.localeCompare(b.title))
}

function partitionProjects(projects: Project[]): ProjectListResult {
	const result: ProjectListResult = {
		projects: [],
		favoriteProject: null,
		savedFilterProjects: [],
	}
	const seenProjectIds = new Set<number>()

	for (const rawProject of projects) {
		const project = normalizeProject(rawProject)
		if (seenProjectIds.has(project.id)) {
			continue
		}
		seenProjectIds.add(project.id)
		if (project.id > 0) {
			result.projects.push(project)
		} else if (project.id === -1) {
			result.favoriteProject = project
		} else {
			result.savedFilterProjects.push(project)
		}
	}

	result.projects = sortProjects(result.projects)
	result.savedFilterProjects = sortSavedFilters(result.savedFilterProjects)
	return result
}

async function fetchAllProjects(): Promise<ProjectListResult> {
	const projects = await fetchAllPages(async page => (await projectsList({
		query: {is_archived: true, expand: 'permissions', page, per_page: API_MAX_PER_PAGE},
	})).data)
	return partitionProjects(projects)
}

export function projectsQuery() {
	return queryOptions({
		queryKey: projectKeys.list(),
		queryFn: fetchAllProjects,
		staleTime: 5 * 60 * 1000,
	})
}

export function projectQuery(id: number) {
	return queryOptions({
		queryKey: projectKeys.detail(id),
		queryFn: async () => {
			const {data} = await projectsRead({path: {id}})
			return normalizeProject(data)
		},
		enabled: id !== 0,
	})
}

export function ensureProjects(): Promise<ProjectListResult> {
	return queryClient.ensureQueryData(projectsQuery())
}

export function refreshProjects(): Promise<ProjectListResult> {
	return queryClient.fetchQuery({...projectsQuery(), staleTime: 0})
}

export function ensureProject(id: number): Promise<ProjectResponse> {
	return queryClient.ensureQueryData(projectQuery(id))
}

export function refreshProject(id: number): Promise<ProjectResponse> {
	return queryClient.fetchQuery({...projectQuery(id), staleTime: 0})
}

export function getCachedProject(id: number, client: QueryClient = queryClient): ProjectResponse | undefined {
	const list = client.getQueryData<ProjectListResult>(projectKeys.list())

	// Pseudo projects have no detail endpoint, so the navigation list is their only source.
	if (id === -1) {
		return list?.favoriteProject ?? undefined
	}
	if (id < 0) {
		return list?.savedFilterProjects.find(project => project.id === id)
	}

	return client.getQueryData<ProjectResponse>(projectKeys.detail(id))
		?? list?.projects.find(project => project.id === id)
}

export function findProjectByExactTitle(
	projects: readonly ProjectResponse[],
	title: string,
): ProjectResponse | null {
	return projects.find(project => project.title.toLowerCase() === title.toLowerCase()) ?? null
}

export function getSavedFilterIdFromProjectId(projectId: number): number {
	return Math.max(0, projectId * -1 - 1)
}

export function getProjectIdFromSavedFilterId(savedFilterId: number): number {
	return savedFilterId > 0 ? savedFilterId * -1 - 1 : 0
}

export function isSavedFilterProject(project: Pick<Project, 'id'> | null | undefined): boolean {
	return getSavedFilterIdFromProjectId(project?.id ?? 0) > 0
}

function projectBody(project: ProjectWritable): ProjectWritable {
	return {
		title: project.title,
		description: project.description,
		hex_color: typeof project.hex_color === 'undefined'
			? undefined
			: colorFromHex(project.hex_color),
		identifier: project.identifier,
		is_archived: project.is_archived,
		is_favorite: project.is_favorite,
		parent_project_id: project.parent_project_id,
		position: project.position,
	}
}

function replaceProjectInList(
	current: ProjectListResult,
	project: ProjectResponse,
): ProjectListResult {
	return {...current, projects: sortProjects([
		...current.projects.filter(existing => existing.id !== project.id),
		project,
	])}
}

export function mapProjectNavigationItem(
	current: ProjectListResult,
	id: number,
	updater: (project: ProjectResponse) => ProjectResponse,
): ProjectListResult {
	const update = (project: ProjectResponse) => project.id === id ? updater(project) : project
	return {
		projects: sortProjects(current.projects.map(update)),
		favoriteProject: current.favoriteProject ? update(current.favoriteProject) : null,
		savedFilterProjects: sortSavedFilters(current.savedFilterProjects.map(update)),
	}
}

function mergeProjectMetadata(previous: ProjectResponse, updated: ProjectResponse): ProjectResponse {
	return {
		...previous,
		...updated,
		max_permission: updated.max_permission ?? previous.max_permission,
		views: updated.views.length ? updated.views : previous.views,
	}
}

// Project callers let failures reach the global Vue error handler, which toasts them.
const toastError = () => false

export function createProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (project: ProjectWritable) => {
			const {data} = await projectsCreate({body: projectBody(project)})
			return normalizeProject(data)
		},
		onSuccess: (created, _input, client) => {
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? replaceProjectInList(current, created) : current,
			)
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: projectKeys.list()}),
		successMessage: () => i18n.global.t('project.create.createdSuccess'),
		toastError,
	})
}

export function updateProjectMutationOptions(successMessage?: string) {
	return contextMutationOptions({
		mutationFn: async ({id, ...project}: UpdateProjectInput) => {
			const {data} = await projectsUpdate({
				path: {id},
				body: projectBody(project),
			})
			return normalizeProject({...project, ...data})
		},
		optimistic: {
			queryKeys: ({id}) => [projectKeys.list(), projectKeys.detail(id)],
			update: ({id, ...project}, client) => {
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? mapProjectNavigationItem(current, id, existing => normalizeProject({...existing, ...project})) : current,
				)
				client.setQueryData<ProjectResponse>(projectKeys.detail(id), current =>
					current ? normalizeProject({...current, ...project}) : current,
				)
			},
		},
		onSuccess: (updated, _input, client) => {
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? mapProjectNavigationItem(current, updated.id, previous =>
					mergeProjectMetadata(previous, updated),
				) : current,
			)
			client.setQueryData<ProjectResponse>(projectKeys.detail(updated.id), current =>
				current ? mergeProjectMetadata(current, updated) : current,
			)
		},
		onSettled: ({id}, client) => Promise.all([
			client.invalidateQueries({queryKey: projectKeys.list()}),
			client.invalidateQueries({queryKey: projectKeys.detail(id)}),
		]),
		successMessage: () => successMessage,
		toastError,
	})
}

export function patchProjectFavoriteMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({id, isFavorite}: {id: number; isFavorite: boolean}) => {
			const {data} = await patchProjectsRead({
				path: {id},
				body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
			})
			return normalizeProject(data)
		},
		optimistic: {
			queryKeys: ({id}) => [projectKeys.list(), projectKeys.detail(id)],
			update: ({id, isFavorite}, client) => {
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? mapProjectNavigationItem(current, id, project => ({...project, is_favorite: isFavorite})) : current,
				)
				client.setQueryData<ProjectResponse>(projectKeys.detail(id), current =>
					current ? {...current, is_favorite: isFavorite} : current,
				)
			},
		},
		onSuccess: (updated, {id}, client) => {
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? mapProjectNavigationItem(current, id, project => ({...project, is_favorite: updated.is_favorite})) : current,
			)
			client.setQueryData<ProjectResponse>(projectKeys.detail(id), current =>
				current ? {...current, is_favorite: updated.is_favorite} : current,
			)
		},
		onSettled: ({id}, client) => Promise.all([
			client.invalidateQueries({queryKey: projectKeys.list()}),
			client.invalidateQueries({queryKey: projectKeys.detail(id)}),
		]),
		toastError,
	})
}

export function setProjectSubscriptionMutationOptions() {
	return contextMutationOptions({
		mutationFn: ({projectId, subscribed}: {projectId: number; subscribed: boolean}) =>
			setSubscription('project', projectId, subscribed),
		onSuccess: (subscription, {projectId}, client) => {
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? mapProjectNavigationItem(current, projectId, project => ({...project, subscription})) : current,
			)
			client.setQueryData<ProjectResponse>(projectKeys.detail(projectId), current =>
				current ? {...current, subscription} : current,
			)
		},
		// Refetching the list reloads every page, so it is only marked stale; sub projects
		// inherit the subscription and pick it up on the next list load.
		onSettled: ({projectId}, client) => {
			const list = client.getQueryData<ProjectListResult>(projectKeys.list())
			const ids = list ? descendantIds(list.projects, projectId) : [projectId]
			return Promise.all([
				client.invalidateQueries({queryKey: projectKeys.list(), refetchType: 'none'}),
				...ids.map(id => client.invalidateQueries({queryKey: projectKeys.detail(id)})),
				// Tasks inherit the project subscription.
				client.invalidateQueries({queryKey: taskKeys.details}),
			])
		},
		successMessage: (_subscription, {subscribed}) => i18n.global.t(subscribed
			? 'task.subscription.subscribeSuccessProject'
			: 'task.subscription.unsubscribeSuccessProject'),
		toastError,
	})
}

function descendantIds(projects: readonly ProjectResponse[], projectId: number): number[] {
	const children = projects.filter(project => project.parent_project_id === projectId)
	return [projectId, ...children.flatMap(project => descendantIds(projects, project.id))]
}

export function deleteProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async (id: number) => {
			await projectsDelete({path: {id}})
		},
		optimistic: {
			queryKeys: () => [projectKeys.list()],
			update: (id, client) => {
				const list = client.getQueryData<ProjectListResult>(projectKeys.list())
				const ids = new Set(list ? descendantIds(list.projects, id) : [id])
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? {...current, projects: current.projects.filter(project => !ids.has(project.id))} : current,
				)
				return ids
			},
		},
		onSuccess: (_data, id, client, ids) => {
			ids.forEach(projectId => client.removeQueries({queryKey: projectKeys.detail(projectId)}))
			removeProjectFromHistory({id})
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: projectKeys.list()}),
		successMessage: () => i18n.global.t('project.delete.success'),
		toastError,
	})
}

export function duplicateProjectMutationOptions() {
	return contextMutationOptions({
		mutationFn: async ({
			projectId,
			parentProjectId = 0,
			duplicateShares = false,
		}: DuplicateProjectInput) => {
			const {data} = await projectsDuplicate({
				path: {projectid: projectId},
				body: {parent_project_id: parentProjectId, duplicate_shares: duplicateShares},
			})
			if (!data.duplicated_project) {
				throw new Error('Project duplicate response is missing the duplicated project')
			}
			return normalizeProject({...data.duplicated_project, max_permission: PERMISSIONS.ADMIN})
		},
		onSuccess: (duplicate, _input, client) => {
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? replaceProjectInList(current, duplicate) : current,
			)
		},
		onSettled: (_input, client) => client.invalidateQueries({queryKey: projectKeys.list()}),
		successMessage: () => i18n.global.t('project.duplicate.success'),
		toastError,
	})
}

export function useCreateProjectMutation() {
	return useMutation(createProjectMutationOptions())
}

export function useUpdateProjectMutation(successMessage?: string) {
	return useMutation(updateProjectMutationOptions(successMessage))
}

export function usePatchProjectFavoriteMutation() {
	return useMutation(patchProjectFavoriteMutationOptions())
}

export function useSetProjectSubscriptionMutation() {
	return useMutation(setProjectSubscriptionMutationOptions())
}

export function useDeleteProjectMutation() {
	return useMutation(deleteProjectMutationOptions())
}

export function useDuplicateProjectMutation() {
	return useMutation(duplicateProjectMutationOptions())
}
