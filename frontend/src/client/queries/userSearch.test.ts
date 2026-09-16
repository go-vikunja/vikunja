import {defineComponent, h, ref} from 'vue'
import {flushPromises, mount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, onTestFinished, vi} from 'vitest'

const sdk = vi.hoisted(() => ({usersSearch: vi.fn(), projectsUsersSearch: vi.fn()}))
vi.mock('@/client/generated', () => sdk)

import {projectUserSearchQuery, userSearchQuery} from './userSearch'
import {useProjectUserSearch, useUserSearch} from '@/composables/useUserSearch'

beforeEach(() => vi.resetAllMocks())

describe('user search', () => {
	it('uses the v2 global search argument and keeps wire fields intact', async () => {
		const user = {id: 1, username: 'sam', bot_owner_id: 2, created: '2026-01-01T00:00:00Z'}
		sdk.usersSearch.mockResolvedValue({data: {items: [user]}})
		const client = new QueryClient()
		expect(await client.fetchQuery(userSearchQuery('sam'))).toEqual([user])
		expect(sdk.usersSearch).toHaveBeenCalledWith({query: {q: 'sam'}, signal: expect.any(AbortSignal)})
		client.clear()
	})

	it('searches project users independently, including the empty preload', async () => {
		sdk.projectsUsersSearch.mockResolvedValue({data: {items: null}})
		const client = new QueryClient()
		expect(await client.fetchQuery(projectUserSearchQuery(7, ''))).toEqual([])
		expect(sdk.projectsUsersSearch).toHaveBeenCalledWith({path: {project: 7}, query: {q: ''}, signal: expect.any(AbortSignal)})
		expect(sdk.usersSearch).not.toHaveBeenCalled()
		client.clear()
	})
})

function withQueryClient<T>(setup: () => T): T {
	let result!: T
	const queryClient = new QueryClient()
	const wrapper = mount(defineComponent({
		setup() {
			result = setup()
			return () => h('div')
		},
	}), {
		global: {
			plugins: [[VueQueryPlugin, {queryClient}]],
		},
	})
	onTestFinished(() => {
		wrapper.unmount()
		queryClient.clear()
	})
	return result
}

describe('useUserSearch', () => {
	it('does not fetch an empty search', async () => {
		const {users} = withQueryClient(() => useUserSearch(''))
		await flushPromises()
		expect(sdk.usersSearch).not.toHaveBeenCalled()
		expect(users.value).toEqual([])
	})

	it('keeps previous results while typing and clears them with the input', async () => {
		const sam = {id: 1, username: 'sam'}
		sdk.usersSearch.mockResolvedValueOnce({data: {items: [sam]}})
		const search = ref('sa')
		const {users} = withQueryClient(() => useUserSearch(search))
		await flushPromises()
		expect(users.value).toEqual([sam])

		sdk.usersSearch.mockReturnValueOnce(new Promise(() => {}))
		search.value = 'sam'
		await flushPromises()
		expect(users.value).toEqual([sam])

		search.value = ''
		await flushPromises()
		expect(users.value).toEqual([])
		expect(sdk.usersSearch).toHaveBeenCalledTimes(2)
	})
})

describe('useProjectUserSearch', () => {
	it('does not fetch until enabled', async () => {
		const {users} = withQueryClient(() => useProjectUserSearch(7, '', false))
		await flushPromises()
		expect(sdk.projectsUsersSearch).not.toHaveBeenCalled()
		expect(users.value).toEqual([])
	})

	it('does not fetch without a project', async () => {
		const {users} = withQueryClient(() => useProjectUserSearch(0, 'sam', true))
		await flushPromises()
		expect(sdk.projectsUsersSearch).not.toHaveBeenCalled()
		expect(users.value).toEqual([])
	})

	it('preloads project users with an empty search once enabled', async () => {
		const sam = {id: 1, username: 'sam'}
		sdk.projectsUsersSearch.mockResolvedValue({data: {items: [sam]}})
		const enabled = ref(false)
		const {users} = withQueryClient(() => useProjectUserSearch(7, '', enabled))
		await flushPromises()
		expect(sdk.projectsUsersSearch).not.toHaveBeenCalled()

		enabled.value = true
		await flushPromises()
		expect(sdk.projectsUsersSearch).toHaveBeenCalledWith({path: {project: 7}, query: {q: ''}, signal: expect.any(AbortSignal)})
		expect(users.value).toEqual([sam])
	})

	it('returns no users once disabled', async () => {
		sdk.projectsUsersSearch.mockResolvedValue({data: {items: [{id: 1, username: 'sam'}]}})
		const enabled = ref(true)
		const {users} = withQueryClient(() => useProjectUserSearch(7, 'sam', enabled))
		await flushPromises()
		expect(users.value).toHaveLength(1)

		enabled.value = false
		await flushPromises()
		expect(users.value).toEqual([])
	})

	it('does not show the previous project\'s users while the new project is loading', async () => {
		const sam = {id: 1, username: 'sam'}
		sdk.projectsUsersSearch.mockResolvedValueOnce({data: {items: [sam]}})
		const projectId = ref(7)
		const {users} = withQueryClient(() => useProjectUserSearch(projectId, '', true))
		await flushPromises()
		expect(users.value).toEqual([sam])

		sdk.projectsUsersSearch.mockReturnValueOnce(new Promise(() => {}))
		projectId.value = 8
		await flushPromises()
		expect(users.value).toEqual([])
	})
})
