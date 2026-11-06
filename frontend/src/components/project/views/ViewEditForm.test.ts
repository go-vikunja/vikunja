import {shallowMount} from '@vue/test-utils'
import {defineComponent} from 'vue'
import {describe, expect, it, vi} from 'vitest'

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

const FilterEditorStub = defineComponent({
	name: 'FilterEditor',
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
			stubs: {FilterEditor: FilterEditorStub},
			mocks: {$t: (key: string) => key},
			directives: {focus: () => {}},
		},
	})
}

describe('ViewEditForm', () => {
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
