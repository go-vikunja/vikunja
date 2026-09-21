import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {avatarKeys, updateAvatarProviderMutationOptions} from './avatars'
const sdk = vi.hoisted(() => ({userSetAvatarProvider: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))

describe('avatar provider mutations', () => {
	beforeEach(() => vi.clearAllMocks())
	it('invalidates every size of the changed avatar without touching another user', async () => {
		const client = new QueryClient()
		for (const size of [20, 50]) client.setQueryData(avatarKeys.image('sam', size), new Blob())
		client.setQueryData(avatarKeys.image('other', 50), new Blob())
		sdk.userSetAvatarProvider.mockResolvedValue({data: {}})
		await client.getMutationCache().build(client, updateAvatarProviderMutationOptions()).execute({username: 'sam', provider: 'initials'})
		expect(sdk.userSetAvatarProvider).toHaveBeenCalledWith({body: {avatar_provider: 'initials'}})
		for (const size of [20, 50]) expect(client.getQueryState(avatarKeys.image('sam', size))?.isInvalidated).toBe(true)
		expect(client.getQueryState(avatarKeys.image('other', 50))?.isInvalidated).toBe(false)
	})
})
