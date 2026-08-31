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
import {cancelProjectQueries, updateProjectInCache} from '@/client/queries/projects'

export const projectBackgroundKeys = {
	all: ['project-backgrounds'] as const,
	project: (projectId: number) => ['project-backgrounds', 'project', projectId] as const,
	unsplash: () => ['project-backgrounds', 'unsplash'] as const,
	search: (query: string) => ['project-backgrounds', 'unsplash', 'search', query] as const,
	thumbnail: (imageId: string) => ['project-backgrounds', 'unsplash', 'thumbnail', imageId] as const,
}

export function projectBackgroundQuery(projectId: number) {
	return queryOptions({
		queryKey: projectBackgroundKeys.project(projectId),
		queryFn: async () => {
			const {data} = await projectsBackgroundGet({path: {project: projectId}})
			return data as Blob
		},
	})
}

export function unsplashBackgroundSearchQuery(query: string) {
	return infiniteQueryOptions({
		queryKey: projectBackgroundKeys.search(query),
		queryFn: async ({pageParam}) => {
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
			return data as Blob
		},
	})
}

type UnsplashAuthorInfo = {
	author: string
	author_name: string
}

export function getUnsplashAuthorInfo(image: Pick<Image, 'info'>): UnsplashAuthorInfo | null {
	if (typeof image.info !== 'object' || image.info === null) {
		return null
	}

	const info = image.info as Record<string, unknown>
	if (typeof info.author !== 'string' || typeof info.author_name !== 'string') {
		return null
	}

	return {
		author: info.author,
		author_name: info.author_name,
	}
}

async function updateBackgroundCache(projectId: number, project: Project): Promise<Project> {
	updateProjectInCache(projectId, current => ({
		...current,
		background_information: project.background_information,
		background_blur_hash: project.background_blur_hash,
	}))
	await queryClient.invalidateQueries({queryKey: projectBackgroundKeys.project(projectId)})
	return project
}

export async function setUnsplashProjectBackground({
	projectId,
	imageId,
}: {
	projectId: number
	imageId: string
}): Promise<Project> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectBackgroundKeys.project(projectId)}),
		cancelProjectQueries(projectId),
	])
	const {data} = await projectsBackgroundUnsplashSet({
		path: {project: projectId},
		body: {id: imageId},
	})
	return updateBackgroundCache(projectId, data)
}

export async function uploadProjectBackground({
	projectId,
	file,
}: {
	projectId: number
	file: Blob | File
}): Promise<Project> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectBackgroundKeys.project(projectId)}),
		cancelProjectQueries(projectId),
	])
	const {data} = await projectsBackgroundUpload({
		path: {project: projectId},
		body: {background: file},
	})
	return updateBackgroundCache(projectId, data)
}

export async function deleteProjectBackground(projectId: number): Promise<Project> {
	await Promise.all([
		queryClient.cancelQueries({queryKey: projectBackgroundKeys.project(projectId)}),
		cancelProjectQueries(projectId),
	])
	const {data} = await projectsBackgroundDelete({path: {project: projectId}})
	updateProjectInCache(projectId, current => ({
		...current,
		background_information: null,
		background_blur_hash: '',
	}))
	queryClient.removeQueries({queryKey: projectBackgroundKeys.project(projectId)})
	return data
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
