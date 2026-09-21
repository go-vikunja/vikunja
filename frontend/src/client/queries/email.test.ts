import {it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {accountKeys} from './account'
import {updateEmailMutationOptions} from './email'
const sdk = vi.hoisted(() => ({userUpdateEmail: vi.fn(), userShow: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

it('replaces cached account facts with the server-confirmed email state', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, email: 'old@example.test', pending_email: ''})
	sdk.userUpdateEmail.mockResolvedValue({data: {}})
	const account = {id: 1, email: 'old@example.test', pending_email: 'next@example.test'}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, updateEmailMutationOptions()).execute({new_email: 'next@example.test', password: 'secret'})
	expect(sdk.userUpdateEmail).toHaveBeenCalledWith({body: {new_email: 'next@example.test', password: 'secret'}})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(account)
})
