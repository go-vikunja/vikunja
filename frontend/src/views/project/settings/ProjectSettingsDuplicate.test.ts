import {defineComponent, nextTick, reactive, ref, type Ref} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {normalizeProject, type ProjectResponse} from '@/client/queries/projects'

const state = vi.hoisted(() => ({
	project: undefined as Ref<ProjectResponse> | undefined,
	projects: undefined as Record<number, ProjectResponse> | undefined,
}))

vi.mock('@/composables/useProject', () => ({
	useProject: () => ({
		project: state.project,
		isLoading: ref(false),
		duplicateProject: vi.fn(),
	}),
}))

vi.mock('@/composables/useProjectNavigation', () => ({
	useProjectNavigation: () => ({projects: state.projects}),
}))

vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))
vi.mock('@/message', () => ({success: vi.fn()}))
vi.mock('vue-router', () => ({useRoute: () => ({params: {projectId: '101'}})}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

const ProjectSearchStub = defineComponent({
	name: 'ProjectSearch',
	props: ['modelValue'],
	template: '<div />',
})
const SlotStub = defineComponent({template: '<div><slot /></div>'})

import ProjectSettingsDuplicate from './ProjectSettingsDuplicate.vue'

describe('ProjectSettingsDuplicate', () => {
	beforeEach(() => {
		state.project = ref(normalizeProject({id: 101, parent_project_id: 100}))
		state.projects = reactive({})
	})

	it('selects the parent when navigation projects load after project detail', async () => {
		const wrapper = shallowMount(ProjectSettingsDuplicate, {
			global: {
				stubs: {
					CreateEdit: SlotStub,
					FancyCheckbox: SlotStub,
					ProjectSearch: ProjectSearchStub,
				},
				mocks: {$t: (key: string) => key},
			},
		})

		const projectSearch = wrapper.findComponent({name: 'ProjectSearch'})
		expect(projectSearch.props('modelValue') == null).toBe(true)

		state.projects![100] = normalizeProject({id: 100, title: 'Parent Project'})
		await nextTick()

		expect(projectSearch.props('modelValue')).toMatchObject({
			id: 100,
			title: 'Parent Project',
		})
	})
})
