import {describe, it, expect, vi} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
import {ref} from 'vue'

vi.mock('@/composables/useIncludeSubprojects', () => ({
	useIncludeSubprojects: () => ({includeSubprojects: ref(false), fromView: ref(false)}),
}))

import FilterPopup from './FilterPopup.vue'

function mountPopup() {
	return shallowMount(FilterPopup, {
		props: {modelValue: {filter: '', q: '', filter_include_nulls: false, per_page: 50}},
		global: {
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('FilterPopup', () => {
	it('carries every field the filters edited to the parent', async () => {
		const wrapper = mountPopup()
		const filters = wrapper.findComponent({name: 'Filters'})

		filters.vm.$emit('update:modelValue', {sort_by: ['due_date'], order_by: ['asc'], filter: 'due_date > now', filter_include_nulls: true, s: '', filter_timezone: 'Europe/Berlin'})
		filters.vm.$emit('showResults')
		await flushPromises()

		const emitted = wrapper.emitted('update:modelValue') ?? []
		expect(emitted[emitted.length - 1][0]).toEqual({
			per_page: 50,
			sort_by: ['due_date'],
			order_by: ['asc'],
			filter: 'due_date > now',
			filter_include_nulls: true,
			filter_timezone: 'Europe/Berlin',
			q: '',
		})
	})
})
