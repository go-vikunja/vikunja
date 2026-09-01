import {ref, type Ref} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import type {Project, ProjectView} from '@/client/generated'

const state = vi.hoisted(() => ({
	views: undefined as Ref<ProjectView[]> | undefined,
	project: undefined as Ref<Project> | undefined,
}))

vi.mock('@tanstack/vue-query', async importOriginal => {
	const {ref} = await import('vue')
	return {
		...await importOriginal<typeof import('@tanstack/vue-query')>(),
		useQuery: () => ({
			data: state.views,
			error: ref(null),
			isPending: ref(false),
		}),
	}
})

vi.mock('@/composables/useProject', async () => {
	const {ref} = await import('vue')
	return {
		useProject: () => ({
			project: state.project,
			error: ref(null),
		}),
	}
})

vi.mock('@/client/queries/projectViews', async importOriginal => ({
	...await importOriginal<typeof import('@/client/queries/projectViews')>(),
	createProjectView: vi.fn(),
	deleteProjectView: vi.fn(),
	updateProjectView: vi.fn(),
}))

vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectSettingsViews from './ProjectSettingsViews.vue'

type ProjectSettingsViewsVm = {
	showCreateForm: boolean
	newView: ProjectView & {project_id?: number}
	viewIdToDelete: number | null
	showDeleteModal: boolean
	viewToEdit: ProjectView | null
}

describe('ProjectSettingsViews', () => {
	beforeEach(() => {
		state.views = ref([])
		state.project = ref({id: 1, max_permission: 2})
	})

	it('resets route-scoped editor state when the project changes', async () => {
		const wrapper = shallowMount(ProjectSettingsViews, {
			props: {projectId: 1},
			global: {
				mocks: {$t: (key: string) => key},
				stubs: {
					CreateEdit: {template: '<div><slot /></div>'},
					Icon: true,
					Modal: true,
					draggable: true,
				},
			},
		})
		const vm = wrapper.vm as unknown as ProjectSettingsViewsVm
		const staleView: ProjectView = {
			id: 10,
			project_id: 1,
			title: 'Stale view',
			view_kind: 'list',
		}
		const staleDraft = vm.newView

		vm.showCreateForm = true
		vm.newView.title = 'Stale draft'
		vm.viewToEdit = staleView
		vm.viewIdToDelete = staleView.id ?? null
		vm.showDeleteModal = true

		await wrapper.setProps({projectId: 2})

		expect(vm.showCreateForm).toBe(false)
		expect(vm.newView).not.toBe(staleDraft)
		expect(vm.newView).toMatchObject({project_id: 2, title: ''})
		expect(vm.viewToEdit).toBeNull()
		expect(vm.viewIdToDelete).toBeNull()
		expect(vm.showDeleteModal).toBe(false)
	})
})
