import {describe, it, expect, afterEach, beforeEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'

import UserAvatar from './UserAvatar.vue'
import {avatarCacheVersions, fetchAvatarBlobUrl, invalidateAvatarCache} from '@/models/user'

vi.mock('@/models/user', async (importOriginal) => {
	const original = await importOriginal<typeof import('@/models/user')>()
	return {
		...original,
		fetchAvatarBlobUrl: vi.fn(async () => 'blob:avatar'),
	}
})

const fetchAvatarBlobUrlMock = vi.mocked(fetchAvatarBlobUrl)

const wrappers: VueWrapper[] = []

// A leftover mounted avatar would refetch too once the cache version bumps.
function mountAvatar(props: InstanceType<typeof UserAvatar>['$props']) {
	const wrapper = mount(UserAvatar, {props})
	wrappers.push(wrapper)
	return wrapper
}

beforeEach(() => {
	fetchAvatarBlobUrlMock.mockReset()
	fetchAvatarBlobUrlMock.mockResolvedValue('blob:avatar')
	avatarCacheVersions.clear()
})

afterEach(() => {
	wrappers.splice(0).forEach(wrapper => wrapper.unmount())
})

describe('UserAvatar.vue', () => {
	it('renders no img while the avatar is loading', () => {
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})

		expect(wrapper.find('img').exists()).toBe(false)
		expect(wrapper.attributes('style')).toContain('--user-avatar-size: 40px')
	})

	it('renders the img once the blob url resolved', async () => {
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})
		await flushPromises()

		const image = wrapper.find('img')
		expect(image.attributes('src')).toBe('blob:avatar')
		expect(image.attributes('width')).toBe('40')
		expect(image.attributes('height')).toBe('40')
		expect(image.attributes('alt')).toBe('')
	})

	it('uses the alt prop when one is passed', async () => {
		const wrapper = mountAvatar({user: {username: 'user1'}, alt: "user1's profile image"})
		await flushPromises()

		expect(wrapper.find('img').attributes('alt')).toBe("user1's profile image")
	})

	it('refetches when the user changes', async () => {
		fetchAvatarBlobUrlMock.mockResolvedValueOnce('blob:user1')
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})
		await flushPromises()
		expect(wrapper.find('img').attributes('src')).toBe('blob:user1')

		fetchAvatarBlobUrlMock.mockResolvedValueOnce('blob:user2')
		await wrapper.setProps({user: {username: 'user2'}})
		await flushPromises()

		expect(fetchAvatarBlobUrlMock).toHaveBeenCalledTimes(2)
		expect(fetchAvatarBlobUrlMock).toHaveBeenLastCalledWith({username: 'user2'}, 40)
		expect(wrapper.find('img').attributes('src')).toBe('blob:user2')
	})

	it('ignores a stale fetch resolving after a newer one', async () => {
		let resolveFirst: (url: string) => void = () => {}
		fetchAvatarBlobUrlMock.mockReturnValueOnce(new Promise(resolve => {
			resolveFirst = resolve
		}))

		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})

		fetchAvatarBlobUrlMock.mockResolvedValueOnce('blob:user2')
		await wrapper.setProps({user: {username: 'user2'}})
		await flushPromises()

		resolveFirst('blob:user1')
		await flushPromises()

		expect(wrapper.find('img').attributes('src')).toBe('blob:user2')
	})

	it('renders no img when the fetch rejects', async () => {
		fetchAvatarBlobUrlMock.mockRejectedValue(new Error('nope'))
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})
		await flushPromises()

		expect(wrapper.find('img').exists()).toBe(false)
	})

	it('renders no img without a user', async () => {
		const wrapper = mountAvatar({user: null, size: 40})
		await flushPromises()

		expect(wrapper.find('img').exists()).toBe(false)
		expect(fetchAvatarBlobUrlMock).not.toHaveBeenCalled()
	})

	it('does not refetch when rerendered with an equal user object', async () => {
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})
		await flushPromises()

		await wrapper.setProps({user: {username: 'user1'}})

		expect(fetchAvatarBlobUrlMock).toHaveBeenCalledTimes(1)
		expect(wrapper.find('img').exists()).toBe(true)
		expect(wrapper.find('img').attributes('src')).toBe('blob:avatar')
	})

	it('does not refetch when another user was invalidated', async () => {
		const wrapper = mountAvatar({user: {username: 'userB'}, size: 40})
		await flushPromises()

		invalidateAvatarCache({username: 'userA'})
		await flushPromises()

		expect(fetchAvatarBlobUrlMock).toHaveBeenCalledTimes(1)
		expect(wrapper.find('img').exists()).toBe(true)
		expect(wrapper.find('img').attributes('src')).toBe('blob:avatar')
	})

	it('refetches after the avatar cache was invalidated', async () => {
		const wrapper = mountAvatar({user: {username: 'user1'}, size: 40})
		await flushPromises()

		fetchAvatarBlobUrlMock.mockResolvedValueOnce('blob:new')
		invalidateAvatarCache({username: 'user1'})
		await flushPromises()

		expect(wrapper.find('img').attributes('src')).toBe('blob:new')
	})
})
