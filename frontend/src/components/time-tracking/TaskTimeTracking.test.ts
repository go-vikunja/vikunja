import {
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {flushPromises, mount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import type {Directive} from 'vue'

import {useAuthStore} from '@/stores/auth'
import {useConfigStore} from '@/stores/config'
import {AUTH_TYPES} from '@/constants/auth'
import {PRO_FEATURE} from '@/constants/proFeatures'

const RUNNING = {
	id: 5,
	user_id: 1,
	task_id: 42,
	project_id: 0,
	comment: 'Writing tests',
	start_time: new Date(Date.now() - 65_000).toISOString(),
	end_time: null,
}

const sdk = vi.hoisted(() => ({
	running: null as Record<string, unknown> | null,
	timeEntriesList: vi.fn(),
	timeEntriesTimerStop: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

import TaskTimeTracking from './TaskTimeTracking.vue'

const cy: Directive<HTMLElement, string> = {
	mounted(el, {value}) {
		el.setAttribute('data-cy', value)
	},
}

async function mountFor(taskId: number) {
	const pinia = createPinia()
	setActivePinia(pinia)
	const auth = useAuthStore()
	auth.setAuthenticated(true)
	auth.setSession({
		id: 1,
		type: AUTH_TYPES.USER,
		exp: 0,
	})
	useConfigStore().enabled_pro_features = [PRO_FEATURE.TIME_TRACKING]

	const wrapper = mount(TaskTimeTracking, {
		props: {taskId},
		global: {
			plugins: [
				pinia,
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			directives: {
				tooltip: {},
				cy,
			},
			stubs: {
				TimeEntryForm: true,
				TimeEntryList: true,
				PaginationEmit: true,
				XButton: {
					emits: ['click'],
					template: '<button @click="$emit(\'click\')"><slot /></button>',
				},
			},
			mocks: {$t: (key: string) => key},
		},
	})
	await flushPromises()
	return wrapper
}

describe('TaskTimeTracking', () => {
	beforeEach(() => {
		sdk.running = RUNNING
		sdk.timeEntriesList.mockImplementation(async ({query}: {query: {filter: string}}) => ({
			data: query.filter.includes('end_time = null')
				? {items: sdk.running ? [sdk.running] : [], total: sdk.running ? 1 : 0, total_pages: 1}
				: {items: [], total: 0, total_pages: 0},
		}))
		sdk.timeEntriesTimerStop.mockImplementation(async () => {
			const stopped = {...RUNNING, end_time: new Date().toISOString()}
			sdk.running = null
			return {data: stopped}
		})
	})

	it('shows the running timer with its elapsed time when it belongs to this task', async () => {
		const wrapper = await mountFor(42)

		const timer = wrapper.find('[data-cy="taskRunningTimer"]')
		expect(timer.exists()).toBe(true)
		expect(timer.text()).toContain('01:0')
		expect(timer.text()).toContain('Writing tests')
	})

	it('hides the running timer when it belongs to another task', async () => {
		const wrapper = await mountFor(7)

		expect(wrapper.find('[data-cy="taskRunningTimer"]').exists()).toBe(false)
	})

	it('stops the timer from the task', async () => {
		const wrapper = await mountFor(42)

		await wrapper.find('[data-cy="stopTaskTimer"]').trigger('click')
		await flushPromises()

		expect(sdk.timeEntriesTimerStop).toHaveBeenCalledOnce()
		expect(wrapper.find('[data-cy="taskRunningTimer"]').exists()).toBe(false)
	})
})
