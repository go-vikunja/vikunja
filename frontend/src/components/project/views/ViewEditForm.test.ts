import {shallowMount} from '@vue/test-utils'
import {defineComponent, nextTick} from 'vue'
import {beforeEach, describe, expect, it, vi} from 'vitest'

const labelState = vi.hoisted(() => ({
	labels: undefined as {value: Array<{id?: number, title?: string}>} | undefined,
	isPending: undefined as {value: boolean} | undefined,
}))

vi.mock('@/composables/useLabels', async () => {
	const {ref} = await import('vue')
	labelState.labels = ref([])
	labelState.isPending = ref(true)

	return {
		useLabels: () => ({
			labels: labelState.labels,
			isPending: labelState.isPending,
			getLabelById: (id: number) => labelState.labels?.value.find(label => label.id === id),
			getLabelByExactTitle: (title: string) => labelState.labels?.value.find(label => label.title === title),
		}),
	}
})

vi.mock('@/composables/useProjectNavigation', () => ({
	useProjectNavigation: () => ({
		projects: {},
		findProjectByExactname: () => null,
	}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

const FilterInputStub = defineComponent({
	name: 'FilterInput',
	props: {modelValue: {type: String, default: ''}},
	emits: ['update:modelValue'],
	template: '<div />',
})

import ViewEditForm from './ViewEditForm.vue'
import type {ProjectView} from '@/client/generated'

function mountForm(modelValue: ProjectView = {
	id: 1,
	project_id: 1,
	title: 'List',
	view_kind: 'kanban',
	bucket_configuration_mode: 'filter',
	filter: {filter: 'labels = 1', filter_include_nulls: false},
	bucket_configuration: [{
		title: 'Bucket',
		filter: {filter: 'labels = 1', filter_include_nulls: false},
	}],
}) {
	return shallowMount(ViewEditForm, {
		props: {
			modelValue,
		},
		global: {
			stubs: {FilterInput: FilterInputStub},
			mocks: {$t: (key: string) => key},
			directives: {focus: () => {}},
		},
	})
}

describe('ViewEditForm', () => {
	beforeEach(() => {
		labelState.labels!.value = []
		labelState.isPending!.value = true
	})

	it('resolves saved label ids after labels finish loading', async () => {
		const wrapper = mountForm()
		expect(wrapper.findAllComponents(FilterInputStub).map(input => input.props('modelValue')))
			.toEqual(['labels = 1', 'labels = 1'])

		labelState.labels!.value = [{id: 1, title: 'Work'}]
		labelState.isPending!.value = false
		await nextTick()

		expect(wrapper.findAllComponents(FilterInputStub).map(input => input.props('modelValue')))
			.toEqual(['labels = Work', 'labels = Work'])
	})

	it('keeps a filter edited before labels finish loading', async () => {
		const wrapper = mountForm()
		wrapper.findAllComponents(FilterInputStub)[0].vm.$emit('update:modelValue', 'labels = Personal')
		await nextTick()

		labelState.labels!.value = [{id: 1, title: 'Work'}]
		labelState.isPending!.value = false
		await nextTick()

		expect(wrapper.findAllComponents(FilterInputStub)[0].props('modelValue')).toBe('labels = Personal')
	})

	it('preserves model sorting when saving', async () => {
		const modelValue: ProjectView = {
			id: 1,
			project_id: 1,
			title: 'List',
			view_kind: 'kanban',
			bucket_configuration_mode: 'filter',
			filter: {
				sort_by: ['done', 'id'],
				order_by: ['asc', 'desc'],
				filter: 'done = false',
				filter_include_nulls: true,
				s: '',
			},
			bucket_configuration: [{
				title: 'Bucket',
				filter: {
					sort_by: ['position'],
					order_by: ['desc'],
					filter: 'done = false',
					filter_include_nulls: false,
					s: '',
				},
			}],
		}
		const wrapper = mountForm(modelValue)

		await wrapper.find('form').trigger('submit')

		const emitted = wrapper.emitted('update:modelValue')
		const saved = emitted?.[emitted.length - 1]?.[0] as ProjectView
		expect(saved.filter).toMatchObject({
			sort_by: ['done', 'id'],
			order_by: ['asc', 'desc'],
		})
		expect(saved.bucket_configuration?.[0].filter).toMatchObject({
			sort_by: ['position'],
			order_by: ['desc'],
		})
	})

	it('does not save while focus moves to another control in the form', async () => {
		const wrapper = mountForm({
			id: 1,
			project_id: 1,
			title: 'Board',
			view_kind: 'kanban',
			bucket_configuration_mode: 'manual',
		})
		const manualMode = wrapper.find<HTMLInputElement>('input[value="manual"]')
		const filterMode = wrapper.find<HTMLInputElement>('input[value="filter"]')

		await manualMode.trigger('focusout', {relatedTarget: filterMode.element})

		expect(wrapper.emitted('update:modelValue')).toBeUndefined()
		await filterMode.trigger('click')
		expect(filterMode.element.checked).toBe(true)
		expect(wrapper.emitted('update:modelValue')).toBeUndefined()
	})

	it('saves when focus leaves the form without explicit save buttons', async () => {
		const wrapper = mountForm()

		await wrapper.find('form').trigger('focusout', {relatedTarget: document.body})

		expect(wrapper.emitted('update:modelValue')).toHaveLength(1)
	})
})
