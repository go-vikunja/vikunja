import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {accountKeys} from './account'
import {cancelDeletionMutationOptions} from './accountDeletion'
const sdk = vi.hoisted(() => ({userDeletionCancel: vi.fn(), userShow: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('replaces the scheduled deletion with refreshed account facts', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, deletion_scheduled_at: '2030-01-01T00:00:00Z'})
	sdk.userDeletionCancel.mockResolvedValue({})
	const account = {id: 1, deletion_scheduled_at: '0001-01-01T00:00:00Z'}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, cancelDeletionMutationOptions()).execute('password')
	expect(sdk.userDeletionCancel).toHaveBeenCalledWith({body: {password: 'password'}})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(account)
})
