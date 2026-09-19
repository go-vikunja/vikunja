import {
	afterEach,
	beforeEach,
	expect,
	it,
	vi,
} from 'vitest'
import {
	defineComponent,
	ref,
	type Ref,
} from 'vue'
import {
	flushPromises,
	mount,
} from '@vue/test-utils'
import {
	QueryClient,
	VueQueryPlugin,
} from '@tanstack/vue-query'
import {useServerCacheEvents} from './useServerCacheEvents'
import {
	normalizeTimeEntry,
	timeEntryKeys,
} from '@/client/queries/timeEntries'
import {commentKeys} from '@/client/queries/comments'

const USER_ID = 7

const ws = vi.hoisted(() => ({
	authenticated: undefined as unknown as Ref<boolean>,
	mayHaveMissedEvents: undefined as unknown as Ref<boolean>,
	subscribedAt: undefined as unknown as Ref<number>,
	handlers: new Map<string, (message: {data: unknown}) => void>(),
	unsubscribed: [] as string[],
}))

vi.mock('./useWebSocket', () => ({
	useWebSocket: () => ({
		authenticated: ws.authenticated,
		mayHaveMissedEvents: ws.mayHaveMissedEvents,
		subscribedAt: ws.subscribedAt,
		subscribe: (event: string, callback: (message: {data: unknown}) => void) => {
			ws.handlers.set(event, callback)
			return () => ws.unsubscribed.push(event)
		},
	}),
}))
vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({info: {id: USER_ID}}),
}))
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

let client: QueryClient

beforeEach(() => {
	ws.authenticated = ref(false)
	ws.mayHaveMissedEvents = ref(false)
	ws.subscribedAt = ref(0)
	ws.handlers.clear()
	ws.unsubscribed = []
	client = new QueryClient()
})

afterEach(() => {
	client.clear()
})

function mountEvents() {
	return mount(defineComponent({
		setup() {
			useServerCacheEvents()
			return () => null
		},
	}), {global: {plugins: [[VueQueryPlugin, {queryClient: client}]]}})
}

const RUNNING_TIMER = {
	id: 4,
	user_id: USER_ID,
	task_id: 3,
	start_time: '2026-09-19T08:00:00Z',
	end_time: null,
}

it('patches the active timer from an event for the current user', async () => {
	client.setQueryData(timeEntryKeys.active(USER_ID), null)
	mountEvents()
	ws.handlers.get('timer.created')!({data: RUNNING_TIMER})
	await flushPromises()
	expect(client.getQueryData(timeEntryKeys.active(USER_ID))).toEqual(normalizeTimeEntry(RUNNING_TIMER))
})

it('ignores an event for another user', async () => {
	client.setQueryData(timeEntryKeys.active(USER_ID), null)
	mountEvents()
	ws.handlers.get('timer.created')!({data: {
		...RUNNING_TIMER,
		user_id: USER_ID + 1,
	}})
	await flushPromises()
	expect(client.getQueryData(timeEntryKeys.active(USER_ID))).toBeNull()
	expect(client.getQueryData(timeEntryKeys.active(USER_ID + 1))).toBeUndefined()
	expect(client.getMutationCache().getAll()).toHaveLength(0)
})

it('sweeps only on a re-authentication after a drop', async () => {
	const commentKey = commentKeys.page(1, 'asc', 1, 50)
	const subscribedAt = Date.now()
	client.setQueryData(commentKey, {items: []}, {updatedAt: subscribedAt + 1})
	mountEvents()

	ws.subscribedAt.value = subscribedAt
	ws.authenticated.value = true
	await flushPromises()
	expect(client.getQueryState(commentKey)?.isInvalidated).toBe(false)

	ws.authenticated.value = false
	await flushPromises()
	ws.authenticated.value = true
	await flushPromises()
	expect(client.getQueryState(commentKey)?.isInvalidated).toBe(true)
	expect(client.getMutationCache().getAll()).toHaveLength(2)
})

it('reconciles only queries loaded before the first subscription', async () => {
	const subscribedAt = Date.now()
	const beforeKey = commentKeys.page(1, 'asc', 1, 50)
	const afterKey = commentKeys.page(2, 'asc', 1, 50)
	client.setQueryData(beforeKey, {items: []}, {updatedAt: subscribedAt - 1})
	client.setQueryData(afterKey, {items: []}, {updatedAt: subscribedAt + 1})
	mountEvents()

	ws.subscribedAt.value = subscribedAt
	ws.authenticated.value = true
	await flushPromises()

	expect(client.getQueryState(beforeKey)?.isInvalidated).toBe(true)
	expect(client.getQueryState(afterKey)?.isInvalidated).toBe(false)
	expect(client.getMutationCache().getAll()).toHaveLength(1)
})

it('sweeps on a first authentication that followed a failed connection attempt', async () => {
	const commentKey = commentKeys.page(1, 'asc', 1, 50)
	const subscribedAt = Date.now()
	client.setQueryData(commentKey, {items: []}, {updatedAt: subscribedAt + 1})
	mountEvents()

	ws.subscribedAt.value = subscribedAt
	ws.mayHaveMissedEvents.value = true
	ws.authenticated.value = true
	await flushPromises()

	expect(client.getQueryState(commentKey)?.isInvalidated).toBe(true)
	expect(client.getMutationCache().getAll()).toHaveLength(1)
})

it('unsubscribes every event when the scope is disposed', () => {
	const wrapper = mountEvents()
	expect([...ws.handlers.keys()]).toEqual([
		'timer.created',
		'timer.updated',
		'timer.deleted',
		'notification.created',
	])
	wrapper.unmount()
	expect(ws.unsubscribed).toEqual([
		'timer.created',
		'timer.updated',
		'timer.deleted',
		'notification.created',
	])
})
