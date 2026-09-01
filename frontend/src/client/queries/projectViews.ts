import {queryOptions} from '@tanstack/vue-query'

import {
	projectViewsCreate,
	projectViewsDelete,
	projectViewsList,
	projectViewsRead,
	projectViewsUpdate,
} from '@/client/generated'
import type {
	ProjectView,
	ProjectViewsListData,
	ProjectViewWritable,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {cancelProjectQueries, updateProjectInCache} from './projects'

type ProjectViewListArgs = Pick<NonNullable<ProjectViewsListData['query']>, 'q'>

export type ProjectViewDraft = Required<Omit<ProjectViewWritable, 'bucket_configuration'>> & {
	bucket_configuration: NonNullable<ProjectViewWritable['bucket_configuration']>
}

export type CreateProjectViewInput = {
	projectId: number
	view: ProjectViewWritable
}

export type UpdateProjectViewInput = CreateProjectViewInput & {
	viewId: number
}

export type DeleteProjectViewInput = {
	projectId: number
	viewId: number
}

export const projectViewKeys = {
	all: ['project-views'] as const,
	lists: (projectId: number) => [...projectViewKeys.all, 'list', projectId] as const,
	list: (projectId: number, args: ProjectViewListArgs = {}) => [
		...projectViewKeys.lists(projectId),
		args,
	] as const,
	details: (projectId: number) => [...projectViewKeys.all, 'detail', projectId] as const,
	detail: (projectId: number, viewId: number) => [
		...projectViewKeys.details(projectId),
		viewId,
	] as const,
}

export function sortProjectViewsByPosition(views: ProjectView[]): ProjectView[] {
	return [...views].sort((a, b) => (a.position ?? 0) - (b.position ?? 0))
}

export function createProjectViewDraft(view: Partial<ProjectViewWritable> = {}): ProjectViewDraft {
	return {
		title: view.title ?? '',
		view_kind: view.view_kind ?? 'list',
		filter: {
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
			filter: 'done = false',
			filter_include_nulls: true,
			s: '',
			...view.filter,
		},
		position: view.position ?? 0,
		bucket_configuration_mode: view.bucket_configuration_mode ?? 'manual',
		bucket_configuration: view.bucket_configuration ?? [],
		default_bucket_id: view.default_bucket_id ?? 0,
		done_bucket_id: view.done_bucket_id ?? 0,
	}
}

async function fetchProjectViews(projectId: number, args: ProjectViewListArgs): Promise<ProjectView[]> {
	const {data} = await projectViewsList({
		path: {project: projectId},
		query: {page: 1, per_page: 1000, ...args},
	})
	return data.items ?? []
}

export function projectViewsQuery(projectId: number, args: ProjectViewListArgs = {}) {
	return queryOptions({
		queryKey: projectViewKeys.list(projectId, args),
		queryFn: () => fetchProjectViews(projectId, args),
		select: sortProjectViewsByPosition,
	})
}

export function projectViewQuery(projectId: number, viewId: number) {
	return queryOptions({
		queryKey: projectViewKeys.detail(projectId, viewId),
		queryFn: async () => {
			const {data} = await projectViewsRead({path: {project: projectId, view: viewId}})
			return data
		},
	})
}

function replaceView(views: readonly ProjectView[] | null | undefined, view: ProjectView): ProjectView[] {
	return sortProjectViewsByPosition([
		...(views ?? []).filter(existing => existing.id !== view.id),
		view,
	])
}

function setProjectViewInCache(projectId: number, view: ProjectView) {
	queryClient.setQueryData<ProjectView[]>(
		projectViewKeys.list(projectId),
		current => current ? replaceView(current, view) : current,
	)
	if (typeof view.id !== 'undefined') {
		queryClient.setQueryData(projectViewKeys.detail(projectId, view.id), view)
	}
	updateProjectInCache(projectId, project => ({
		...project,
		views: replaceView(project.views, view),
	}))
}

export async function createProjectView({projectId, view}: CreateProjectViewInput): Promise<ProjectView> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectViewKeys.lists(projectId)}),
		cancelProjectQueries(projectId),
	])
	const {data} = await projectViewsCreate({path: {project: projectId}, body: view})
	setProjectViewInCache(projectId, data)
	await queryClient.invalidateQueries({queryKey: projectViewKeys.lists(projectId)})
	return data
}

export async function updateProjectView({
	projectId,
	viewId,
	view,
}: UpdateProjectViewInput): Promise<ProjectView> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectViewKeys.lists(projectId)}),
		queryClient.cancelQueries({queryKey: projectViewKeys.detail(projectId, viewId)}),
		cancelProjectQueries(projectId),
	])
	const {data} = await projectViewsUpdate({
		path: {project: projectId, view: viewId},
		body: view,
	})
	setProjectViewInCache(projectId, data)
	await queryClient.invalidateQueries({queryKey: projectViewKeys.lists(projectId)})
	return data
}

export async function deleteProjectView({projectId, viewId}: DeleteProjectViewInput): Promise<void> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectViewKeys.lists(projectId)}),
		queryClient.cancelQueries({queryKey: projectViewKeys.detail(projectId, viewId)}),
		cancelProjectQueries(projectId),
	])
	await projectViewsDelete({path: {project: projectId, view: viewId}})
	queryClient.setQueryData<ProjectView[]>(
		projectViewKeys.list(projectId),
		current => current
			? sortProjectViewsByPosition(current.filter(view => view.id !== viewId))
			: current,
	)
	queryClient.removeQueries({queryKey: projectViewKeys.detail(projectId, viewId)})
	updateProjectInCache(projectId, project => ({
		...project,
		views: sortProjectViewsByPosition((project.views ?? []).filter(view => view.id !== viewId)),
	}))
	await queryClient.invalidateQueries({queryKey: projectViewKeys.lists(projectId)})
}
