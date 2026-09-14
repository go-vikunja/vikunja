import {defineComponent, effectScope, h} from 'vue'
import {mount} from '@vue/test-utils'
import {VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {
	projectKeys,
	type ProjectListResult,
} from '@/client/queries/projects'
import {removeToken, saveToken} from '@/helpers/auth'
vi.mock('vue-router', async importOriginal => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => ({push: vi.fn()}),
}))

import {useProjectNavigation} from './useProjectNavigation'

function token(id: number): string {
	const payload = btoa(JSON.stringify({id, type: 1}))
	return `header.${payload}.signature`
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
