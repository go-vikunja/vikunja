import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {
	identityKeys,
	identitiesQuery,
	linkIdentityMutationOptions,
	unlinkIdentityMutationOptions,
} from './identities'

const sdk = vi.hoisted(() => ({
	userIdentitiesList: vi.fn(),
	userIdentitiesLink: vi.fn(),
	userIdentitiesUnlink: vi.fn(),
}))
const message = vi.hoisted(() => ({success: vi.fn(), error: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => message)

const dex = {
	id: 1,
	issuer: 'http://dex:5556',
	provider_key: 'dex',
	provider_name: 'Dex',
}
const removed = {
	id: 2,
	issuer: 'https://gone.example.com',
	provider_key: '',
	provider_name: '',
}

beforeEach(() => {
	vi.clearAllMocks()
})

it('unwraps the identity list', async () => {
	const client = new QueryClient()
	sdk.userIdentitiesList.mockResolvedValue({data: {items: [dex], total: 1, page: 1, per_page: 50, total_pages: 1}})
	expect(await client.fetchQuery(identitiesQuery())).toEqual([dex])
})

it('treats a missing identity list as empty', async () => {
	const client = new QueryClient()
	sdk.userIdentitiesList.mockResolvedValue({data: {items: null}})
	expect(await client.fetchQuery(identitiesQuery())).toEqual([])
})

it('does not toast a failed link', async () => {
	const client = new QueryClient()
	sdk.userIdentitiesLink.mockRejectedValue({status: 403, code: 1011})

	await expect(client.getMutationCache().build(client, linkIdentityMutationOptions()).execute({
		provider: 'dex',
		code: 'code',
		redirect_url: 'url',
		password: 'wrong',
	})).rejects.toEqual({status: 403, code: 1011})

	expect(message.error).not.toHaveBeenCalled()
})

it('removes an unlinked identity and sends the reauth fields as the body', async () => {
	const client = new QueryClient()
	client.setQueryData(identityKeys.list, [dex, removed])
	sdk.userIdentitiesUnlink.mockResolvedValue({})

	await client.getMutationCache().build(client, unlinkIdentityMutationOptions()).execute({
		id: 2,
		password: 'secret',
		totp_passcode: '123456',
	})

	expect(sdk.userIdentitiesUnlink).toHaveBeenCalledWith({
		path: {identity: 2},
		body: {
			password: 'secret',
			totp_passcode: '123456',
		},
	})
	expect(client.getQueryData(identityKeys.list)).toEqual([dex])
	expect(client.getQueryState(identityKeys.list)?.isInvalidated).toBe(true)
})
