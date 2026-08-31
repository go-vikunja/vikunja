import {queryOptions, useMutation} from '@tanstack/vue-query'

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
import {PERMISSIONS} from '@/constants/permissions'
import {getToken, getTokenIdentity} from '@/helpers/auth'
import {colorFromHex} from '@/helpers/color/colorFromHex'
import {getApiV2BaseUrl} from '@/helpers/fetcher'
import {removeProjectFromHistory} from '@/modules/projectHistory'

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

export type ProjectNavigationItem = ProjectResponse & {
	kind: 'favorites' | 'saved-filter'
	savedFilterId?: number
}

export type ProjectListResult = {
	projects: ProjectResponse[]
	favoriteProject: ProjectNavigationItem | null
	savedFilterProjects: ProjectNavigationItem[]
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

type ProjectMutationContext = {
	identity: ReturnType<typeof getTokenIdentity>
	apiV2BaseUrl: string
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

function getProjectMutationContext(): ProjectMutationContext {
	return {
		identity: getTokenIdentity(getToken()),
		apiV2BaseUrl: getApiV2BaseUrl(),
	}
}

function assertProjectMutationContext(context: ProjectMutationContext): void {
	const identity = getTokenIdentity(getToken())
	const identityMatches = context.identity === null
		? identity === null
		: identity?.id === context.identity.id && identity.type === context.identity.type
	if (!identityMatches || context.apiV2BaseUrl !== getApiV2BaseUrl()) {
		throw new DOMException('Project mutation context changed', 'AbortError')
	}
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

function toNavigationItem(project: ProjectResponse): ProjectNavigationItem {
	if (project.id === -1) {
		return {...project, kind: 'favorites'}
	}

	return {
		...project,
		kind: 'saved-filter',
		savedFilterId: getSavedFilterIdFromProjectId(project.id),
	}
}

function sortProjects(projects: ProjectResponse[]): ProjectResponse[] {
	return [...projects].sort((a, b) => a.position - b.position)
}

function sortSavedFilters(projects: ProjectNavigationItem[]): ProjectNavigationItem[] {
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
			result.favoriteProject = toNavigationItem(project)
		} else {
			result.savedFilterProjects.push(toNavigationItem(project))
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

function setProjectInDefaultList(project: ProjectResponse) {
	queryClient.setQueryData<ProjectListResult>(
		projectKeys.list(),
		current => current
			? {...current, projects: sortProjects([
				...current.projects.filter(existing => existing.id !== project.id),
				project,
			])}
			: current,
	)
}

function setProjectDetail(
	project: ProjectResponse,
	format: 'html' | 'markdown',
) {
	queryClient.removeQueries({
		queryKey: projectKeys.detailRoot(project.id),
		predicate: query => query.queryKey[3] !== format,
	})
	queryClient.setQueryData(projectKeys.detail(project.id, format), project)
}

export function updateProjectInCache(
	projectId: number,
	updater: (project: ProjectResponse) => Project,
): void {
	queryClient.setQueriesData<ProjectListResult>(
		{queryKey: projectKeys.lists()},
		current => current
			? {
				...current,
				projects: sortProjects(current.projects.map(project =>
					project.id === projectId
						? normalizeProject({...project, ...updater(project)})
						: project,
				)),
			}
			: current,
	)
	queryClient.setQueriesData<ProjectResponse>(
		{queryKey: projectKeys.detailRoot(projectId)},
		current => current
			? normalizeProject({...current, ...updater(current)})
			: current,
	)
}

export async function cancelProjectQueries(projectId: number): Promise<void> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectKeys.lists()}),
		queryClient.cancelQueries({queryKey: projectKeys.detailRoot(projectId)}),
	])
}

export function updateProjectNavigationItemInCache(
	projectId: number,
	updater: (project: ProjectResponse) => Project,
): void {
	queryClient.setQueriesData<ProjectListResult>(
		{queryKey: projectKeys.lists()},
		current => {
			if (!current) {
				return current
			}

			const update = (project: ProjectResponse) => project.id === projectId
				? normalizeProject({...project, ...updater(project)})
				: project
			return {
				projects: sortProjects(current.projects.map(update)),
				favoriteProject: current.favoriteProject?.id === projectId
					? toNavigationItem(update(current.favoriteProject))
					: current.favoriteProject,
				savedFilterProjects: sortSavedFilters(current.savedFilterProjects.map(project =>
					toNavigationItem(update(project)),
				)),
			}
		},
	)
}

export async function createProject(
	project: ProjectWritable,
	format: 'html' | 'markdown' = 'html',
): Promise<ProjectResponse> {
	const context = getProjectMutationContext()
	await queryClient.cancelQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	const {data} = await projectsCreate({body: projectBody(project), query: {format}})
	assertProjectMutationContext(context)
	const created = normalizeProject(data)
	if (format === 'html') {
		setProjectInDefaultList(created)
	}
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	return created
}

export async function updateProject(
	{id, ...project}: UpdateProjectInput,
	format: 'html' | 'markdown' = 'html',
): Promise<ProjectResponse> {
	const context = getProjectMutationContext()
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectKeys.lists()}),
		queryClient.cancelQueries({queryKey: projectKeys.detailRoot(id)}),
	])
	assertProjectMutationContext(context)
	const previous = queryClient.getQueryData<ProjectResponse>(projectKeys.detail(id))
		?? queryClient.getQueriesData<ProjectListResult>({queryKey: projectKeys.lists()})
			.flatMap(([, result]) => result?.projects ?? [])
			.find(project => project.id === id)
	const {data} = await projectsUpdate({
		path: {id},
		body: projectBody(project),
		query: {format},
	})
	assertProjectMutationContext(context)
	const updated = normalizeProject({
		...previous,
		...data,
		max_permission: data.max_permission ?? previous?.max_permission,
		views: data.views ?? previous?.views,
	})
	if (format === 'html') {
		setProjectInDefaultList(updated)
	}
	setProjectDetail(updated, format)
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	return updated
}

export async function patchProjectFavorite(id: number, isFavorite: boolean): Promise<ProjectResponse> {
	const context = getProjectMutationContext()
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectKeys.lists()}),
		queryClient.cancelQueries({queryKey: projectKeys.detailRoot(id)}),
	])
	assertProjectMutationContext(context)
	const {data} = await patchProjectsRead({
		path: {id},
		body: [{op: 'replace', path: '/is_favorite', value: isFavorite}],
	})
	assertProjectMutationContext(context)
	const updated = normalizeProject(data)
	updateProjectInCache(id, project => ({...project, is_favorite: updated.is_favorite}))
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	return updated
}

function descendantIds(projects: readonly ProjectResponse[], projectId: number): number[] {
	const children = projects.filter(project => project.parent_project_id === projectId)
	return [
		projectId,
		...children.flatMap(project => descendantIds(projects, project.id)),
	]
}

export async function deleteProject(projectId: number): Promise<void> {
	const context = getProjectMutationContext()
	await queryClient.cancelQueries({queryKey: projectKeys.all})
	assertProjectMutationContext(context)
	await projectsDelete({path: {id: projectId}})
	assertProjectMutationContext(context)
	const ids = new Set<number>([projectId])
	queryClient.setQueriesData<ProjectListResult>(
		{queryKey: projectKeys.lists()},
		current => {
			if (!current) {
				return current
			}
			descendantIds(current.projects, projectId).forEach(id => ids.add(id))
			return {
				...current,
				projects: current.projects.filter(project => !ids.has(project.id)),
			}
		},
	)
	ids.forEach(id => queryClient.removeQueries({queryKey: projectKeys.detailRoot(id)}))
	removeProjectFromHistory({id: projectId})
}

export async function duplicateProject({
	projectId,
	parentProjectId = 0,
	duplicateShares = false,
}: DuplicateProjectInput): Promise<ProjectResponse> {
	const context = getProjectMutationContext()
	await queryClient.cancelQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	const {data} = await projectsDuplicate({
		path: {projectid: projectId},
		body: {
			parent_project_id: parentProjectId,
			duplicate_shares: duplicateShares,
		},
	})
	assertProjectMutationContext(context)
	if (!data.duplicated_project) {
		throw new Error('Project duplicate response is missing the duplicated project')
	}
	const duplicate = normalizeProject({
		...data.duplicated_project,
		max_permission: PERMISSIONS.ADMIN,
	})
	setProjectInDefaultList(duplicate)
	setProjectDetail(duplicate, 'html')
	await queryClient.invalidateQueries({queryKey: projectKeys.lists()})
	assertProjectMutationContext(context)
	return duplicate
}

export function useCreateProjectMutation() {
	return useMutation({mutationFn: (project: ProjectWritable) => createProject(project)})
}

export function useUpdateProjectMutation() {
	return useMutation({mutationFn: (project: UpdateProjectInput) => updateProject(project)})
}

export function useDeleteProjectMutation() {
	return useMutation({mutationFn: deleteProject})
}

export function useDuplicateProjectMutation() {
	return useMutation({mutationFn: duplicateProject})
}
