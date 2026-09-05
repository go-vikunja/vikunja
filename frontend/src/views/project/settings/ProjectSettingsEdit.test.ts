import {defineComponent, nextTick, reactive, ref, type Ref} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {beforeEach, describe, expect, it, vi} from 'vitest'

import {normalizeProject, type ProjectResponse} from '@/client/queries/projects'

const state = vi.hoisted(() => ({
	project: undefined as Ref<ProjectResponse> | undefined,
	projects: undefined as Record<number, ProjectResponse> | undefined,
	savedParentProjectIds: [] as number[],
}))

vi.mock('@/composables/useProject', () => ({
	useProject: () => ({
		project: state.project,
		isLoading: ref(false),
		save: vi.fn(() => state.savedParentProjectIds.push(state.project!.value.parent_project_id)),
	}),
}))

vi.mock('@/composables/useProjectNavigation', () => ({
	useProjectNavigation: () => ({projects: state.projects}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({handleSetCurrentProject: vi.fn()}),
}))

vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))
vi.mock('vue-router', () => ({useRouter: () => ({back: vi.fn()})}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

const ProjectSearchStub = defineComponent({
	name: 'ProjectSearch',
	props: ['modelValue'],
	emits: ['update:modelValue'],
	template: '<div />',
})
const SlotStub = defineComponent({template: '<div><slot /></div>'})

import ProjectSettingsEdit from './ProjectSettingsEdit.vue'

describe('ProjectSettingsEdit', () => {
	beforeEach(() => {
		state.project = ref(normalizeProject({id: 101, parent_project_id: 100}))
		state.projects = reactive({})
		state.savedParentProjectIds = []
	})

	it('preserves an inaccessible parent when saving without changing it', async () => {
		const wrapper = shallowMount(ProjectSettingsEdit, {
			props: {projectId: 101},
			global: {
				stubs: {
					CreateEdit: SlotStub,
					FormField: SlotStub,
					ProjectSearch: ProjectSearchStub,
				},
				mocks: {$t: (key: string) => key, $router: {push: vi.fn()}},
				directives: {focus: () => {}, tooltip: () => {}},
			},
		})

		await (wrapper.vm as unknown as {save: () => Promise<void>}).save()

		expect(state.savedParentProjectIds).toEqual([100])
		expect(state.project!.value.parent_project_id).toBe(100)
	})

	it('selects the parent when navigation projects load after project detail', async () => {
		const wrapper = shallowMount(ProjectSettingsEdit, {
			props: {projectId: 101},
			global: {
				stubs: {
					CreateEdit: SlotStub,
					FormField: SlotStub,
					ProjectSearch: ProjectSearchStub,
				},
				mocks: {$t: (key: string) => key, $router: {push: vi.fn()}},
				directives: {focus: () => {}, tooltip: () => {}},
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

	it('resets the parent selection when the route reuses the component', async () => {
		state.projects![100] = normalizeProject({id: 100, title: 'First Parent'})
		const wrapper = shallowMount(ProjectSettingsEdit, {
			props: {projectId: 101},
			global: {
				stubs: {
					CreateEdit: SlotStub,
					FormField: SlotStub,
					ProjectSearch: ProjectSearchStub,
				},
				mocks: {$t: (key: string) => key, $router: {push: vi.fn()}},
				directives: {focus: () => {}, tooltip: () => {}},
			},
		})

		const projectSearch = wrapper.findComponent({name: 'ProjectSearch'})
		projectSearch.vm.$emit('update:modelValue', normalizeProject({id: 200, title: 'Changed Parent'}))
		await nextTick()
		expect(projectSearch.props('modelValue')).toMatchObject({id: 200})

		await wrapper.setProps({projectId: 102})
		state.project!.value = normalizeProject({id: 102, parent_project_id: 300})
		await nextTick()
		expect(projectSearch.props('modelValue')).toBeNull()

		state.projects![300] = normalizeProject({id: 300, title: 'Second Parent'})
		await nextTick()
		expect(projectSearch.props('modelValue')).toMatchObject({
			id: 300,
			title: 'Second Parent',
		})
	})

	it('rehydrates a shared parent when cached project details replace each other', async () => {
		state.projects![100] = normalizeProject({id: 100, title: 'Shared Parent'})
		const wrapper = shallowMount(ProjectSettingsEdit, {
			props: {projectId: 101},
			global: {
				stubs: {
					CreateEdit: SlotStub,
					FormField: SlotStub,
					ProjectSearch: ProjectSearchStub,
				},
				mocks: {$t: (key: string) => key, $router: {push: vi.fn()}},
				directives: {focus: () => {}, tooltip: () => {}},
			},
		})

		const projectSearch = wrapper.findComponent({name: 'ProjectSearch'})
		projectSearch.vm.$emit('update:modelValue', normalizeProject({id: 200, title: 'Changed Parent'}))
		await nextTick()

		const routeChange = wrapper.setProps({projectId: 102})
		state.project!.value = normalizeProject({id: 102, parent_project_id: 100})
		await routeChange

		expect(projectSearch.props('modelValue')).toMatchObject({
			id: 100,
			title: 'Shared Parent',
		})
		await (wrapper.vm as unknown as {save: () => Promise<void>}).save()
		expect(state.savedParentProjectIds).toEqual([100])
	})
})
