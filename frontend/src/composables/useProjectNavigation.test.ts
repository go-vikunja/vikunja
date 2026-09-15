import {defineComponent, effectScope, h} from 'vue'
import {mount} from '@vue/test-utils'
import {VueQueryPlugin} from '@tanstack/vue-query'
import {beforeEach, describe, expect, it, onTestFinished, vi} from 'vitest'

import {queryClient} from '@/client/queryClient'
import {
	normalizeProject,
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

function navigationWith(list: Partial<ProjectListResult>) {
	queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
		projects: [],
		favoriteProject: null,
		savedFilterProjects: [],
		...list,
	})
	const scope = effectScope()
	onTestFinished(() => scope.stop())
	return scope.run(() => useProjectNavigation())!
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

	it('derives roots, children, ancestors, and effective drag parents', () => {
		const projects = [
			normalizeProject({id: 1, title: 'Root', position: 200}),
			normalizeProject({id: 2, title: 'Child', parent_project_id: 1, position: 300}),
			normalizeProject({id: 3, title: 'Orphan', parent_project_id: 99, position: 100}),
			normalizeProject({id: 4, title: 'Archived', is_archived: true, position: 400}),
			normalizeProject({id: 5, title: 'Grandchild', parent_project_id: 2, position: 50}),
			normalizeProject({id: 6, title: 'Sibling', parent_project_id: 1, position: 100}),
		]
		const navigation = navigationWith({projects})

		expect(navigation.notArchivedRootProjects.map(project => project.id)).toEqual([3, 1])
		expect(navigation.getChildProjects(1).map(project => project.id)).toEqual([6, 2])
		expect(navigation.getAncestors(projects[4]).map(project => project.id)).toEqual([1, 2, 5])
		expect(navigation.getAncestors(projects[2]).map(project => project.id)).toEqual([3])
		expect(navigation.getEffectiveParentProjectId(projects[2], 0)).toBe(99)
		expect(navigation.getEffectiveParentProjectId(projects[2], 1)).toBe(1)
		expect(navigation.getEffectiveParentProjectId(projects[1], 0)).toBe(0)
	})

	it('searches projects and saved filters case-insensitively', () => {
		const navigation = navigationWith({
			projects: [
				normalizeProject({id: 1, title: 'Root'}),
				normalizeProject({id: 2, title: 'Child', description: 'nested work'}),
				normalizeProject({id: 4, title: 'Archived', is_archived: true}),
			],
			savedFilterProjects: [normalizeProject({id: -2, title: 'Nested filter'})],
		})

		expect(navigation.searchProject('NEST').map(project => project.id)).toEqual([2])
		expect(navigation.searchProject('archived')).toEqual([])
		expect(navigation.searchProject('archived', true).map(project => project.id)).toEqual([4])
		expect(navigation.searchProject('')).toEqual([])
		expect(navigation.searchSavedFilter('nest').map(project => project.id)).toEqual([-2])
		expect(navigation.searchProjectAndFilter('nest').map(project => project.id)).toEqual([2, -2])
	})

	it('combines favorite pseudo, project, and saved-filter navigation items', () => {
		const navigation = navigationWith({
			projects: [
				normalizeProject({id: 1, title: 'Favorite project', is_favorite: true, position: 100}),
				normalizeProject({id: 2, title: 'Archived favorite', is_favorite: true, is_archived: true, position: 200}),
			],
			favoriteProject: normalizeProject({id: -1, title: 'Favorites', is_favorite: true, position: -1}),
			savedFilterProjects: [
				normalizeProject({id: -2, title: 'Favorite filter', is_favorite: true}),
				normalizeProject({id: -3, title: 'Other filter'}),
			],
		})

		expect(navigation.favoriteProjects.map(project => project.id)).toEqual([-1, -2, 1])
	})
})
