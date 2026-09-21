import {
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {shallowMount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import {normalizeTimeEntry} from '@/client/queries/timeEntries'
import {useAuthStore} from '@/stores/auth'
import {AUTH_TYPES, type AuthType} from '@/constants/auth'
import type {SessionUser} from '@/stores/auth'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(async () => ({data: {
		items: [],
		total_pages: 1,
	}})),
	tasksRead: vi.fn(async () => ({data: {}})),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

import TimeEntryList from './TimeEntryList.vue'
import BaseButton from '@/components/base/BaseButton.vue'

function mountList(
	entries: Parameters<typeof normalizeTimeEntry>[0][],
	authType: AuthType = AUTH_TYPES.USER,
	paged = false,
) {
	const pinia = createPinia()
	setActivePinia(pinia)
	const authStore = useAuthStore()
	authStore.setAuthenticated(true)
	authStore.setUser({
		id: 1,
		type: authType,
		username: 'user',
	} as SessionUser)

	return shallowMount(TimeEntryList, {
		props: {
			entries: entries.map(normalizeTimeEntry),
			hideLabelColumn: true,
			card: false,
			paged,
		},
		global: {
			plugins: [
				pinia,
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			directives: {
				tooltip: {},
				cy: {},
			},
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('TimeEntryList', () => {
	it('renders no NaN for an entry with an unusable start time', () => {
		const wrapper = mountList([{
			id: 1,
			user_id: 1,
			task_id: 0,
			project_id: 0,
			start_time: '0001-01-01T00:00:00Z',
			end_time: '2026-06-07T10:00:00Z',
			comment: '',
		}])

		expect(wrapper.text()).not.toContain('NaN')
		expect(wrapper.findAll('tbody td')[1]?.text()).toBe('')
		expect(wrapper.findAll('tbody td')[2]?.text()).toBe('')
	})

	it('shows the duration of a completed entry', () => {
		const wrapper = mountList([{
			id: 2,
			user_id: 1,
			task_id: 0,
			project_id: 0,
			start_time: '2026-06-07T09:00:00Z',
			end_time: '2026-06-07T10:30:00Z',
			comment: '',
		}])

		expect(wrapper.findAll('tbody td')[2]?.text()).toBe('1h 30m')
	})

	const entry = {
		id: 3,
		user_id: 1,
		task_id: 0,
		project_id: 0,
		start_time: '2026-06-07T09:00:00Z',
		end_time: '2026-06-07T10:00:00Z',
		comment: '',
	}

	it('shows the row actions to the owning user', () => {
		expect(mountList([entry]).findAllComponents(BaseButton)).toHaveLength(2)
	})

	it('hides the row actions from a link share whose id matches the entry author', () => {
		expect(mountList([entry], AUTH_TYPES.LINK_SHARE).findAllComponents(BaseButton)).toHaveLength(0)
	})

	it('only claims a page-local sum when the entries are one page of a longer list', () => {
		expect(mountList([entry]).find('tfoot td')?.text()).toBe('timeTracking.list.total')
		expect(mountList([entry], AUTH_TYPES.USER, true).find('tfoot td')?.text()).toBe('timeTracking.list.pageTotal')
	})
})
