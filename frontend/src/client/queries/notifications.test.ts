import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	notificationKeys,
	notificationsQuery,
	markNotificationReadMutationOptions,
	markAllNotificationsReadMutationOptions,
	clearNotificationsMutationOptions,
} from './notifications'
const sdk = vi.hoisted(() => ({
	notificationsList: vi.fn(),
	notificationsMarkRead: vi.fn(),
	notificationsMarkAllRead: vi.fn(),
	notificationsDeleteAll: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('loads every page and patches only the read notification in an existing list', async () => {
	const client = new QueryClient()
	sdk.notificationsList.mockImplementation(({query}) => Promise.resolve({
		data: {
			items: [{id: query.page}],
			total_pages: 2,
		},
	}))
	await client.fetchQuery(notificationsQuery())
	sdk.notificationsMarkRead.mockResolvedValue({data: {id: 1, read_at: '2026-09-21T12:00:00Z'}})
	await client.getMutationCache().build(client, markNotificationReadMutationOptions()).execute(1)
	expect(client.getQueryData(notificationKeys.all)).toEqual([{id: 1, read_at: '2026-09-21T12:00:00Z'}, {id: 2}])
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
	expect(sdk.notificationsMarkRead).toHaveBeenCalledWith({path: {notificationid: 1}, body: {read: true}})
})
it('marks everything read without touching the cached list', async () => {
	const client = new QueryClient()
	sdk.notificationsMarkAllRead.mockResolvedValue({data: {}})
	client.setQueryData(notificationKeys.all, [{id: 1}])
	await client.getMutationCache().build(client, markAllNotificationsReadMutationOptions()).execute(undefined)
	expect(sdk.notificationsMarkAllRead).toHaveBeenCalledTimes(1)
	expect(client.getQueryData(notificationKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
})
it('clears the existing list without creating an absent cache', async () => {
	const client = new QueryClient()
	sdk.notificationsDeleteAll.mockResolvedValue({})
	client.setQueryData(notificationKeys.all, [{id: 1}])
	await client.getMutationCache().build(client, clearNotificationsMutationOptions()).execute(undefined)
	expect(client.getQueryData(notificationKeys.all)).toEqual([])
	client.removeQueries({queryKey: notificationKeys.all})
	await client.getMutationCache().build(client, clearNotificationsMutationOptions()).execute(undefined)
	expect(client.getQueryCache().getAll()).toHaveLength(0)
})
