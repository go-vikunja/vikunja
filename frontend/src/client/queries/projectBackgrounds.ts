import {infiniteQueryOptions, keepPreviousData, mutationOptions, queryOptions, useMutation, type QueryClient} from '@tanstack/vue-query'

import {
	backgroundsUnsplashSearch,
	backgroundsUnsplashThumb,
	projectsBackgroundDelete,
	projectsBackgroundGet,
	projectsBackgroundUnsplashSet,
	projectsBackgroundUpload,
} from '@/client/generated'
import type {Image, Project} from '@/client/generated'
import {mapProjectNavigationItem, projectKeys} from '@/client/queries/projects'
import type {ProjectListResult, ProjectResponse} from '@/client/queries/projects'
import {assertClientRequestContext, captureClientRequestContext, isClientRequestContextCurrent} from '@/client/requestContext'
import {i18n} from '@/i18n'
import {error, success} from '@/message'

export const projectBackgroundKeys = {
	all: ['project-backgrounds'] as const,
	project: (projectId: number) => ['project-backgrounds', 'project', projectId] as const,
	search: (query: string) => ['project-backgrounds', 'unsplash', 'search', query] as const,
	thumbnail: (imageId: string) => ['project-backgrounds', 'unsplash', 'thumbnail', imageId] as const,
}

function asBackgroundBlob(data: unknown): Blob {
	if (!(data instanceof Blob)) {
		throw new Error('Background response was not an image')
	}
	return data
}

export function unsplashAuthor(info: unknown): {author: string, author_name: string} | null {
	if (
		typeof info !== 'object' ||
		info === null ||
		!('author' in info) ||
		typeof info.author !== 'string' ||
		!('author_name' in info) ||
		typeof info.author_name !== 'string'
	) {
		return null
	}

	return {author: info.author, author_name: info.author_name}
}

export function projectBackgroundQuery(projectId: number) {
	return queryOptions({
		queryKey: projectBackgroundKeys.project(projectId),
		queryFn: async () => {
			const {data} = await projectsBackgroundGet({path: {project: projectId}})
			return asBackgroundBlob(data)
		},
	})
}

export function unsplashBackgroundSearchQuery(query: string) {
	return infiniteQueryOptions({
		queryKey: projectBackgroundKeys.search(query),
		queryFn: async ({pageParam}): Promise<Image[]> => {
			const {data} = await backgroundsUnsplashSearch({
				query: {q: query, page: pageParam},
			})
			return data.items ?? []
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage, _pages, lastPageParam) =>
			lastPage.length > 0 ? lastPageParam + 1 : undefined,
		placeholderData: keepPreviousData,
	})
}

export function unsplashBackgroundThumbnailQuery(imageId: string) {
	return queryOptions({
		queryKey: projectBackgroundKeys.thumbnail(imageId),
		queryFn: async () => {
			const {data} = await backgroundsUnsplashThumb({path: {image: imageId}})
			return asBackgroundBlob(data)
		},
	})
}

type BackgroundFields = Required<Pick<Project, 'background_information' | 'background_blur_hash'>>
type ShouldNotify = (projectId: number) => boolean
type BackgroundMutationContext = {request: ReturnType<typeof captureClientRequestContext>}

function backgroundFromResponse(project: Project): BackgroundFields {
	return {
		background_information: project.background_information ?? null,
		background_blur_hash: project.background_blur_hash ?? '',
	}
}

function backgroundMutationCallbacks<TInput>(
	projectIdFromInput: (input: TInput) => number,
	shouldNotify: ShouldNotify,
	successMessage: () => string,
) {
	return {
		onMutate: () => ({request: captureClientRequestContext()}),
		onSuccess: (background: BackgroundFields, input: TInput, context: BackgroundMutationContext, {client}: {client: QueryClient}) => {
			assertClientRequestContext(context.request)
			const projectId = projectIdFromInput(input)
			const update = (project: ProjectResponse) => ({...project, ...background})
			client.setQueryData<ProjectListResult>(projectKeys.list(), current =>
				current ? mapProjectNavigationItem(current, projectId, update) : current,
			)
			client.setQueryData<ProjectResponse>(projectKeys.detail(projectId), current =>
				current ? update(current) : current,
			)
			if (shouldNotify(projectId)) {
				success({message: successMessage()})
			}
		},
		onError: (cause: Error, input: TInput, context: BackgroundMutationContext | undefined) => {
			if (context && isClientRequestContextCurrent(context.request) && shouldNotify(projectIdFromInput(input))) {
				error(cause)
			}
		},
		onSettled: async (
			_data: unknown,
			_error: unknown,
			input: TInput,
			context: BackgroundMutationContext | undefined,
			{client}: {client: QueryClient},
		) => {
			if (context && isClientRequestContextCurrent(context.request)) {
				const projectId = projectIdFromInput(input)
				await Promise.all([
					client.invalidateQueries({queryKey: projectKeys.list()}),
					client.invalidateQueries({queryKey: projectKeys.detail(projectId)}),
					client.invalidateQueries({queryKey: projectBackgroundKeys.project(projectId), exact: true}),
				])
			}
		},
	}
}

export function setUnsplashProjectBackgroundMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		...backgroundMutationCallbacks(
			(input: {projectId: number; imageId: string}) => input.projectId,
			shouldNotify,
			() => i18n.global.t('project.background.success'),
		),
		mutationFn: async ({projectId, imageId}: {projectId: number; imageId: string}) => {
			const request = captureClientRequestContext()
			const {data} = await projectsBackgroundUnsplashSet({
				path: {project: projectId},
				body: {id: imageId},
			})
			assertClientRequestContext(request)
			return backgroundFromResponse(data)
		},
	})
}

export function uploadProjectBackgroundMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		...backgroundMutationCallbacks(
			(input: {projectId: number; file: Blob | File}) => input.projectId,
			shouldNotify,
			() => i18n.global.t('project.background.success'),
		),
		mutationFn: async ({projectId, file}: {projectId: number; file: Blob | File}) => {
			const request = captureClientRequestContext()
			const {data} = await projectsBackgroundUpload({
				path: {project: projectId},
				body: {background: file},
			})
			assertClientRequestContext(request)
			return backgroundFromResponse(data)
		},
	})
}

export function deleteProjectBackgroundMutationOptions(shouldNotify: ShouldNotify = () => true) {
	return mutationOptions({
		...backgroundMutationCallbacks(
			(projectId: number) => projectId,
			shouldNotify,
			() => i18n.global.t('project.background.removeSuccess'),
		),
		mutationFn: async (projectId: number) => {
			const request = captureClientRequestContext()
			const {data} = await projectsBackgroundDelete({path: {project: projectId}})
			assertClientRequestContext(request)
			return backgroundFromResponse(data)
		},
	})
}

export function useSetUnsplashProjectBackgroundMutation(shouldNotify?: ShouldNotify) {
	return useMutation(setUnsplashProjectBackgroundMutationOptions(shouldNotify))
}

export function useUploadProjectBackgroundMutation(shouldNotify?: ShouldNotify) {
	return useMutation(uploadProjectBackgroundMutationOptions(shouldNotify))
}

export function useDeleteProjectBackgroundMutation(shouldNotify?: ShouldNotify) {
	return useMutation(deleteProjectBackgroundMutationOptions(shouldNotify))
}
