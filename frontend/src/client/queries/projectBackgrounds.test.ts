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
	projectsRead: vi.fn(),
}))

const projects = vi.hoisted(() => ({
	cancelProjectQueries: vi.fn(),
	updateProjectInCache: vi.fn(),
}))

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number, type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/client/queries/projects', async importOriginal => ({
	...await importOriginal<typeof import('@/client/queries/projects')>(),
	...projects,
}))
vi.mock('@/helpers/auth', () => ({
	getAuthSessionEpoch: () => requestContext.sessionEpoch,
	getToken: () => null,
	getTokenIdentity: () => requestContext.identity,
}))
vi.mock('@/helpers/fetcher', () => ({
	getApiV2BaseUrl: () => requestContext.apiV2BaseUrl,
}))

import {
	deleteProjectBackground,
	projectBackgroundKeys,
	projectBackgroundQuery,
	setUnsplashProjectBackground,
	unsplashAuthor,
	unsplashBackgroundSearchQuery,
	unsplashBackgroundThumbnailQuery,
	uploadProjectBackground,
} from './projectBackgrounds'
import {normalizeProject, projectKeys} from './projects'

import type {Project} from '@/client/generated'

const projectBefore: Project = {
	id: 7,
	title: 'Before',
	background_information: null,
	background_blur_hash: '',
	max_permission: 2,
}

const backgroundInformation = {
	author: 'ada',
	author_name: 'Ada Lovelace',
}
const blurHash = 'LEHV6nWB2yk8pyo0adR*.7kCMdnj'

const backgroundResponse: Project = {
	id: 7,
	background_information: backgroundInformation,
	background_blur_hash: blurHash,
	max_permission: null,
}

const cachedProjectWithBackground = normalizeProject({
	...projectBefore,
	background_information: backgroundInformation,
	background_blur_hash: blurHash,
})

function expectProjectCacheUpdate(background: Partial<Project>) {
	expect(projects.updateProjectInCache).toHaveBeenCalledOnce()
	const [projectId, updater] = projects.updateProjectInCache.mock.calls[0]
	expect(projectId).toBe(7)
	expect(updater(projectBefore)).toEqual({...projectBefore, ...background})
}

function resetMocks() {
	queryClient.clear()
	vi.restoreAllMocks()
	Object.values(sdk).forEach(mock => mock.mockReset())
	projects.updateProjectInCache.mockReset()
	projects.cancelProjectQueries.mockReset()
	requestContext.identity = {id: 1, type: 1}
	requestContext.sessionEpoch = 1
	requestContext.apiV2BaseUrl = 'https://identity-a.example/api/v2/'
}

describe('project background queries', () => {
	beforeEach(resetMocks)

	it('rejects background responses which are not images', async () => {
		sdk.projectsBackgroundGet.mockResolvedValue({data: {message: 'not an image'}})
		sdk.backgroundsUnsplashThumb.mockResolvedValue({data: 'not an image'})

		await expect(queryClient.fetchQuery({...projectBackgroundQuery(7), retry: false}))
			.rejects.toThrowError('Background response was not an image')
		await expect(queryClient.fetchQuery({...unsplashBackgroundThumbnailQuery('image-1'), retry: false}))
			.rejects.toThrowError('Background response was not an image')
	})

	it('returns raw Unsplash search pages and paginates while pages are not empty', async () => {
		const items = [
			{id: 'image-1', blur_hash: 'hash', info: backgroundInformation},
			{blur_hash: 'no-id', info: backgroundInformation},
			{id: 'image-2'},
		]
		sdk.backgroundsUnsplashSearch.mockResolvedValue({data: {items}})

		const options = unsplashBackgroundSearchQuery('forest')
		const result = await queryClient.fetchInfiniteQuery(options)
		const getNextPageParam = unref(toValue(options).getNextPageParam)
		const firstPage = result.pages[0]

		expect(firstPage).toEqual(items)
		expect(sdk.backgroundsUnsplashSearch).toHaveBeenCalledWith({query: {q: 'forest', page: 1}})
		expect(getNextPageParam(firstPage, [firstPage], 1, [1])).toBe(2)
		expect(getNextPageParam([], [firstPage, []], 2, [1, 2])).toBeUndefined()
	})

	it('returns an empty page when the Unsplash response has no items', async () => {
		sdk.backgroundsUnsplashSearch.mockResolvedValue({data: {}})

		const result = await queryClient.fetchInfiniteQuery(unsplashBackgroundSearchQuery('forest'))

		expect(result.pages).toEqual([[]])
	})

	it.each([
		{name: 'a complete author info object', info: backgroundInformation, expected: backgroundInformation},
		{name: 'missing info', info: undefined, expected: null},
		{name: 'null info', info: null, expected: null},
		{name: 'info without an author name', info: {author: 'ada'}, expected: null},
		{name: 'info with non-string fields', info: {author: 42, author_name: 'Ada Lovelace'}, expected: null},
	])('narrows $name to Unsplash attribution fields', ({info, expected}) => {
		expect(unsplashAuthor(info)).toEqual(expected)
	})
})

describe('project background mutations', () => {
	beforeEach(resetMocks)

	it('sets an Unsplash background and returns the merged cached project', async () => {
		sdk.projectsBackgroundUnsplashSet.mockResolvedValue({data: backgroundResponse})
		queryClient.setQueryData(projectKeys.detail(7), cachedProjectWithBackground)
		const cancelQueries = vi.spyOn(queryClient, 'cancelQueries')
		const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')

		await expect(setUnsplashProjectBackground({
			projectId: 7,
			imageId: 'image-1',
		})).resolves.toBe(cachedProjectWithBackground)

		expect(sdk.projectsBackgroundUnsplashSet).toHaveBeenCalledWith({
			path: {project: 7},
			body: {id: 'image-1'},
		})
		expect(cancelQueries).toHaveBeenCalledWith({queryKey: projectBackgroundKeys.project(7)})
		expect(cancelQueries.mock.invocationCallOrder[0])
			.toBeLessThan(sdk.projectsBackgroundUnsplashSet.mock.invocationCallOrder[0])
		expectProjectCacheUpdate({
			background_information: backgroundInformation,
			background_blur_hash: blurHash,
		})
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
	})

	it('uploads a background and returns the merged cached project', async () => {
		const file = new File(['background'], 'background.png', {type: 'image/png'})
		sdk.projectsBackgroundUpload.mockResolvedValue({data: backgroundResponse})
		queryClient.setQueryData(projectKeys.detail(7), cachedProjectWithBackground)
		const cancelQueries = vi.spyOn(queryClient, 'cancelQueries')
		const invalidateQueries = vi.spyOn(queryClient, 'invalidateQueries')

		await expect(uploadProjectBackground({projectId: 7, file})).resolves.toBe(cachedProjectWithBackground)

		expect(sdk.projectsBackgroundUpload).toHaveBeenCalledWith({
			path: {project: 7},
			body: {background: file},
		})
		expect(cancelQueries).toHaveBeenCalledWith({queryKey: projectBackgroundKeys.project(7)})
		expect(cancelQueries.mock.invocationCallOrder[0])
			.toBeLessThan(sdk.projectsBackgroundUpload.mock.invocationCallOrder[0])
		expectProjectCacheUpdate({
			background_information: backgroundInformation,
			background_blur_hash: blurHash,
		})
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(invalidateQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
	})

	it('deletes a background, removes the stale blob and returns the merged cached project', async () => {
		const cachedProject = normalizeProject(projectBefore)
		sdk.projectsBackgroundDelete.mockResolvedValue({data: {id: 7}})
		queryClient.setQueryData(projectKeys.detail(7), cachedProject)
		queryClient.setQueryData(
			projectBackgroundKeys.project(7),
			new Blob(['old-background'], {type: 'image/jpeg'}),
		)
		const cancelQueries = vi.spyOn(queryClient, 'cancelQueries')
		const removeQueries = vi.spyOn(queryClient, 'removeQueries')

		await expect(deleteProjectBackground(7)).resolves.toBe(cachedProject)

		expect(sdk.projectsBackgroundDelete).toHaveBeenCalledWith({path: {project: 7}})
		expect(cancelQueries).toHaveBeenCalledWith({queryKey: projectBackgroundKeys.project(7)})
		expect(cancelQueries.mock.invocationCallOrder[0])
			.toBeLessThan(sdk.projectsBackgroundDelete.mock.invocationCallOrder[0])
		expectProjectCacheUpdate({background_information: null, background_blur_hash: ''})
		expect(projects.cancelProjectQueries).toHaveBeenCalledWith(7)
		expect(removeQueries).toHaveBeenCalledWith({
			queryKey: projectBackgroundKeys.project(7),
		})
		expect(queryClient.getQueryData(projectBackgroundKeys.project(7))).toBeUndefined()
	})

	it('fetches the project when it is not cached', async () => {
		sdk.projectsBackgroundDelete.mockResolvedValue({data: {id: 7}})
		sdk.projectsRead.mockResolvedValue({data: projectBefore})

		await expect(deleteProjectBackground(7)).resolves.toEqual(normalizeProject(projectBefore))
		expect(sdk.projectsRead).toHaveBeenCalledWith({path: {id: 7}, query: {format: 'html'}})
	})

	it.each([
		{
			name: 'delete',
			mock: sdk.projectsBackgroundDelete,
			run: () => deleteProjectBackground(7),
		},
		{
			name: 'upload',
			mock: sdk.projectsBackgroundUpload,
			run: () => uploadProjectBackground({
				projectId: 7,
				file: new File(['background'], 'background.png', {type: 'image/png'}),
			}),
		},
	])('keeps the cached background when the $name request fails', async ({mock, run}) => {
		const background = new Blob(['old-background'], {type: 'image/jpeg'})
		queryClient.setQueryData(projectBackgroundKeys.project(7), background)
		mock.mockRejectedValue(new Error('request failed'))

		await expect(run()).rejects.toThrowError('request failed')

		expect(projects.updateProjectInCache).not.toHaveBeenCalled()
		expect(queryClient.getQueryData(projectBackgroundKeys.project(7))).toBe(background)
	})

	it.each([
		{
			name: 'delete',
			mock: sdk.projectsBackgroundDelete,
			run: () => deleteProjectBackground(7),
		},
		{
			name: 'upload',
			mock: sdk.projectsBackgroundUpload,
			run: () => uploadProjectBackground({
				projectId: 7,
				file: new File(['background'], 'background.png', {type: 'image/png'}),
			}),
		},
		{
			name: 'unsplash set',
			mock: sdk.projectsBackgroundUnsplashSet,
			run: () => setUnsplashProjectBackground({projectId: 7, imageId: 'image-1'}),
		},
	])('discards a delayed $name completion after the identity changed', async ({mock, run}) => {
		mock.mockImplementation(async () => {
			requestContext.identity = {id: 2, type: 1}
			return {data: backgroundResponse}
		})

		await expect(run()).rejects.toThrowError('Client request context changed')
		expect(projects.updateProjectInCache).not.toHaveBeenCalled()
	})
})
