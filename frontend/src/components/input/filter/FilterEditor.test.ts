import {flushPromises, shallowMount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'

const settings = vi.hoisted(() => ({mode: 'simple' as 'simple' | 'advanced'}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => ({
	session: null,
	settings: {frontend_settings: {filter_editor_mode: settings.mode}},
})}))
vi.mock('vue-i18n', async (importOriginal) => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))
import FilterEditor from './FilterEditor.vue'

function mountEditor(query: string) {
	return shallowMount(FilterEditor, {
		props: {modelValue: query},
		global: {
			renderStubDefaultSlot: true,
			plugins: [[VueQueryPlugin, {queryClient: new QueryClient()}]],
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('FilterEditor', () => {
	beforeEach(() => { settings.mode = 'simple' })
	it('opens canonical filters in Simple and emits API queries immediately', async () => {
		const wrapper = mountEditor('done = false')
		expect(wrapper.find('[role="tab"][aria-selected="true"]').text()).toBe('filters.modes.simple')
		wrapper.findComponent({name: 'SimpleFilterForm'}).vm.$emit('update:modelValue', {
			showDone: false,
			due: {
				range: null,
				includeOverdue: false,
			},
			labelIds: [5],
			assignees: [],
			projectIds: [],
			minPriority: null,
		})
		expect(wrapper.emitted('update:modelValue')?.[0]).toEqual(['done = false && labels in 5'])
	})

	it('honors the Advanced preference for canonical filters', () => {
		settings.mode = 'advanced'
		const wrapper = mountEditor('')
		expect(wrapper.find('[role="tab"][aria-selected="true"]').text()).toBe('filters.modes.advanced')
	})

	it('keeps unsupported filters intact and disables Simple with a notice', async () => {
		const wrapper = mountEditor('priority > 3')
		expect(wrapper.find('[aria-disabled="true"]').exists()).toBe(true)
		expect(wrapper.text()).toContain('filters.modes.tooAdvanced')
		await wrapper.find('[aria-disabled="true"]').trigger('click')
		expect(wrapper.emitted('update:modelValue')).toBeUndefined()
		const clearing = wrapper.find('[data-test="clear-simple"]').trigger('click')
		await wrapper.setProps({modelValue: ''})
		await clearing
		await flushPromises()
		expect(wrapper.emitted('update:modelValue')?.[0]).toEqual([''])
		expect(wrapper.find('[role="tab"][aria-selected="true"]').text()).toBe('filters.modes.simple')
	})
})
