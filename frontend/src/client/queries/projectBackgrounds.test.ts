import {beforeEach, describe, expect, it, vi} from 'vitest'
import {toValue, unref} from 'vue'

import {queryClient} from '@/client/queryClient'

const sdk = vi.hoisted(() => ({
	backgroundsUnsplashSearch: vi.fn(),
	backgroundsUnsplashThumb: vi.fn(),
	projectsBackgroundGet: vi.fn(),
	projectsBackgroundUnsplashSet: vi.fn(),
	projectsBackgroundUpload: vi.fn(),
	projectsBackgroundDelete: vi.fn(),
}))

const projects = vi.hoisted(() => ({
	cancelProjectQueries: vi.fn(),
	updateProjectInCache: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/client/queries/projects', () => projects)

import {
	deleteProjectBackground,
	getUnsplashAuthorInfo,
	projectBackgroundKeys,
	projectBackgroundQuery,
	setUnsplashProjectBackground,
	unsplashBackgroundSearchQuery,
	unsplashBackgroundThumbnailQuery,
	uploadProjectBackground,
} from './projectBackgrounds'

import type {Image, Project} from '@/client/generated'

const projectBefore: Project = {
	id: 7,
	title: 'Before',
	background_information: null,
	background_blur_hash: '',
	max_permission: 2,
}

const projectWithBackground: Project = {
	...projectBefore,
	background_information: {
		author: 'ada',
		author_name: 'Ada Lovelace',
	},
	background_blur_hash: 'LEHV6nWB2yk8pyo0adR*.7kCMdnj',
}

const backgroundResponse: Project = {
	id: 7,
	background_information: projectWithBackground.background_information,
	background_blur_hash: projectWithBackground.background_blur_hash,
	max_permission: null,
}

function expectProjectCacheUpdate(project: Project) {
	expect(projects.updateProjectInCache).toHaveBeenCalledOnce()
	const [projectId, updater] = projects.updateProjectInCache.mock.calls[0]
	expect(projectId).toBe(project.id)
	expect(updater(projectBefore)).toEqual(project)
}

describe('project background queries', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.restoreAllMocks()
		Object.values(sdk).forEach(mock => mock.mockReset())
		projects.updateProjectInCache.mockReset()
		projects.cancelProjectQueries.mockReset()
	})

	it('uses separate stable keys for project images and Unsplash resources', () => {
		expect(projectBackgroundKeys.all).toEqual(['project-backgrounds'])
		expect(projectBackgroundKeys.project(7)).toEqual(['project-backgrounds', 'project', 7])
		expect(projectBackgroundKeys.search('forest')).toEqual([
			'project-backgrounds',
			'unsplash',
			'search',
			'forest',
		])
		expect(projectBackgroundKeys.thumbnail('image-1')).toEqual([
			'project-backgrounds',
			'unsplash',
			'thumbnail',
			'image-1',
		])
	})

	it('fetches and caches a project background as a blob', async () => {
		const background = new Blob(['project-background'], {type: 'image/jpeg'})
		sdk.projectsBackgroundGet.mockResolvedValue({data: background})

		await expect(queryClient.fetchQuery(projectBackgroundQuery(7))).resolves.toBe(background)
		expect(sdk.projectsBackgroundGet).toHaveBeenCalledWith({path: {project: 7}})
		expect(queryClient.getQueryData(projectBackgroundKeys.project(7))).toBe(background)
	})

	it('searches Unsplash by term and page without relying on total counts', async () => {
		const image: Image = {id: 'image-1', blur_hash: 'hash'}
		sdk.backgroundsUnsplashSearch.mockResolvedValue({data: {items: [image]}})

		const options = unsplashBackgroundSearchQuery('forest')
		const result = await queryClient.fetchInfiniteQuery(options)
		const getNextPageParam = unref(toValue(options).getNextPageParam)

		expect(result.pages).toEqual([[image]])
		expect(sdk.backgroundsUnsplashSearch).toHaveBeenCalledWith({query: {q: 'forest', page: 1}})
		expect(getNextPageParam([image], [[image]], 1, [1])).toBe(2)
		expect(getNextPageParam([], [[image], []], 2, [1, 2])).toBeUndefined()
	})

	it('fetches and caches an Unsplash thumbnail as a blob', async () => {
		const thumbnail = new Blob(['thumbnail'], {type: 'image/jpeg'})
		sdk.backgroundsUnsplashThumb.mockResolvedValue({data: thumbnail})

		await expect(queryClient.fetchQuery(unsplashBackgroundThumbnailQuery('image-1'))).resolves.toBe(thumbnail)
		expect(sdk.backgroundsUnsplashThumb).toHaveBeenCalledWith({path: {image: 'image-1'}})
		expect(queryClient.getQueryData(projectBackgroundKeys.thumbnail('image-1'))).toBe(thumbnail)
	})
})

describe('Unsplash author information', () => {
	it('narrows valid provider information without changing wire casing', () => {
		expect(getUnsplashAuthorInfo({
			info: {
				author: 'ada',
				author_name: 'Ada Lovelace',
			},
		})).toEqual({
			author: 'ada',
			author_name: 'Ada Lovelace',
		})
	})

	it('rejects invalid provider information', () => {
		for (const info of [
			undefined,
			null,
			'not-an-object',
			{author: 'ada'},
			{author: 42, author_name: 'Ada Lovelace'},
		]) {
			expect(getUnsplashAuthorInfo({info})).toBeNull()
		}
	})
})

describe('project background mutations', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.restoreAllMocks()
		Object.values(sdk).forEach(mock => mock.mockReset())
		projects.updateProjectInCache.mockReset()
		projects.cancelProjectQueries.mockReset()
	})

	it('sets an Unsplash background and refreshes the cached background blob', async () => {
		sdk.projectsBackgroundUnsplashSet.mockResolvedValue({data: backgroundResponse})
		const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')

		await expect(setUnsplashProjectBackground({
			projectId: 7,
			imageId: 'image-1',
		})).resolves.toEqual(backgroundResponse)

		expect(sdk.projectsBackgroundUnsplashSet).toHaveBeenCalledWith({
			path: {project: 7},
			body: {id: 'image-1'},
		})
		expectProjectCacheUpdate(projectWithBackground)
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
	})

	it('uploads a background and refreshes the cached background blob', async () => {
		const file = new File(['background'], 'background.png', {type: 'image/png'})
		sdk.projectsBackgroundUpload.mockResolvedValue({data: backgroundResponse})
		const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')

		await expect(uploadProjectBackground({projectId: 7, file})).resolves.toEqual(backgroundResponse)

		expect(sdk.projectsBackgroundUpload).toHaveBeenCalledWith({
			path: {project: 7},
			body: {background: file},
		})
		expectProjectCacheUpdate(projectWithBackground)
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
	})

	it('deletes a background and removes the stale background blob', async () => {
		const response = {id: 7}
		sdk.projectsBackgroundDelete.mockResolvedValue({data: response})
		queryClient.setQueryData(
			projectBackgroundKeys.project(7),
			new Blob(['old-background'], {type: 'image/jpeg'}),
		)
		const removeQueries = vi.spyOn(queryClient, 'removeQueries')

		await expect(deleteProjectBackground(7)).resolves.toEqual(response)

		expect(sdk.projectsBackgroundDelete).toHaveBeenCalledWith({path: {project: 7}})
		expectProjectCacheUpdate(projectBefore)
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(removeQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
		expect(queryClient.getQueryData(projectBackgroundKeys.project(7))).toBeUndefined()
	})
})
