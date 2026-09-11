import {shallowMount} from '@vue/test-utils'
import {describe, expect, it, vi} from 'vitest'

import type {ProjectResponse} from '@/client/queries/projects'

const {currentProject} = vi.hoisted(() => ({
	currentProject: {value: null as ProjectResponse | null},
}))

vi.mock('@/composables/useCurrentProject', () => ({
	useCurrentProject: () => ({currentProject, isPending: {value: false}}),
}))

vi.mock('@/stores/viewFilters', () => ({
	useViewFiltersStore: () => ({getViewQuery: () => ({})}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

vi.mock('@/composables/useTitle', () => ({useTitle: () => {}}))

import ProjectWrapper from './ProjectWrapper.vue'

function mountWrapper(projectId: number) {
	return shallowMount(ProjectWrapper, {
		props: {
			isLoadingProject: false,
			projectId,
			viewId: 10,
		},
		global: {
			mocks: {$t: (key: string) => key},
			stubs: {RouterLink: true},
		},
	})
}

describe('ProjectWrapper', () => {
	// The task detail modal renders a project view as its backdrop, where the project id can
	// resolve to NaN – the query then has no project to hand out.
	it('renders when the project is not loaded', () => {
		currentProject.value = null

		const wrapper = mountWrapper(NaN)

		expect(wrapper.find('.switch-view-container').exists()).toBe(true)
		expect(wrapper.find('.switch-view').exists()).toBe(false)
	})

	it('renders the view switcher for a loaded project', () => {
		currentProject.value = {
			id: 1,
			title: 'Test',
			views: [
				{id: 10, title: 'List', view_kind: 'list'},
				{id: 11, title: 'Kanban', view_kind: 'kanban'},
			],
		} as ProjectResponse

		const wrapper = mountWrapper(1)

		expect(wrapper.find('.switch-view').exists()).toBe(true)
	})
})
