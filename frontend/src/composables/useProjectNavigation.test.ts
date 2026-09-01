import {defineComponent, effectScope, h} from 'vue'
import {mount} from '@vue/test-utils'
import {VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {
	normalizeProject,
	projectKeys,
	type ProjectListResult,
} from '@/client/queries/projects'
import {removeToken, saveToken} from '@/helpers/auth'

const savedFilterQueries = vi.hoisted(() => ({
	patchSavedFilterFavorite: vi.fn(),
}))

vi.mock('@/client/queries/savedFilters', () => savedFilterQueries)
vi.mock('vue-router', async importOriginal => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => ({push: vi.fn()}),
}))

import {useProjectNavigation} from './useProjectNavigation'

function token(id: number): string {
	const payload = btoa(JSON.stringify({id, type: 1}))
	return `header.${payload}.signature`
}

function deferred<T>() {
	let resolve!: (value: T) => void
	let reject!: (reason: unknown) => void
	const promise = new Promise<T>((resolvePromise, rejectPromise) => {
		resolve = resolvePromise
		reject = rejectPromise
	})
	return {promise, resolve, reject}
}

function mountProjectNavigation() {
	let navigation!: ReturnType<typeof useProjectNavigation>
	const wrapper = mount(defineComponent({
		setup() {
			navigation = useProjectNavigation()
			return () => h('div')
		},
	}), {
		global: {
			plugins: [[VueQueryPlugin, {queryClient}]],
		},
	})
	return {navigation, wrapper}
}

describe('useProjectNavigation', () => {
	beforeEach(() => {
		queryClient.clear()
		savedFilterQueries.patchSavedFilterFavorite.mockReset()
		removeToken()
		window.API_URL = 'https://identity-a.example/api/v1/'
		saveToken(token(1), false)
	})

	it('can be created outside a Vue injection context', () => {
		queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [],
		})
		const scope = effectScope()

		expect(() => scope.run(() => useProjectNavigation())).not.toThrow()
		scope.stop()
	})

	it('rolls back the optimistic favorite only while the request context is current', async () => {
		const listKey = projectKeys.list()
		const original = normalizeProject({id: -2, title: 'Filter', is_favorite: false})
		queryClient.setQueryData<ProjectListResult>(listKey, {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [original],
		})
		const patch = deferred<never>()
		savedFilterQueries.patchSavedFilterFavorite.mockReturnValue(patch.promise)
		const {navigation, wrapper} = mountProjectNavigation()

		const toggle = navigation.toggleProjectFavorite(original)
		await vi.waitFor(() => expect(savedFilterQueries.patchSavedFilterFavorite).toHaveBeenCalledOnce())
		removeToken()
		saveToken(token(1), false)
		queryClient.clear()
		const current = normalizeProject({id: -2, title: 'Current filter', is_favorite: true})
		queryClient.setQueryData<ProjectListResult>(listKey, {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [current],
		})
		patch.reject(new DOMException('Client request context changed', 'AbortError'))

		await expect(toggle).rejects.toMatchObject({name: 'AbortError'})
		expect(queryClient.getQueryData<ProjectListResult>(listKey)?.savedFilterProjects[0]).toEqual(current)
		wrapper.unmount()
	})

	it('shares a single projects query observer across consumers', () => {
		queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [],
		})
		const first = mountProjectNavigation()
		const second = mountProjectNavigation()

		const observersCount = () => queryClient.getQueryCache()
			.find({queryKey: projectKeys.list()})
			?.getObserversCount() ?? 0
		expect(observersCount()).toBe(1)

		first.wrapper.unmount()
		expect(observersCount()).toBe(1)

		second.wrapper.unmount()
		expect(observersCount()).toBe(0)
	})
})
