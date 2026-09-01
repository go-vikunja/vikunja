import {infiniteQueryOptions, queryOptions, useMutation} from '@tanstack/vue-query'

import {
	backgroundsUnsplashSearch,
	backgroundsUnsplashThumb,
	projectsBackgroundDelete,
	projectsBackgroundGet,
	projectsBackgroundUnsplashSet,
	projectsBackgroundUpload,
} from '@/client/generated'
import type {Project} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {cancelProjectQueries, normalizeProject, projectKeys, updateProjectInCache} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import {assertClientRequestContext, captureClientRequestContext} from '@/client/requestContext'

export const projectBackgroundKeys = {
	all: ['project-backgrounds'] as const,
	project: (projectId: number) => ['project-backgrounds', 'project', projectId] as const,
	search: (query: string) => ['project-backgrounds', 'unsplash', 'search', query] as const,
	thumbnail: (imageId: string) => ['project-backgrounds', 'unsplash', 'thumbnail', imageId] as const,
}

export type UnsplashSearchImage = {
	id: string
	blur_hash: string
	author: string
	author_name: string
}

function asBackgroundBlob(data: unknown): Blob {
	if (!(data instanceof Blob)) {
		throw new Error('Background response was not an image')
	}
	return data
}

function isUnsplashAuthorInfo(info: unknown): info is {author: string, author_name: string} {
	return typeof info === 'object' &&
		info !== null &&
		'author' in info &&
		typeof info.author === 'string' &&
		'author_name' in info &&
		typeof info.author_name === 'string'
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
		queryFn: async ({pageParam}): Promise<UnsplashSearchImage[]> => {
			const {data} = await backgroundsUnsplashSearch({
				query: {q: query, page: pageParam},
			})
			return (data.items ?? []).flatMap(image => {
				if (typeof image.id !== 'string') {
					return []
				}

				const info = isUnsplashAuthorInfo(image.info)
					? image.info
					: {author: '', author_name: ''}
				return [{
					id: image.id,
					blur_hash: image.blur_hash ?? '',
					author: info.author,
					author_name: info.author_name,
				}]
			})
		},
		initialPageParam: 1,
		getNextPageParam: (lastPage, _pages, lastPageParam) =>
			lastPage.length > 0 ? lastPageParam + 1 : undefined,
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

type BackgroundFields = {
	background_information: unknown
	background_blur_hash: string
}

// Wire responses are sparse (delete returns only an id), so hand back the merged cache entry.
async function mutateProjectBackground(
	projectId: number,
	request: () => Promise<Project>,
	toBackground: (project: Project) => BackgroundFields,
	cachedBackground: 'invalidate' | 'remove',
): Promise<ProjectResponse> {
	const context = captureClientRequestContext()
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectBackgroundKeys.project(projectId)}),
		cancelProjectQueries(projectId),
	])
	assertClientRequestContext(context)

	const data = await request()
	assertClientRequestContext(context)

	const background = toBackground(data)
	updateProjectInCache(projectId, current => ({...current, ...background}))
	if (cachedBackground === 'remove') {
		queryClient.removeQueries({queryKey: projectBackgroundKeys.project(projectId)})
	} else {
		await queryClient.invalidateQueries({queryKey: projectBackgroundKeys.project(projectId)})
	}
	assertClientRequestContext(context)

	return queryClient.getQueryData<ProjectResponse>(projectKeys.detail(projectId))
		?? normalizeProject({...data, ...background})
}

function backgroundFromResponse(project: Project): BackgroundFields {
	return {
		background_information: project.background_information ?? null,
		background_blur_hash: project.background_blur_hash ?? '',
	}
}

export function setUnsplashProjectBackground({
	projectId,
	imageId,
}: {
	projectId: number
	imageId: string
}): Promise<ProjectResponse> {
	return mutateProjectBackground(
		projectId,
		async () => {
			const {data} = await projectsBackgroundUnsplashSet({
				path: {project: projectId},
				body: {id: imageId},
			})
			return data
		},
		backgroundFromResponse,
		'invalidate',
	)
}

export function uploadProjectBackground({
	projectId,
	file,
}: {
	projectId: number
	file: Blob | File
}): Promise<ProjectResponse> {
	return mutateProjectBackground(
		projectId,
		async () => {
			const {data} = await projectsBackgroundUpload({
				path: {project: projectId},
				body: {background: file},
			})
			return data
		},
		backgroundFromResponse,
		'invalidate',
	)
}

export function deleteProjectBackground(projectId: number): Promise<ProjectResponse> {
	return mutateProjectBackground(
		projectId,
		async () => {
			const {data} = await projectsBackgroundDelete({path: {project: projectId}})
			return data
		},
		() => ({background_information: null, background_blur_hash: ''}),
		'remove',
	)
}

export function useSetUnsplashProjectBackgroundMutation() {
	return useMutation({mutationFn: setUnsplashProjectBackground})
}

export function useUploadProjectBackgroundMutation() {
	return useMutation({mutationFn: uploadProjectBackground})
}

export function useDeleteProjectBackgroundMutation() {
	return useMutation({mutationFn: deleteProjectBackground})
}
