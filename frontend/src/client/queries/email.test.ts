import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {success} from '@/message'
import {accountKeys, normalizeUserInfo} from './account'
import {
	cancelEmailUpdateMutationOptions,
	resendEmailConfirmationMutationOptions,
	updateEmailMutationOptions,
} from './email'
const sdk = vi.hoisted(() => ({
	userUpdateEmail: vi.fn(),
	userCancelEmailUpdate: vi.fn(),
	userResendEmailConfirmation: vi.fn(),
	userShow: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

beforeEach(() => {
	vi.clearAllMocks()
	sdk.userShow.mockResolvedValue({data: {id: 1}})
})

it('replaces cached account facts with the server-confirmed email state', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, email: 'old@example.test', pending_email: ''})
	sdk.userUpdateEmail.mockResolvedValue({data: {}})
	const account = {id: 1, email: 'old@example.test', pending_email: 'next@example.test'}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, updateEmailMutationOptions()).execute({
		id: 1,
		type: 1,
		body: {
			new_email: 'next@example.test',
			password: 'secret',
		},
	})
	expect(sdk.userUpdateEmail).toHaveBeenCalledWith({body: {new_email: 'next@example.test', password: 'secret'}})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(normalizeUserInfo(account))
	expect(success).toHaveBeenCalledWith({message: 'We\'ve sent a confirmation link to your new address. Your current address stays active until you confirm.'})
})

it('reports an immediate change when the server confirms without a pending email', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, email: 'old@example.test', pending_email: ''})
	sdk.userUpdateEmail.mockResolvedValue({data: {}})
	const account = {id: 1, email: 'next@example.test', pending_email: ''}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, updateEmailMutationOptions()).execute({
		id: 1,
		type: 1,
		body: {
			new_email: 'next@example.test',
			password: 'secret',
		},
	})
	expect(client.getQueryData(accountKeys.user(1))).toEqual(normalizeUserInfo(account))
	expect(success).toHaveBeenCalledWith({message: 'Your email address was successfully updated.'})
})

it('clears the cached pending email when the change is cancelled', async () => {
	const client = new QueryClient()
	client.setQueryData(accountKeys.user(1), {id: 1, email: 'old@example.test', pending_email: 'next@example.test'})
	sdk.userCancelEmailUpdate.mockResolvedValue({data: {}})
	const account = {id: 1, email: 'old@example.test', pending_email: ''}
	sdk.userShow.mockResolvedValue({data: account})
	await client.getMutationCache().build(client, cancelEmailUpdateMutationOptions()).execute({
		id: 1,
		type: 1,
	})
	expect(sdk.userCancelEmailUpdate).toHaveBeenCalledTimes(1)
	expect(client.getQueryData(accountKeys.user(1))).toEqual(normalizeUserInfo(account))
	expect(success).toHaveBeenCalledWith({message: 'The email change was cancelled.'})
})

it('reports a resent confirmation link', async () => {
	const client = new QueryClient()
	sdk.userResendEmailConfirmation.mockResolvedValue({data: {}})
	await client.getMutationCache().build(client, resendEmailConfirmationMutationOptions()).execute({
		id: 1,
		type: 1,
	})
	expect(sdk.userResendEmailConfirmation).toHaveBeenCalledTimes(1)
	expect(success).toHaveBeenCalledWith({message: 'We\'ve sent you a new confirmation link.'})
})
