import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient, QueryObserver} from '@tanstack/vue-query'
import {adminKeys, adminUsersQuery, updateAdminUserMutationOptions, deleteAdminUserMutationOptions} from './admin'
import {accountKeys, currentUserQuery} from './account'
const sdk = vi.hoisted(() => ({
	adminUsersList: vi.fn(),
	adminUsersPatchAdmin: vi.fn(),
	adminUsersPatchStatus: vi.fn(),
	adminUsersDelete: vi.fn(),
	userShow: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

beforeEach(() => {
	vi.resetAllMocks()
})

it('uses server pagination and search and invalidates all user pages after a partial update fails', async () => {
	const client = new QueryClient()
	sdk.adminUsersList.mockResolvedValue({data: {items: [{id: 1}], total: 100, total_pages: 2}})
	await client.fetchQuery(adminUsersQuery('name', 2))
	client.setQueryData(adminKeys.usersPage('', 1), {items: [{id: 2}], total: 100, total_pages: 2})
	sdk.adminUsersPatchAdmin.mockResolvedValue({data: {id: 1, is_admin: true}})
	sdk.adminUsersPatchStatus.mockRejectedValue(new Error('status rejected'))
	await expect(client.getMutationCache().build(client, updateAdminUserMutationOptions())
		.execute({id: 1, is_admin: true, status: 2})).rejects.toThrow('status rejected')
	expect(sdk.adminUsersList).toHaveBeenCalledWith(expect.objectContaining({query: {q: 'name', page: 2}}))
	expect(client.getQueryState(adminKeys.usersPage('name', 2))?.isInvalidated).toBe(true)
	expect(client.getQueryState(adminKeys.usersPage('', 1))?.isInvalidated).toBe(true)
})
it('deletion stales users and overview without inventing a page total', async () => {
	const client = new QueryClient()
	const page = {items: [{id: 1}], total: 101, total_pages: 3}
	client.setQueryData(adminKeys.usersPage('', 1), page)
	client.setQueryData(adminKeys.overview, {users: 101})
	sdk.adminUsersDelete.mockResolvedValue({})
	await client.getMutationCache().build(client, deleteAdminUserMutationOptions())
		.execute({id: 1, mode: 'now', username: 'Old'})
	expect(client.getQueryState(adminKeys.usersPage('', 1))?.isInvalidated).toBe(true)
	expect(client.getQueryState(adminKeys.overview)?.isInvalidated).toBe(true)
	expect(client.getQueryData(adminKeys.usersPage('', 1))).toEqual(page)
})
it('demoting the acting admin flips is_admin in the account cache', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, username: 'admin', is_admin: true})
	sdk.adminUsersPatchAdmin.mockResolvedValue({data: {id: 1, username: 'admin', is_admin: false}})
	await client.getMutationCache().build(client, updateAdminUserMutationOptions())
		.execute({id: 1, is_admin: false})
	expect(client.getQueryData(accountKeys.user(1))).toEqual({id: 1, username: 'admin', is_admin: false})
})
it('demoting another user leaves the account cache alone', async () => {
	const client = new QueryClient()
	const account = {id: 1, username: 'admin', is_admin: true}
	client.setQueryData(accountKeys.user(1), account)
	sdk.adminUsersPatchAdmin.mockResolvedValue({data: {id: 2, username: 'other', is_admin: false}})
	await client.getMutationCache().build(client, updateAdminUserMutationOptions())
		.execute({id: 2, is_admin: false})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(account)
})
it('refetches the acting admin account when a self-targeting update fails halfway', async () => {
	const client = new QueryClient()
	sdk.userShow.mockResolvedValue({
		data: {
			id: 1,
			username: 'admin',
			is_admin: true,
		},
	})
	const unsubscribe = new QueryObserver(client, {
		queryKey: accountKeys.user(1),
		queryFn: currentUserQuery(1).queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(client.getQueryData(accountKeys.user(1))).toMatchObject({is_admin: true}))
	sdk.adminUsersPatchAdmin.mockResolvedValue({
		data: {
			id: 1,
			is_admin: false,
		},
	})
	sdk.adminUsersPatchStatus.mockRejectedValue(new Error('status rejected'))
	sdk.userShow.mockResolvedValue({
		data: {
			id: 1,
			username: 'admin',
			is_admin: false,
		},
	})
	await expect(client.getMutationCache().build(client, updateAdminUserMutationOptions()).execute({
		id: 1,
		is_admin: false,
		status: 2,
	})).rejects.toThrow('status rejected')
	expect(sdk.userShow).toHaveBeenCalledTimes(2)
	expect(client.getQueryData(accountKeys.user(1))).toMatchObject({is_admin: false})
	unsubscribe()
})
it('does not refetch the acting admin account when another user is updated', async () => {
	const client = new QueryClient()
	sdk.userShow.mockResolvedValue({
		data: {
			id: 1,
			username: 'admin',
			is_admin: true,
		},
	})
	const unsubscribe = new QueryObserver(client, {
		queryKey: accountKeys.user(1),
		queryFn: currentUserQuery(1).queryFn,
	}).subscribe(() => {})
	await vi.waitFor(() => expect(sdk.userShow).toHaveBeenCalledTimes(1))
	sdk.adminUsersPatchAdmin.mockResolvedValue({
		data: {
			id: 2,
			is_admin: false,
		},
	})
	sdk.adminUsersPatchStatus.mockRejectedValue(new Error('status rejected'))
	await expect(client.getMutationCache().build(client, updateAdminUserMutationOptions()).execute({
		id: 2,
		is_admin: false,
		status: 2,
	})).rejects.toThrow('status rejected')
	expect(sdk.userShow).toHaveBeenCalledTimes(1)
	unsubscribe()
})
