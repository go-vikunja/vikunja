import {useMutation, type QueryClient} from '@tanstack/vue-query'

import {
	projectViewsCreate,
	projectViewsDelete,
	projectViewsUpdate,
} from '@/client/generated'
import type {
	ProjectView,
	ProjectViewWritable,
} from '@/client/generated'
import {contextMutationOptions} from './contextMutation'
import {mapProjectNavigationItem, projectKeys, type ProjectListResult, type ProjectResponse} from './projects'
import {i18n} from '@/i18n'

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

function replaceView(views: readonly ProjectView[] | null | undefined, view: ProjectView): ProjectView[] {
	return sortProjectViewsByPosition([
		...(views ?? []).filter(existing => existing.id !== view.id),
		view,
	])
}

// Every project response embeds its views, so they live in the project caches instead of their own.
function updateProjectViews(client: QueryClient, projectId: number, update: (views: ProjectView[]) => ProjectView[]) {
	const updateProject = (project: ProjectResponse) => ({...project, views: update(project.views)})
	client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
		current ? mapProjectNavigationItem(current, projectId, updateProject) : current,
	)
	client.setQueryData<ProjectResponse>(projectKeys.detail(projectId), current =>
		current ? updateProject(current) : current,
	)
}

type ShouldNotify = (input: {projectId: number}) => boolean

function invalidateProjectViews(client: QueryClient, projectId: number) {
	return Promise.all([
		// Refetching the list reloads every page for one project's views, so it is only marked stale.
		client.invalidateQueries({queryKey: projectKeys.list(), refetchType: 'none'}),
		client.invalidateQueries({queryKey: projectKeys.detail(projectId)}),
	])
}

export function createProjectViewMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return contextMutationOptions({
		mutationFn: async ({projectId, view}: CreateProjectViewInput) => {
			const {data} = await projectViewsCreate({path: {project: projectId}, body: view})
			return data
		},
		onSuccess: (created, input, client) => {
			updateProjectViews(client, input.projectId, views => replaceView(views, created))
		},
		onSettled: ({projectId}, client) => invalidateProjectViews(client, projectId),
		successMessage: (_created, input) => shouldNotify(input) ? i18n.global.t('project.views.createSuccess') : undefined,
		toastError: shouldNotify,
	})
}

export function updateProjectViewMutationOptions(successMessage?: string, shouldNotify: ShouldNotify = () => true) {
	return contextMutationOptions({
		mutationFn: async ({projectId, viewId, view}: UpdateProjectViewInput) => {
			const {data} = await projectViewsUpdate({path: {project: projectId, view: viewId}, body: view})
			return data
		},
		optimistic: {
			queryKeys: ({projectId}) => [projectKeys.list(), projectKeys.detail(projectId)],
			update: ({projectId, viewId, view}, client) => {
				updateProjectViews(client, projectId, views => sortProjectViewsByPosition(views.map(existing =>
					existing.id === viewId ? {...existing, ...view} : existing,
				)))
			},
		},
		onSuccess: (updated, input, client) => {
			updateProjectViews(client, input.projectId, views => replaceView(views, updated))
		},
		onSettled: ({projectId}, client) => invalidateProjectViews(client, projectId),
		successMessage: (_updated, input) => shouldNotify(input) ? successMessage ?? i18n.global.t('project.views.updateSuccess') : undefined,
		toastError: shouldNotify,
	})
}

export function deleteProjectViewMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return contextMutationOptions({
		mutationFn: async ({projectId, viewId}: DeleteProjectViewInput) => {
			await projectViewsDelete({path: {project: projectId, view: viewId}})
		},
		optimistic: {
			queryKeys: ({projectId}) => [projectKeys.list(), projectKeys.detail(projectId)],
			update: ({projectId, viewId}, client) => {
				updateProjectViews(client, projectId, views => views.filter(view => view.id !== viewId))
			},
		},
		onSettled: ({projectId}, client) => invalidateProjectViews(client, projectId),
		toastError: shouldNotify,
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
