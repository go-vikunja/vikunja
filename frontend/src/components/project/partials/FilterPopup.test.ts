import {describe, it, expect, vi, beforeEach} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(async () => ({data: {items: [], total_pages: 1}})),
}))

vi.mock('@/client/generated', () => sdk)

import FilterPopup from './FilterPopup.vue'

function mountPopup() {
	return shallowMount(FilterPopup, {
		props: {modelValue: {filter: '', q: '', filter_include_nulls: false, per_page: 50}},
		global: {
			plugins: [createPinia(), [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('FilterPopup', () => {
	beforeEach(() => setActivePinia(createPinia()))

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
