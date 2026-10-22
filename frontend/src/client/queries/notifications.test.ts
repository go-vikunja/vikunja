import {it, expect, vi} from 'vitest'
import {toValue, unref} from 'vue'
import {QueryClient} from '@tanstack/vue-query'
import {
	notificationKeys,
	notificationsQuery,
	markNotificationReadMutationOptions,
	markAllNotificationsReadMutationOptions,
	clearNotificationsMutationOptions,
	type NotificationPage,
	type NotificationPages,
} from './notifications'
const sdk = vi.hoisted(() => ({
	notificationsList: vi.fn(),
	notificationsMarkRead: vi.fn(),
	notificationsMarkAllRead: vi.fn(),
	notificationsDeleteAll: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

const UNREAD = '0001-01-01T00:00:00Z'

function page(pageNumber: number, items: NotificationPage['items']): NotificationPage {
	return {
		items,
		page: pageNumber,
		per_page: 20,
		total: 3,
		total_pages: 2,
	}
}

function twoPages(): NotificationPages {
	return {
		pages: [
			page(1, [
				{
					id: 3,
					read_at: UNREAD,
				},
				{
					id: 2,
					read_at: '2026-09-20T08:00:00Z',
				},
			]),
			page(2, [
				{
					id: 1,
					read_at: UNREAD,
				},
			]),
		],
		pageParams: [1, 2],
	}
}

it('loads one page of 20 at a time until total_pages is reached', async () => {
	const client = new QueryClient()
	sdk.notificationsList.mockImplementation(({query}) => Promise.resolve({
		data: {
			items: [{id: query.page}],
			page: query.page,
			per_page: 20,
			total: 21,
			total_pages: 2,
		},
	}))
	await client.fetchInfiniteQuery(notificationsQuery())
	expect(sdk.notificationsList).toHaveBeenCalledTimes(1)
	expect(sdk.notificationsList).toHaveBeenNthCalledWith(1, {
		query: {
			page: 1,
			per_page: 20,
		},
		signal: expect.anything(),
	})
	const getNextPageParam = unref(toValue(notificationsQuery()).getNextPageParam)
	const [first, second] = [page(1, []), page(2, [])]
	expect(getNextPageParam(first, [first], 1, [1])).toBe(2)
	expect(getNextPageParam(second, [first, second], 2, [1, 2])).toBeUndefined()
	expect(getNextPageParam({...first, total_pages: 0}, [first], 1, [1])).toBeUndefined()
	await client.fetchInfiniteQuery({
		...notificationsQuery(),
		pages: 2,
	})
	expect(sdk.notificationsList).toHaveBeenLastCalledWith({
		query: {
			page: 2,
			per_page: 20,
		},
		signal: expect.anything(),
	})
})
it('patches the read notification on whichever cached page holds it', async () => {
	const client = new QueryClient()
	client.setQueryData(notificationKeys.all, twoPages())
	sdk.notificationsMarkRead.mockResolvedValue({
		data: {
			id: 1,
			read_at: '2026-09-21T12:00:00Z',
		},
	})
	await client.getMutationCache().build(client, markNotificationReadMutationOptions()).execute(1)
	expect(sdk.notificationsMarkRead).toHaveBeenCalledWith({
		path: {notificationid: 1},
		body: {read: true},
	})
	const expected = twoPages()
	expected.pages[1].items = [{
		id: 1,
		read_at: '2026-09-21T12:00:00Z',
	}]
	expect(client.getQueryData(notificationKeys.all)).toEqual(expected)
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
})
it('marks every unread notification on every cached page read', async () => {
	vi.useFakeTimers({now: new Date('2026-09-22T10:00:00Z')})
	const client = new QueryClient()
	client.setQueryData(notificationKeys.all, twoPages())
	sdk.notificationsMarkAllRead.mockResolvedValue({data: {}})
	await client.getMutationCache().build(client, markAllNotificationsReadMutationOptions()).execute(undefined)
	vi.useRealTimers()
	expect(sdk.notificationsMarkAllRead).toHaveBeenCalledTimes(1)
	const expected = twoPages()
	expected.pages[0].items[0] = {
		id: 3,
		read_at: '2026-09-22T10:00:00.000Z',
	}
	expected.pages[1].items[0] = {
		id: 1,
		read_at: '2026-09-22T10:00:00.000Z',
	}
	expect(client.getQueryData(notificationKeys.all)).toEqual(expected)
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
})
it('clears to a single empty page without creating an absent cache', async () => {
	const client = new QueryClient()
	sdk.notificationsDeleteAll.mockResolvedValue({})
	client.setQueryData(notificationKeys.all, twoPages())
	await client.getMutationCache().build(client, clearNotificationsMutationOptions()).execute(undefined)
	expect(client.getQueryData(notificationKeys.all)).toEqual({
		pages: [{
			items: [],
			page: 1,
			per_page: 20,
			total: 0,
			total_pages: 0,
		}],
		pageParams: [1],
	})
	expect(client.getQueryState(notificationKeys.all)?.isInvalidated).toBe(true)
	client.removeQueries({queryKey: notificationKeys.all})
	await client.getMutationCache().build(client, clearNotificationsMutationOptions()).execute(undefined)
	expect(client.getQueryCache().getAll()).toHaveLength(0)
})
