import {afterEach, beforeEach, expect, it, vi} from 'vitest'
import {flushPromises, shallowMount, type VueWrapper} from '@vue/test-utils'
import {ref, type Ref} from 'vue'
import {createI18n} from 'vue-i18n'
import {createMemoryHistory, createRouter} from 'vue-router'
import NotificationModel from '@/models/notification'
import Notifications from './Notifications.vue'

const mocks = vi.hoisted(() => ({
	connected: undefined as unknown as Ref<boolean>,
	authenticated: undefined as unknown as Ref<boolean>,
	getAll: vi.fn(),
	subscribe: vi.fn(() => vi.fn()),
}))

vi.mock('@/composables/useWebSocket', () => ({
	useWebSocket: () => mocks,
}))
vi.mock('@/services/notification', () => ({
	default: class {
		getAll = mocks.getAll
	},
}))
vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		info: {id: 1},
		settings: {frontendSettings: {dateDisplay: 'relative', timeFormat: '24h'}},
	}),
}))
vi.mock('@/message', () => ({success: vi.fn()}))

let wrapper: VueWrapper | undefined

beforeEach(() => {
	mocks.connected = ref(true)
	mocks.authenticated = ref(true)
	mocks.getAll.mockReset().mockResolvedValue([])
	vi.useFakeTimers()
})

afterEach(() => {
	wrapper?.unmount()
	vi.useRealTimers()
	vi.restoreAllMocks()
})

async function mountNotifications() {
	wrapper = shallowMount(Notifications, {
		global: {
			plugins: [
				createI18n({legacy: false, locale: 'en', missingWarn: false, fallbackWarn: false}),
				createRouter({history: createMemoryHistory(), routes: [{path: '/', component: {template: '<div />'}}]}),
			],
			stubs: {
				Icon: true,
				BaseButton: {template: '<button><slot /></button>'},
				CustomTransition: {template: '<div><slot /></div>'},
			},
			directives: {tooltip: {}},
		},
	})
	await flushPromises()
}

it('reloads notifications missed while disconnected once the socket authenticates again', async () => {
	await mountNotifications()
	mocks.connected.value = false
	mocks.authenticated.value = false
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(2)

	mocks.getAll.mockResolvedValue([new NotificationModel({
		id: 42,
		name: 'task.reminder',
		notification: {
			task: {id: 1, title: 'Created during outage'},
			project: {id: 1, title: 'Work'},
		} as NotificationModel['notification'],
		created: new Date(),
	})])
	mocks.connected.value = true
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(2)

	mocks.authenticated.value = true
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(3)
	await wrapper!.find('button').trigger('click')
	expect(wrapper!.text()).toContain('Created during outage')
})

it('retries a failed reconnect refresh until it succeeds', async () => {
	vi.spyOn(console, 'warn').mockImplementation(() => {})
	await mountNotifications()
	mocks.authenticated.value = false
	await flushPromises()

	mocks.getAll.mockRejectedValueOnce(new Error('Temporarily unavailable'))
	mocks.authenticated.value = true
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(3)

	await vi.advanceTimersByTimeAsync(10000)
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(4)

	await vi.advanceTimersByTimeAsync(10000)
	expect(mocks.getAll).toHaveBeenCalledTimes(4)
})

it('periodically refreshes healthy sockets after a jittered interval', async () => {
	vi.spyOn(Math, 'random').mockReturnValue(0.5)
	await mountNotifications()
	await vi.advanceTimersByTimeAsync(320000)
	expect(mocks.getAll).toHaveBeenCalledTimes(1)

	await vi.advanceTimersByTimeAsync(10000)
	await flushPromises()
	expect(mocks.getAll).toHaveBeenCalledTimes(2)
	await vi.advanceTimersByTimeAsync(10000)
	expect(mocks.getAll).toHaveBeenCalledTimes(2)
})

it('ignores an older refresh response that arrives after reconnection', async () => {
	await mountNotifications()
	let finishOlderRequest: (value: NotificationModel[]) => void = () => {}
	mocks.getAll.mockImplementationOnce(() => new Promise<NotificationModel[]>(resolve => {
		finishOlderRequest = resolve
	}))
	mocks.authenticated.value = false
	await flushPromises()

	mocks.getAll.mockResolvedValueOnce([new NotificationModel({
		id: 43,
		name: 'task.reminder',
		notification: {
			task: {id: 2, title: 'Latest notification'},
			project: {id: 1, title: 'Work'},
		} as NotificationModel['notification'],
		created: new Date(),
	})])
	mocks.authenticated.value = true
	await flushPromises()
	finishOlderRequest([])
	await flushPromises()
	await wrapper!.find('button').trigger('click')
	expect(wrapper!.text()).toContain('Latest notification')
})
