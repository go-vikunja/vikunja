import {beforeEach, it, expect, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {linkIdentityMutationOptions} from './identities'

const sdk = vi.hoisted(() => ({
	userIdentitiesLink: vi.fn(),
}))
const message = vi.hoisted(() => ({
	success: vi.fn(),
	error: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => message)

beforeEach(() => {
	vi.clearAllMocks()
})

it('does not toast a failed link', async () => {
	const client = new QueryClient()
	sdk.userIdentitiesLink.mockRejectedValue({
		status: 403,
		code: 1011,
	})

	await expect(client.getMutationCache().build(client, linkIdentityMutationOptions()).execute({
		provider: 'dex',
		code: 'code',
		redirect_url: 'url',
		password: 'wrong',
	})).rejects.toEqual({
		status: 403,
		code: 1011,
	})

	expect(message.error).not.toHaveBeenCalled()
})
