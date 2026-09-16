import {beforeEach, describe, expect, it, vi} from 'vitest'
import {toValue, unref} from 'vue'
import {QueryClient, type MutationOptions} from '@tanstack/vue-query'

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

const requestContext = vi.hoisted(() => ({
	identity: {id: 1, type: 1} as {id: number, type: number} | null,
	sessionEpoch: 1,
	apiV2BaseUrl: 'https://identity-a.example/api/v2/',
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/helpers/auth', () => ({
	getAuthSessionEpoch: () => requestContext.sessionEpoch,
	getToken: () => null,
	getTokenIdentity: () => requestContext.identity,
}))
vi.mock('@/helpers/fetcher', () => ({
	getApiV2BaseUrl: () => requestContext.apiV2BaseUrl,
}))

import {
	deleteProjectBackgroundMutationOptions,
	projectBackgroundKeys,
	projectBackgroundQuery,
	setUnsplashProjectBackgroundMutationOptions,
	unsplashAuthor,
	unsplashBackgroundSearchQuery,
	unsplashBackgroundThumbnailQuery,
	uploadProjectBackgroundMutationOptions,
} from './projectBackgrounds'
import {normalizeProject, projectKeys, type ProjectResponse, type ProjectListResult} from './projects'
import {success, error} from '@/message'

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

function execute<TData, TVariables, TContext>(
	options: MutationOptions<TData, Error, TVariables, TContext>,
	variables: TVariables,
	client = queryClient,
) {
	return client.getMutationCache().build(client, options).execute(variables)
}

function resetMocks() {
	queryClient.clear()
	vi.restoreAllMocks()
	Object.values(sdk).forEach(mock => mock.mockReset())
	vi.mocked(success).mockReset()
	vi.mocked(error).mockReset()
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

const mutations = [
	{
		name: 'Unsplash',
		mock: sdk.projectsBackgroundUnsplashSet,
		run: () => execute(setUnsplashProjectBackgroundMutationOptions(), {projectId: 7, imageId: 'image-1'}),
		response: backgroundResponse,
		expected: {background_information: backgroundInformation, background_blur_hash: blurHash},
	},
	{
		name: 'upload',
		mock: sdk.projectsBackgroundUpload,
		run: () => execute(uploadProjectBackgroundMutationOptions(), {
			projectId: 7, file: new File(['background'], 'background.png', {type: 'image/png'}),
		}),
		response: backgroundResponse,
		expected: {background_information: backgroundInformation, background_blur_hash: blurHash},
	},
	{
		name: 'delete',
		mock: sdk.projectsBackgroundDelete,
		run: () => execute(deleteProjectBackgroundMutationOptions(), {projectId: 7}),
		response: {id: 7},
		expected: {background_information: null, background_blur_hash: ''},
	},
]

function seedProjectCaches(client = queryClient) {
	client.setQueryData<ProjectListResult>(projectKeys.list(), {
		projects: [cachedProjectWithBackground], favoriteProject: null, savedFilterProjects: [],
	})
	client.setQueryData(projectKeys.detail(7), cachedProjectWithBackground)
	client.setQueryData(projectBackgroundKeys.project(7), new Blob(['old'], {type: 'image/png'}))
}

describe('project background mutations', () => {
	beforeEach(resetMocks)

	it.each(mutations)('merges only background fields from a $name response into existing project caches', async ({mock, run, response, expected}) => {
		seedProjectCaches()
		mock.mockResolvedValue({data: response})

		await expect(run()).resolves.toEqual(expected)

		expect(queryClient.getQueryData(projectKeys.detail(7))).toEqual({...cachedProjectWithBackground, ...expected})
		expect(queryClient.getQueryData<ProjectListResult>(projectKeys.list())?.projects[0])
			.toEqual({...cachedProjectWithBackground, ...expected})
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectBackgroundKeys.project(7))?.isInvalidated).toBe(true)
		expect(sdk.projectsRead).not.toHaveBeenCalled()
		expect(success).toHaveBeenCalledOnce()
	})

	it('shapes Unsplash and upload requests for the generated client', async () => {
		const file = new File(['background'], 'photo.png', {type: 'image/png'})
		sdk.projectsBackgroundUnsplashSet.mockResolvedValue({data: backgroundResponse})
		sdk.projectsBackgroundUpload.mockResolvedValue({data: backgroundResponse})
		sdk.projectsBackgroundDelete.mockResolvedValue({data: {id: 7}})

		await execute(setUnsplashProjectBackgroundMutationOptions(), {projectId: 7, imageId: 'image-1'})
		await execute(uploadProjectBackgroundMutationOptions(), {projectId: 7, file})
		await execute(deleteProjectBackgroundMutationOptions(), {projectId: 7})

		expect(sdk.projectsBackgroundUnsplashSet).toHaveBeenCalledWith({path: {project: 7}, body: {id: 'image-1'}})
		expect(sdk.projectsBackgroundUpload).toHaveBeenCalledWith({path: {project: 7}, body: {background: file}})
		expect(sdk.projectsBackgroundDelete).toHaveBeenCalledWith({path: {project: 7}})
	})

	it.each(mutations)('does not materialize unmounted caches after $name', async ({mock, run, response}) => {
		mock.mockResolvedValue({data: response})

		await run()

		expect(queryClient.getQueryCache().getAll()).toEqual([])
		expect(sdk.projectsRead).not.toHaveBeenCalled()
	})

	it.each(mutations)('preserves cached data and settles invalidation after a failed $name', async ({mock, run}) => {
		seedProjectCaches()
		const image = queryClient.getQueryData(projectBackgroundKeys.project(7))
		mock.mockRejectedValue(new Error('Request failed'))

		await expect(run()).rejects.toThrow('Request failed')

		expect(queryClient.getQueryData(projectKeys.detail(7))).toEqual(cachedProjectWithBackground)
		expect(queryClient.getQueryData(projectBackgroundKeys.project(7))).toBe(image)
		expect(queryClient.getQueryState(projectKeys.list())?.isInvalidated).toBe(true)
		expect(queryClient.getQueryState(projectBackgroundKeys.project(7))?.isInvalidated).toBe(true)
		expect(success).not.toHaveBeenCalled()
	})

	it.each(mutations)('discards delayed $name completion after the session changes', async ({mock, run, response}) => {
		const nextProject = normalizeProject({id: 7, title: 'New session', background_information: null})
		mock.mockImplementation(async () => {
			requestContext.sessionEpoch++
			queryClient.clear()
			queryClient.setQueryData(projectKeys.detail(7), nextProject)
			return {data: response}
		})

		await expect(run()).rejects.toMatchObject({name: 'AbortError'})

		expect(queryClient.getQueryData(projectKeys.detail(7))).toEqual(nextProject)
		expect(queryClient.getQueryState(projectKeys.detail(7))?.isInvalidated).toBe(false)
		expect(success).not.toHaveBeenCalled()
		expect(error).not.toHaveBeenCalled()
	})

	it('uses the executing client without overwriting unrelated project metadata', async () => {
		const client = new QueryClient()
		seedProjectCaches(client)
		sdk.projectsBackgroundUpload.mockResolvedValue({data: {...backgroundResponse, title: 'Ignore', views: []}})

		await execute(uploadProjectBackgroundMutationOptions(), {projectId: 7, file: new Blob(['image'])}, client)

		expect(client.getQueryData<ProjectResponse>(projectKeys.detail(7)))
			.toMatchObject({title: 'Before', max_permission: 2, background_information: backgroundInformation})
		expect(queryClient.getQueryCache().getAll()).toEqual([])
		client.clear()
	})

	it('suppresses notifications for a project the caller has left', async () => {
		sdk.projectsBackgroundUnsplashSet.mockResolvedValue({data: backgroundResponse})
		await execute(setUnsplashProjectBackgroundMutationOptions(() => false), {projectId: 7, imageId: 'image-1'})

		expect(success).not.toHaveBeenCalled()
	})
})
