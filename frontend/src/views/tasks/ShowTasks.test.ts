import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {reactive} from 'vue'
import {enableAutoUnmount, flushPromises, shallowMount} from '@vue/test-utils'
import {createMemoryHistory, createRouter} from 'vue-router'

const loadTasks = vi.fn(async () => [])
const auth = reactive({
	authenticated: true,
	settings: {frontendSettings: {filterIdUsedOnOverview: undefined, sidebarWidth: 250}},
})

vi.mock('@/stores/auth', () => ({useAuthStore: () => auth}))
vi.mock('@/stores/tasks', () => ({useTaskStore: () => ({loadTasks, isLoading: false})}))
vi.mock('@/stores/projects', () => ({useProjectStore: () => ({projects: {}})}))
vi.mock('@/composables/useLabels', () => ({useLabels: () => ({getLabelById: vi.fn()})}))
vi.mock('@/helpers/setTitle', () => ({setTitle: vi.fn()}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ShowTasks from './ShowTasks.vue'

enableAutoUnmount(afterEach)

async function mountUpcoming() {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/', component: {render: () => null}}],
	})
	await router.push('/')
	return shallowMount(ShowTasks, {
		props: {dateFrom: 'now/d', dateTo: 'now/d+1d'},
		global: {
			plugins: [router],
			mocks: {$t: (key: string) => key},
			stubs: {Card: true, XButton: true, 'i18n-t': true, DatepickerWithRange: {render: () => null}},
			directives: {tooltip: () => {}, cy: () => {}},
		},
	})
}

describe('Upcoming filters', () => {
	beforeEach(() => loadTasks.mockClear())

	it('reloads tasks when overdue tasks are enabled and disabled', async () => {
		const wrapper = await mountUpcoming()
		await flushPromises()
		expect(loadTasks).toHaveBeenCalledTimes(1)
		expect(loadTasks).toHaveBeenLastCalledWith(expect.objectContaining({
			filter: "done = false && due_date < 'now/d+1d' && due_date > 'now/d'",
		}), null)

		await wrapper.setProps({showOverdue: true})
		expect(loadTasks).toHaveBeenCalledTimes(2)
		expect(loadTasks).toHaveBeenLastCalledWith(expect.objectContaining({
			filter: "done = false && due_date < 'now/d+1d'",
		}), null)

		await wrapper.setProps({showOverdue: false})
		expect(loadTasks).toHaveBeenCalledTimes(3)
		expect(loadTasks).toHaveBeenLastCalledWith(expect.objectContaining({
			filter: "done = false && due_date < 'now/d+1d' && due_date > 'now/d'",
		}), null)
	})

	it('reloads tasks when undated tasks are enabled and disabled', async () => {
		const wrapper = await mountUpcoming()
		await flushPromises()

		await wrapper.setProps({showNulls: true})
		expect(loadTasks).toHaveBeenCalledTimes(2)
		expect(loadTasks).toHaveBeenLastCalledWith(expect.objectContaining({filter_include_nulls: true}), null)

		await wrapper.setProps({showNulls: false})
		expect(loadTasks).toHaveBeenCalledTimes(3)
		expect(loadTasks).toHaveBeenLastCalledWith(expect.objectContaining({filter_include_nulls: false}), null)
	})

	it('does not reload tasks when sidebar settings change', async () => {
		await mountUpcoming()
		await flushPromises()

		auth.settings = {frontendSettings: {filterIdUsedOnOverview: undefined, sidebarWidth: 300}}
		await flushPromises()
		expect(loadTasks).toHaveBeenCalledTimes(1)
	})
})
