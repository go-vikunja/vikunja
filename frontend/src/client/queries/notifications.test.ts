import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {notificationKeys, notificationsQuery, markNotificationReadMutationOptions, clearNotificationsMutationOptions} from './notifications'
import {parseServerCacheEvent, serverCacheEventMutationOptions} from './serverEvents'
const sdk = vi.hoisted(() => ({notificationsList: vi.fn(), notificationsMarkRead: vi.fn(), notificationsMarkAllRead: vi.fn(), notificationsDeleteAll: vi.fn()}))
vi.mock('@/client/generated', async importOriginal => ({...await importOriginal<object>(), ...sdk}))
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('loads every page and patches only the read notification in an existing list', async () => {
	const client = new QueryClient()
	sdk.notificationsList.mockImplementation(({query}) => Promise.resolve({data: {items: [{id: query.page}], total_pages: 2}}))
	await client.fetchQuery(notificationsQuery())
	sdk.notificationsMarkRead.mockResolvedValue({data: {id: 1, read_at: '2026-09-21T12:00:00Z'}})
	await client.getMutationCache().build(client, markNotificationReadMutationOptions()).execute(1)
	expect(client.getQueryData(notificationKeys.all)).toEqual([{id: 1, read_at: '2026-09-21T12:00:00Z'}, {id: 2}])
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
	expect(sdk.notificationsMarkRead).toHaveBeenCalledWith({path: {notificationid: 1}, body: {read: true}})
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
it.each(['notification', 'comment', 'reconnect', 'subscribe'])('invalidates notification data on %s without replacing it with the event payload', async kind => {
	const client = new QueryClient()
	client.setQueryData(notificationKeys.all, [{id: 1}], {updatedAt: 10})
	const event = kind === 'reconnect' ? {kind: 'reconnect'} as const
		: kind === 'subscribe' ? {kind: 'subscribed', since: 20} as const
			: parseServerCacheEvent('notification.created', {id: 2, name: kind === 'comment' ? 'task.comment' : 'team.member.added', notification: {task: {id: 3}}}, 7)!
	await client.getMutationCache().build(client, serverCacheEventMutationOptions()).execute(event)
	expect(client.getQueryData(notificationKeys.all)).toEqual([{id: 1}])
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
})
