import {describe, it, expect, vi} from 'vitest'
import {nextTick} from 'vue'

import {fetchAvatarBlobUrl, getDisplayName, invalidateAvatarCache} from './user'
import type {IUser} from '@/modelTypes/IUser'

const {getBlobUrl} = vi.hoisted(() => ({getBlobUrl: vi.fn()}))

vi.mock('@/services/avatar', () => ({
	default: class {
		getBlobUrl = getBlobUrl
	},
}))

function makeUser(overrides: Partial<IUser> = {}): IUser {
	return {
		id: 1,
		email: 'test@example.com',
		username: 'testuser',
		name: '',
		exp: 0,
		type: 1,
		created: new Date(),
		updated: new Date(),
		settings: {} as IUser['settings'],
		isLocalUser: true,
		pendingEmail: '',
		deletionScheduledAt: null,
		...overrides,
	}
}

describe('getDisplayName', () => {
	it('should return the name when set', () => {
		const user = makeUser({name: 'Jane Doe'})
		expect(getDisplayName(user)).toBe('Jane Doe')
	})

	it('should fall back to username when name is empty', () => {
		const user = makeUser({name: '', username: 'janedoe'})
		expect(getDisplayName(user)).toBe('janedoe')
	})
})

describe('fetchAvatarBlobUrl', () => {
	it('should resolve to undefined for a user without a username', async () => {
		await expect(fetchAvatarBlobUrl({} as IUser)).resolves.toBeUndefined()
	})
})

describe('invalidateAvatarCache', () => {
	it('revokes the dropped blob urls, but only once the version bump rendered', async () => {
		const revoke = vi.spyOn(window.URL, 'revokeObjectURL').mockImplementation(() => {})

		getBlobUrl.mockResolvedValueOnce('blob:stale-40')
		await fetchAvatarBlobUrl({username: 'stale'}, 40)
		getBlobUrl.mockResolvedValueOnce('blob:stale-20')
		await fetchAvatarBlobUrl({username: 'stale'}, 20)
		getBlobUrl.mockResolvedValueOnce('blob:kept-40')
		await fetchAvatarBlobUrl({username: 'kept'}, 40)

		invalidateAvatarCache({username: 'stale'})
		// A live <img> still holds the url until the version bump re-rendered.
		expect(revoke).not.toHaveBeenCalled()

		await nextTick()

		expect(revoke).toHaveBeenCalledTimes(2)
		expect(revoke).toHaveBeenCalledWith('blob:stale-40')
		expect(revoke).toHaveBeenCalledWith('blob:stale-20')

		revoke.mockRestore()
	})
})
