import {infiniteQueryOptions, queryOptions, useMutation} from '@tanstack/vue-query'

import {
	backgroundsUnsplashSearch,
	backgroundsUnsplashThumb,
	projectsBackgroundDelete,
	projectsBackgroundGet,
	projectsBackgroundUnsplashSet,
	projectsBackgroundUpload,
} from '@/client/generated'
import type {Image, Project} from '@/client/generated'
import {queryClient} from '@/client/queryClient'
import {cancelProjectQueries, getCachedProject, projectQuery, updateProjectInCache} from '@/client/queries/projects'
import type {ProjectResponse} from '@/client/queries/projects'
import {assertClientRequestContext, captureClientRequestContext} from '@/client/requestContext'

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

// Wire responses are sparse (delete returns only an id), so hand back the merged cache entry.
async function mutateProjectBackground(
	projectId: number,
	request: () => Promise<Project>,
): Promise<ProjectResponse> {
	const context = captureClientRequestContext()
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectBackgroundKeys.project(projectId)}),
		cancelProjectQueries(projectId),
	])
	assertClientRequestContext(context)

	const data = await request()
	assertClientRequestContext(context)

	const background = backgroundFromResponse(data)
	updateProjectInCache(projectId, current => ({...current, ...background}))
	if (background.background_information === null) {
		queryClient.removeQueries({queryKey: projectBackgroundKeys.project(projectId)})
	} else {
		await queryClient.invalidateQueries({queryKey: projectBackgroundKeys.project(projectId)})
	}

	const merged = getCachedProject(projectId) ?? await queryClient.fetchQuery(projectQuery(projectId))
	assertClientRequestContext(context)

	return merged
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
	)
}

export function deleteProjectBackground(projectId: number): Promise<ProjectResponse> {
	return mutateProjectBackground(
		projectId,
		async () => {
			const {data} = await projectsBackgroundDelete({path: {project: projectId}})
			return data
		},
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
