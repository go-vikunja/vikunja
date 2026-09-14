import {mutationOptions, queryOptions, useMutation} from '@tanstack/vue-query'
import type {QueryClient} from '@tanstack/vue-query'

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
	ProjectsListData,
	ProjectView,
	ProjectWritable,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {PERMISSIONS} from '@/constants/permissions'
import {colorFromHex} from '@/helpers/color/colorFromHex'
import {removeProjectFromHistory} from '@/modules/projectHistory'
import {i18n} from '@/i18n'
import {success} from '@/message'

export type ProjectListArgs = Omit<NonNullable<ProjectsListData['query']>, 'page' | 'per_page'>

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

export type ProjectDraft = Required<Pick<ProjectWritable,
	'title' |
	'description' |
	'hex_color' |
	'identifier' |
	'is_archived' |
	'is_favorite' |
	'parent_project_id' |
	'position'
>>

export type UpdateProjectInput = ProjectDraft & {id: number}

export type DuplicateProjectInput = {
	projectId: number
	parentProjectId?: number
	duplicateShares?: boolean
}

export const defaultProjectListArgs = {
	is_archived: true,
	expand: 'permissions',
} as const satisfies ProjectListArgs

export const projectKeys = {
	all: ['projects'] as const,
	lists: () => ['projects', 'list'] as const,
	list: (args: ProjectListArgs = defaultProjectListArgs) => ['projects', 'list', args] as const,
	details: () => ['projects', 'detail'] as const,
	detailRoot: (id: number) => ['projects', 'detail', id] as const,
	detail: (id: number, format: 'html' | 'markdown' = 'html') => [
		'projects',
		'detail',
		id,
		format,
	] as const,
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

async function fetchProjects(args: ProjectListArgs): Promise<ProjectListResult> {
	const projects: Project[] = []
	let page = 1

	while (true) {
		const {data} = await projectsList({
			query: {...args, page, per_page: 1000},
		})
		projects.push(...(data.items ?? []))
		if (page >= (data.total_pages ?? 1)) {
			break
		}
		page++
	}

	return partitionProjects(projects)
}

export function projectsQuery(args: ProjectListArgs = defaultProjectListArgs) {
	return queryOptions({
		queryKey: projectKeys.list(args),
		queryFn: () => fetchProjects(args),
		staleTime: 5 * 60 * 1000,
	})
}

export function projectQuery(id: number, format: 'html' | 'markdown' = 'html') {
	return queryOptions({
		queryKey: projectKeys.detail(id, format),
		queryFn: async () => {
			const {data} = await projectsRead({path: {id}, query: {format}})
			return normalizeProject(data)
		},
		enabled: id !== 0,
	})
}

export function ensureProjects(args: ProjectListArgs = defaultProjectListArgs): Promise<ProjectListResult> {
	return queryClient.ensureQueryData(projectsQuery(args))
}

export function refreshProjects(args: ProjectListArgs = defaultProjectListArgs): Promise<ProjectListResult> {
	return queryClient.fetchQuery({...projectsQuery(args), staleTime: 0})
}

export function ensureProject(id: number, format: 'html' | 'markdown' = 'html'): Promise<ProjectResponse> {
	return queryClient.ensureQueryData(projectQuery(id, format))
}

export function refreshProject(id: number, format: 'html' | 'markdown' = 'html'): Promise<ProjectResponse> {
	return queryClient.fetchQuery({...projectQuery(id, format), staleTime: 0})
}

export function getProjectById(projects: readonly ProjectResponse[], id: number): ProjectResponse | undefined {
	return projects.find(project => project.id === id)
}

export function isOrphanedProject(projects: readonly ProjectResponse[], project: ProjectResponse): boolean {
	return project.parent_project_id !== 0 && !getProjectById(projects, project.parent_project_id)
}

export function getRootProjects(projects: readonly ProjectResponse[]): ProjectResponse[] {
	return sortProjects(projects.filter(project =>
		!project.is_archived && (project.parent_project_id === 0 || isOrphanedProject(projects, project)),
	))
}

export function getChildProjects(projects: readonly ProjectResponse[], id: number): ProjectResponse[] {
	return sortProjects(projects.filter(project => project.parent_project_id === id))
}

export function getProjectAncestors(
	projects: readonly ProjectResponse[],
	project: ProjectResponse | undefined,
): ProjectResponse[] {
	if (!project) {
		return []
	}

	if (project.parent_project_id === 0) {
		return [project]
	}

	const parent = getProjectById(projects, project.parent_project_id)
	return [...getProjectAncestors(projects, parent), project]
}

export function getEffectiveParentProjectId(
	projects: readonly ProjectResponse[],
	project: ProjectResponse,
	parentProjectIdFromDom: number,
): number {
	if (parentProjectIdFromDom === 0 && isOrphanedProject(projects, project)) {
		return project.parent_project_id
	}
	return parentProjectIdFromDom
}

export function findProjectByExactTitle(
	projects: readonly ProjectResponse[],
	title: string,
): ProjectResponse | null {
	return projects.find(project => project.title.toLowerCase() === title.toLowerCase()) ?? null
}

export function findProjectByIdentifier(
	projects: readonly ProjectResponse[],
	identifier: string,
): ProjectResponse | null {
	return projects.find(project => project.identifier.toLowerCase() === identifier.toLowerCase()) ?? null
}

export function searchProjects(
	projects: readonly ProjectResponse[],
	query: string,
	includeArchived = false,
): ProjectResponse[] {
	if (query === '') {
		return []
	}

	const normalizedQuery = query.toLowerCase()
	return projects.filter(project =>
		project.is_archived === includeArchived &&
		(project.title.toLowerCase().includes(normalizedQuery) ||
			project.description.toLowerCase().includes(normalizedQuery)),
	)
}

export function getFavoriteNavigationItems(result: ProjectListResult): ProjectResponse[] {
	return [
		...(result.favoriteProject ? [result.favoriteProject] : []),
		...result.savedFilterProjects.filter(project => !project.is_archived && project.is_favorite),
		...sortProjects(result.projects.filter(project => !project.is_archived && project.is_favorite)),
	]
}

export function getSavedFilterIdFromProjectId(projectId: number): number {
	return Math.max(0, projectId * -1 - 1)
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

export function replaceProjectInList(
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

async function snapshotProjects(client: QueryClient, id?: number) {
	const request = captureClientRequestContext()
	await Promise.all([
		client.cancelQueries({queryKey: projectKeys.lists()}),
		...(id === undefined ? [] : [client.cancelQueries({queryKey: projectKeys.detailRoot(id)})]),
	])
	assertClientRequestContext(request)
	return {
		request,
		lists: client.getQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}),
		details: id === undefined ? [] : client.getQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(id)}),
	}
}

type ProjectSnapshot = Awaited<ReturnType<typeof snapshotProjects>>

function restoreProjects(client: QueryClient, snapshot: ProjectSnapshot | undefined) {
	if (!snapshot || !isClientRequestContextCurrent(snapshot.request)) {
		return
	}
	for (const [key, previous] of [...snapshot.lists, ...snapshot.details]) {
		if (previous) {
			client.setQueryData(key, previous)
		}
	}
}

export function createProjectMutationOptions(format: 'html' | 'markdown' = 'html') {
	return mutationOptions({
		onMutate: () => ({request: captureClientRequestContext()}),
		mutationFn: async (project: ProjectWritable) => {
			const request = captureClientRequestContext()
			const {data} = await projectsCreate({body: projectBody(project), query: {format}})
			assertClientRequestContext(request)
			return normalizeProject(data)
		},
		onSuccess: (created, _input, context, {client}) => {
			assertClientRequestContext(context.request)
			if (format === 'html') {
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? replaceProjectInList(current, created) : current,
				)
			}
			success({message: i18n.global.t('project.create.createdSuccess')})
		},
		onSettled: async (_data, _error, _input, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await client.invalidateQueries({queryKey: projectKeys.lists()})
				assertClientRequestContext(context.request)
			}
		},
	})
}

export function updateProjectMutationOptions(
	format: 'html' | 'markdown' = 'html',
	successMessage?: string,
) {
	return mutationOptions({
		mutationFn: async ({id, ...project}: UpdateProjectInput) => {
			const request = captureClientRequestContext()
			const {data} = await projectsUpdate({
				path: {id},
				body: projectBody(project),
				query: {format},
			})
			assertClientRequestContext(request)
			return normalizeProject({...project, ...data})
		},
		onMutate: async ({id, ...project}, {client}) => {
			const snapshot = await snapshotProjects(client, id)
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? mapProjectNavigationItem(current, id, existing => {
					const {description: _description, ...fields} = project
					return normalizeProject({...existing, ...fields})
				}) : current,
			)
			client.setQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(id)}, current => {
				if (!current) {
					return current
				}
				const {description: _description, ...fields} = project
				return normalizeProject({...current, ...fields})
			})
			client.setQueryData<ProjectResponse>(projectKeys.detail(id, format), current =>
				current ? normalizeProject({...current, ...project}) : current,
			)
			return snapshot
		},
		onError: (_error, _input, context, {client}) => restoreProjects(client, context),
		onSuccess: (updated, _input, context, {client}) => {
			assertClientRequestContext(context.request)
			if (format === 'html') {
				client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
					current ? mapProjectNavigationItem(current, updated.id, previous =>
						mergeProjectMetadata(previous, updated),
					) : current,
				)
			}
			client.setQueryData<ProjectResponse>(projectKeys.detail(updated.id, format), current =>
				current ? mergeProjectMetadata(current, updated) : current,
			)
			if (successMessage) {
				success({message: successMessage})
			}
		},
		onSettled: async (_data, _error, {id}, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await Promise.all([
					client.invalidateQueries({queryKey: projectKeys.lists()}),
					client.invalidateQueries({queryKey: projectKeys.detailRoot(id)}),
				])
				assertClientRequestContext(context.request)
			}
		},
	})
}

export function patchProjectFavoriteMutationOptions() {
	return mutationOptions({
		mutationFn: async ({id, isFavorite}: {id: number; isFavorite: boolean}) => {
			const request = captureClientRequestContext()
			const {data} = await patchProjectsRead({
				path: {id},
				body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
			})
			assertClientRequestContext(request)
			return normalizeProject(data)
		},
		onMutate: async ({id, isFavorite}, {client}) => {
			const snapshot = await snapshotProjects(client, id)
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? mapProjectNavigationItem(current, id, project => ({...project, is_favorite: isFavorite})) : current,
			)
			client.setQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(id)}, current =>
				current ? {...current, is_favorite: isFavorite} : current,
			)
			return snapshot
		},
		onError: (_error, _input, context, {client}) => restoreProjects(client, context),
		onSuccess: (updated, {id}, context, {client}) => {
			assertClientRequestContext(context.request)
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? mapProjectNavigationItem(current, id, project => ({...project, is_favorite: updated.is_favorite})) : current,
			)
			client.setQueriesData<ProjectResponse>({queryKey: projectKeys.detailRoot(id)}, current =>
				current ? {...current, is_favorite: updated.is_favorite} : current,
			)
		},
		onSettled: async (_data, _error, {id}, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await Promise.all([
					client.invalidateQueries({queryKey: projectKeys.lists()}),
					client.invalidateQueries({queryKey: projectKeys.detailRoot(id)}),
				])
				assertClientRequestContext(context.request)
			}
		},
	})
}

function descendantIds(projects: readonly ProjectResponse[], projectId: number): number[] {
	const children = projects.filter(project => project.parent_project_id === projectId)
	return [projectId, ...children.flatMap(project => descendantIds(projects, project.id))]
}

export function deleteProjectMutationOptions() {
	return mutationOptions({
		mutationFn: async (id: number) => {
			const request = captureClientRequestContext()
			await projectsDelete({path: {id}})
			assertClientRequestContext(request)
		},
		onMutate: async (id, {client}) => {
			const snapshot = await snapshotProjects(client)
			const ids = new Set([id])
			snapshot.lists.forEach(([, result]) => {
				if (result) {
					descendantIds(result.projects, id).forEach(child => ids.add(child))
				}
			})
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? {...current, projects: current.projects.filter(project => !ids.has(project.id))} : current,
			)
			return {...snapshot, ids}
		},
		onError: (_error, _input, context, {client}) => restoreProjects(client, context),
		onSuccess: (_data, id, context, {client}) => {
			assertClientRequestContext(context.request)
			context.ids.forEach(projectId => client.removeQueries({queryKey: projectKeys.detailRoot(projectId)}))
			removeProjectFromHistory({id})
			success({message: i18n.global.t('project.delete.success')})
		},
		onSettled: async (_data, _error, _input, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await client.invalidateQueries({queryKey: projectKeys.lists()})
				assertClientRequestContext(context.request)
			}
		},
	})
}

export function duplicateProjectMutationOptions() {
	return mutationOptions({
		onMutate: () => ({request: captureClientRequestContext()}),
		mutationFn: async ({
			projectId,
			parentProjectId = 0,
			duplicateShares = false,
		}: DuplicateProjectInput) => {
			const request = captureClientRequestContext()
			const {data} = await projectsDuplicate({
				path: {projectid: projectId},
				body: {parent_project_id: parentProjectId, duplicate_shares: duplicateShares},
			})
			assertClientRequestContext(request)
			if (!data.duplicated_project) {
				throw new Error('Project duplicate response is missing the duplicated project')
			}
			return normalizeProject({...data.duplicated_project, max_permission: PERMISSIONS.ADMIN})
		},
		onSuccess: (duplicate, _input, context, {client}) => {
			assertClientRequestContext(context.request)
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? replaceProjectInList(current, duplicate) : current,
			)
			success({message: i18n.global.t('project.duplicate.success')})
		},
		onSettled: async (_data, _error, _input, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await client.invalidateQueries({queryKey: projectKeys.lists()})
				assertClientRequestContext(context.request)
			}
		},
	})
}

export function legacySavedFilterFavoriteMutationOptions() {
	return mutationOptions({
		mutationFn: async ({id, isFavorite}: {id: number; isFavorite: boolean}) => {
			const request = captureClientRequestContext()
			const [{default: SavedFilterService}, {default: SavedFilterModel}] = await Promise.all([
				import('@/services/savedFilter'),
				import('@/models/savedFilter'),
			])
			assertClientRequestContext(request)
			const service = new SavedFilterService()
			const filter = await service.get(new SavedFilterModel({id}))
			assertClientRequestContext(request)
			filter.isFavorite = isFavorite
			const updated = await service.update(filter)
			assertClientRequestContext(request)
			return updated
		},
		onMutate: async ({id, isFavorite}, {client}) => {
			const snapshot = await snapshotProjects(client)
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? mapProjectNavigationItem(current, -id - 1, project => ({...project, is_favorite: isFavorite})) : current,
			)
			return snapshot
		},
		onError: (_error, _input, context, {client}) => restoreProjects(client, context),
		onSuccess: (updated, {id}, context, {client}) => {
			assertClientRequestContext(context.request)
			client.setQueriesData<ProjectListResult>({queryKey: projectKeys.lists()}, current =>
				current ? mapProjectNavigationItem(current, -id - 1, project => ({...project, is_favorite: updated.isFavorite})) : current,
			)
		},
		onSettled: async (_data, _error, _input, context, {client}) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				await client.invalidateQueries({queryKey: projectKeys.lists()})
				assertClientRequestContext(context.request)
			}
		},
	})
}

export function useCreateProjectMutation(format: 'html' | 'markdown' = 'html') {
	return useMutation(createProjectMutationOptions(format))
}

export function useUpdateProjectMutation(format: 'html' | 'markdown' = 'html', successMessage?: string) {
	return useMutation(updateProjectMutationOptions(format, successMessage))
}

export function usePatchProjectFavoriteMutation() {
	return useMutation(patchProjectFavoriteMutationOptions())
}

export function useDeleteProjectMutation() {
	return useMutation(deleteProjectMutationOptions())
}

export function useDuplicateProjectMutation() {
	return useMutation(duplicateProjectMutationOptions())
}

export function useLegacySavedFilterFavoriteMutation() {
	return useMutation(legacySavedFilterFavoriteMutationOptions())
}
