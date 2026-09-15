import {mutationOptions, queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'

import {
	projectViewsCreate,
	projectViewsDelete,
	projectViewsList,
	projectViewsUpdate,
} from '@/client/generated'
import type {
	ProjectView,
	ProjectViewsListData,
	ProjectViewWritable,
} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {mapProjectNavigationItem, projectKeys, type ProjectListResult, type ProjectResponse} from './projects'
import {i18n} from '@/i18n'
import {error, success} from '@/message'

type ProjectViewListArgs = Pick<NonNullable<ProjectViewsListData['query']>, 'q'>

export type ProjectViewDraft = Required<Omit<ProjectViewWritable, 'bucket_configuration'>> & {
	bucket_configuration: NonNullable<ProjectViewWritable['bucket_configuration']>
}

export type ProjectViewUpdate = Omit<ProjectViewDraft, 'filter'> & Pick<ProjectViewWritable, 'filter'>

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

export function createProjectViewUpdate(view: Partial<ProjectViewWritable>): ProjectViewUpdate {
	return {
		...createProjectViewDraft(view),
		filter: view.filter,
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

export function refreshProjectViews(projectId: number, args: ProjectViewListArgs = {}) {
	return queryClient.fetchQuery({...projectViewsQuery(projectId, args), staleTime: 0})
}

function replaceView(views: readonly ProjectView[] | null | undefined, view: ProjectView): ProjectView[] {
	return sortProjectViewsByPosition([
		...(views ?? []).filter(existing => existing.id !== view.id),
		view,
	])
}

function setProjectViewInCache(client: QueryClient, projectId: number, view: ProjectView) {
	client.setQueryData<ProjectView[]>(
		projectViewKeys.list(projectId),
		current => current ? replaceView(current, view) : current,
	)
	updateEmbeddedViews(client, projectId, views => replaceView(views, view))
}

function updateEmbeddedViews(client: QueryClient, projectId: number, update: (views: ProjectView[]) => ProjectView[]) {
	const updateProject = (project: ProjectResponse) => ({...project, views: update(project.views)})
	client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
		current ? mapProjectNavigationItem(current, projectId, updateProject) : current,
	)
	client.setQueryData<ProjectResponse>(projectKeys.detail(projectId), current =>
		current ? updateProject(current) : current,
	)
}

async function snapshotProjectViews(client: QueryClient, projectId: number) {
	const request = captureClientRequestContext()
	await Promise.all([
		client.cancelQueries({queryKey: projectViewKeys.lists(projectId)}),
		client.cancelQueries({queryKey: projectKeys.list()}),
		client.cancelQueries({queryKey: projectKeys.detail(projectId)}),
	])
	assertClientRequestContext(request)
	return {
		request,
		previous: [
			...client.getQueriesData<ProjectView[]>({queryKey: projectViewKeys.lists(projectId)}),
			...client.getQueriesData<ProjectListResult>({queryKey: projectKeys.list()}),
			...client.getQueriesData<ProjectResponse>({queryKey: projectKeys.detail(projectId)}),
		],
	}
}

type ViewSnapshot = Awaited<ReturnType<typeof snapshotProjectViews>>
type ShouldNotify = (input: {projectId: number}) => boolean

function restoreProjectViews(client: QueryClient, snapshot: ViewSnapshot | undefined) {
	if (!snapshot || !isClientRequestContextCurrent(snapshot.request)) {
		return
	}
	for (const [key, previous] of snapshot.previous) {
		if (previous) {
			client.setQueryData(key, previous)
		}
	}
}

function settledProjectViews() {
	return async (
		_data: unknown,
		_error: unknown,
		{projectId}: {projectId: number},
		context: {request: ReturnType<typeof captureClientRequestContext>} | undefined,
		{client}: {client: QueryClient},
	) => {
		if (context && isClientRequestContextCurrent(context.request)) {
			await Promise.all([
				client.invalidateQueries({queryKey: projectViewKeys.lists(projectId)}),
				client.invalidateQueries({queryKey: projectKeys.list()}),
				client.invalidateQueries({queryKey: projectKeys.detail(projectId)}),
			])
			assertClientRequestContext(context.request)
		}
	}
}

export function createProjectViewMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		onMutate: () => ({request: captureClientRequestContext()}),
		mutationFn: async ({projectId, view}: CreateProjectViewInput) => {
			const request = captureClientRequestContext()
			const {data} = await projectViewsCreate({path: {project: projectId}, body: view})
			assertClientRequestContext(request)
			return data
		},
		onSuccess: (created, input, context, {client}) => {
			assertClientRequestContext(context.request)
			setProjectViewInCache(client, input.projectId, created)
			if (shouldNotify(input)) {
				success({message: i18n.global.t('project.views.createSuccess')})
			}
		},
		onError: (cause, input, context) => {
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(input)) {
				error(cause)
			}
		},
		onSettled: settledProjectViews(),
	})
}

export function updateProjectViewMutationOptions(successMessage?: string, shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		mutationFn: async ({projectId, viewId, view}: UpdateProjectViewInput) => {
			const request = captureClientRequestContext()
			const {data} = await projectViewsUpdate({path: {project: projectId, view: viewId}, body: view})
			assertClientRequestContext(request)
			return data
		},
		onMutate: async ({projectId, viewId, view}, {client}) => {
			const snapshot = await snapshotProjectViews(client, projectId)
			const update = (views: ProjectView[]) => sortProjectViewsByPosition(views.map(existing =>
				existing.id === viewId ? {...existing, ...view} : existing,
			))
			client.setQueriesData<ProjectView[]>({queryKey: projectViewKeys.lists(projectId)}, current =>
				current ? update(current) : current,
			)
			updateEmbeddedViews(client, projectId, update)
			return snapshot
		},
		onError: (cause, input, context, {client}) => {
			restoreProjectViews(client, context)
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(input)) {
				error(cause)
			}
		},
		onSuccess: (updated, input, context, {client}) => {
			assertClientRequestContext(context.request)
			setProjectViewInCache(client, input.projectId, updated)
			if (shouldNotify(input)) {
				success({message: successMessage ?? i18n.global.t('project.views.updateSuccess')})
			}
		},
		onSettled: settledProjectViews(),
	})
}

export function deleteProjectViewMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		mutationFn: async ({projectId, viewId}: DeleteProjectViewInput) => {
			const request = captureClientRequestContext()
			await projectViewsDelete({path: {project: projectId, view: viewId}})
			assertClientRequestContext(request)
		},
		onMutate: async ({projectId, viewId}, {client}) => {
			const snapshot = await snapshotProjectViews(client, projectId)
			const remove = (views: ProjectView[]) => sortProjectViewsByPosition(views.filter(view => view.id !== viewId))
			client.setQueriesData<ProjectView[]>({queryKey: projectViewKeys.lists(projectId)}, current =>
				current ? remove(current) : current,
			)
			updateEmbeddedViews(client, projectId, remove)
			return snapshot
		},
		onError: (cause, input, context, {client}) => {
			restoreProjectViews(client, context)
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(input)) {
				error(cause)
			}
		},
		onSettled: settledProjectViews(),
	})
}

export function useCreateProjectViewMutation(shouldNotify?: ShouldNotify) {
	return useMutation(createProjectViewMutationOptions(shouldNotify))
}

export function useUpdateProjectViewMutation(successMessage?: string, shouldNotify?: ShouldNotify) {
	return useMutation(updateProjectViewMutationOptions(successMessage, shouldNotify))
}

export function useDeleteProjectViewMutation(shouldNotify?: ShouldNotify) {
	return useMutation(deleteProjectViewMutationOptions(shouldNotify))
}
