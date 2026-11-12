import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {accountKeys, normalizeUserInfo} from './account'
import {cancelDeletionMutationOptions, confirmDeletionMutationOptions} from './accountDeletion'
const sdk = vi.hoisted(() => ({
	userDeletionCancel: vi.fn(),
	userDeletionConfirm: vi.fn(),
	userShow: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

beforeEach(() => {
	vi.clearAllMocks()
})

it('stores the scheduled deletion date a confirmed token produced', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), normalizeUserInfo({
		id: 1,
		deletion_scheduled_at: '0001-01-01T00:00:00Z',
	}))
	sdk.userDeletionConfirm.mockResolvedValue({})
	const account = {
		id: 1,
		deletion_scheduled_at: '2030-01-01T00:00:00Z',
	}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, confirmDeletionMutationOptions()).execute({
		id: 1,
		type: 1,
		token: 'token',
	})
	expect(sdk.userDeletionConfirm).toHaveBeenCalledWith({body: {token: 'token'}})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(normalizeUserInfo(account))
})

it('replaces the scheduled deletion with refreshed account facts', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), normalizeUserInfo({
		id: 1,
		deletion_scheduled_at: '2030-01-01T00:00:00Z',
	}))
	sdk.userDeletionCancel.mockResolvedValue({})
	const account = {
		id: 1,
		deletion_scheduled_at: '0001-01-01T00:00:00Z',
	}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, cancelDeletionMutationOptions()).execute({
		id: 1,
		type: 1,
		password: 'password',
	})
	expect(sdk.userDeletionCancel).toHaveBeenCalledWith({body: {password: 'password'}})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(normalizeUserInfo(account))
})

it('keeps a confirmed deletion successful when the account re-read fails', async () => {
	const client = new QueryClient()
	sdk.userDeletionConfirm.mockResolvedValue({})
	sdk.userShow.mockRejectedValue(new Error('offline'))
	const mutation = client.getMutationCache().build(client, confirmDeletionMutationOptions())
	await mutation.execute({
		id: 1,
		type: 1,
		token: 'token',
	})
	expect(mutation.state.status).toBe('success')
	expect(sdk.userShow).toHaveBeenCalledTimes(1)
})
