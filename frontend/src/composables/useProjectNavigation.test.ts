import {defineComponent, h} from 'vue'
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
import SavedFilterModel from '@/models/savedFilter'

const savedFilterService = vi.hoisted(() => ({
	get: vi.fn(),
	update: vi.fn(),
}))

vi.mock('@/services/savedFilter', () => ({
	default: class {
		get = savedFilterService.get
		update = savedFilterService.update
	},
}))
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
	const promise = new Promise<T>((resolvePromise) => {
		resolve = resolvePromise
	})
	return {promise, resolve}
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
		savedFilterService.get.mockReset()
		savedFilterService.update.mockReset()
		removeToken()
		window.API_URL = 'https://identity-a.example/api/v1/'
		saveToken(token(1), false)
	})

	it('does not update or roll back a saved filter after its session changes', async () => {
		const listKey = projectKeys.list()
		const original = normalizeProject({id: -2, title: 'Filter', is_favorite: false})
		queryClient.setQueryData<ProjectListResult>(listKey, {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [original],
		})
		const get = deferred<SavedFilterModel>()
		savedFilterService.get.mockReturnValue(get.promise)
		savedFilterService.update.mockResolvedValue(undefined)
		const {navigation, wrapper} = mountProjectNavigation()

		const toggle = navigation.toggleProjectFavorite(original)
		await vi.waitFor(() => expect(savedFilterService.get).toHaveBeenCalledOnce())
		removeToken()
		saveToken(token(1), false)
		queryClient.clear()
		const current = normalizeProject({id: -2, title: 'Current filter', is_favorite: true})
		queryClient.setQueryData<ProjectListResult>(listKey, {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [current],
		})
		get.resolve(new SavedFilterModel({id: 1, title: 'Filter'}))

		await expect(toggle).rejects.toMatchObject({name: 'AbortError'})
		expect(savedFilterService.update).not.toHaveBeenCalled()
		expect(queryClient.getQueryData<ProjectListResult>(listKey)?.savedFilterProjects[0]).toEqual(current)
		wrapper.unmount()
	})
})
