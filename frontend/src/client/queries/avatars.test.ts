import {beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient} from '@tanstack/vue-query'
import {success} from '@/message'
import {
	avatarKeys,
	avatarQuery,
	updateAvatarProviderMutationOptions,
	uploadAvatarMutationOptions,
} from './avatars'
const sdk = vi.hoisted(() => ({
	avatarGet: vi.fn(),
	userSetAvatarProvider: vi.fn(),
	userAvatarUpload: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))

describe('avatar mutations', () => {
	beforeEach(() => vi.clearAllMocks())
	it('invalidates every size of the changed avatar without touching another user', async () => {
		const client = new QueryClient()
		for (const size of [20, 50]) client.setQueryData(avatarKeys.image('sam', size), new Blob())
		client.setQueryData(avatarKeys.image('other', 50), new Blob())
		client.setQueryData(avatarKeys.provider, {avatar_provider: 'default'})
		sdk.userSetAvatarProvider.mockResolvedValue({data: {}})
		await client.getMutationCache().build(client, updateAvatarProviderMutationOptions()).execute({
			username: 'sam',
			provider: 'initials',
		})
		expect(sdk.userSetAvatarProvider).toHaveBeenCalledWith({body: {avatar_provider: 'initials'}})
		for (const size of [20, 50]) expect(client.getQueryState(avatarKeys.image('sam', size))?.isInvalidated).toBe(true)
		expect(client.getQueryState(avatarKeys.image('other', 50))?.isInvalidated).toBe(false)
		expect(client.getQueryState(avatarKeys.provider)?.isInvalidated).toBe(true)
		expect(success).toHaveBeenCalledTimes(1)
	})
	it('uploads the blob as a named png and invalidates the avatar and the provider', async () => {
		const client = new QueryClient()
		client.setQueryData(avatarKeys.image('sam', 50), new Blob())
		client.setQueryData(avatarKeys.image('other', 50), new Blob())
		client.setQueryData(avatarKeys.provider, {avatar_provider: 'upload'})
		sdk.userAvatarUpload.mockResolvedValue({data: {}})
		await client.getMutationCache().build(client, uploadAvatarMutationOptions()).execute({
			username: 'sam',
			blob: new Blob(['bytes'], {type: 'image/png'}),
		})
		const {avatar} = sdk.userAvatarUpload.mock.calls[0][0].body
		expect(avatar.name).toBe('avatar.png')
		expect(avatar.type).toBe('image/png')
		expect(client.getQueryState(avatarKeys.image('sam', 50))?.isInvalidated).toBe(true)
		expect(client.getQueryState(avatarKeys.image('other', 50))?.isInvalidated).toBe(false)
		expect(client.getQueryState(avatarKeys.provider)?.isInvalidated).toBe(true)
		expect(success).toHaveBeenCalledTimes(1)
	})
})

describe('avatarQuery', () => {
	beforeEach(() => {
		vi.clearAllMocks()
		vi.unstubAllGlobals()
	})
	it('rejects a response that is not a blob', async () => {
		const client = new QueryClient()
		sdk.avatarGet.mockResolvedValue({data: {message: 'not an image'}})
		await expect(client.fetchQuery(avatarQuery('sam', 50))).rejects.toThrow('Avatar response was not a blob')
	})
	it('caches a non-svg response as bytes', async () => {
		const client = new QueryClient()
		const png = new Blob(['bytes'], {type: 'image/png'})
		sdk.avatarGet.mockResolvedValue({data: png})
		await expect(client.fetchQuery(avatarQuery('sam', 50))).resolves.toBe(png)
	})
	it('downgrades an svg response to an inert data url', async () => {
		const client = new QueryClient()
		sdk.avatarGet.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml'})})
		await expect(client.fetchQuery(avatarQuery('sam', 50)))
			.resolves.toBe('data:image/svg+xml;base64,PHN2ZyAvPg==')
	})
	it('downgrades an svg response with a charset parameter', async () => {
		const client = new QueryClient()
		sdk.avatarGet.mockResolvedValue({data: new Blob(['<svg />'], {type: 'image/svg+xml; charset=utf-8'})})
		await expect(client.fetchQuery(avatarQuery('sam', 50))).resolves.toMatch(/^data:image\/svg\+xml/)
	})
	it('falls back to bytes for svg when FileReader is missing', async () => {
		const client = new QueryClient()
		const svg = new Blob(['<svg />'], {type: 'image/svg+xml'})
		sdk.avatarGet.mockResolvedValue({data: svg})
		vi.stubGlobal('FileReader', undefined)
		await expect(client.fetchQuery(avatarQuery('sam', 50))).resolves.toBe(svg)
	})
})
