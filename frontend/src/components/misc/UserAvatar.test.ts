import {describe, it, expect, afterEach, beforeEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {queryClient} from '@/client/queryClient'
import UserAvatar from './UserAvatar.vue'

const sdk = vi.hoisted(() => ({avatarGet: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
const wrappers: VueWrapper[] = []
function mountAvatar(props: InstanceType<typeof UserAvatar>['$props']) {
	const wrapper = mount(UserAvatar, {props})
	wrappers.push(wrapper)
	return wrapper
}

beforeEach(() => {
	queryClient.clear()
	queryClient.setDefaultOptions({queries: {retry: false, staleTime: Infinity}})
	sdk.avatarGet.mockReset().mockResolvedValue({data: new Blob(['avatar'])})
	vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:avatar')
	vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})
})
afterEach(() => {
	wrappers.splice(0).forEach(wrapper => wrapper.unmount())
	vi.restoreAllMocks()
})

describe('UserAvatar', () => {
	it('shares blob requests and releases component URLs on unmount', async () => {
		const first = mountAvatar({user: {username: 'sam'}, size: 40})
		mountAvatar({user: {username: 'sam'}, size: 40})
		await flushPromises()
		expect(sdk.avatarGet).toHaveBeenCalledTimes(1)
		expect(sdk.avatarGet).toHaveBeenCalledWith(expect.objectContaining({path: {username: 'sam'}, query: {size: 40}, parseAs: 'blob'}))
		expect(first.find('img').attributes('src')).toBe('blob:avatar')
		first.unmount()
		expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:avatar')
	})
	it('does not display a late response for a previous user', async () => {
		let resolveFirst!: (value: {data: Blob}) => void
		sdk.avatarGet.mockReturnValueOnce(new Promise(resolve => {resolveFirst = resolve}))
		const wrapper = mountAvatar({user: {username: 'old'}})
		await wrapper.setProps({user: {username: 'new'}})
		await flushPromises()
		expect(wrapper.find('img').exists()).toBe(true)
		resolveFirst({data: new Blob(['old'])})
		await flushPromises()
		expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
	})
	it('renders a placeholder without a user or when loading fails', async () => {
		const wrapper = mountAvatar({user: null, size: 40})
		await flushPromises()
		expect(sdk.avatarGet).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
		sdk.avatarGet.mockRejectedValue(new Error('missing'))
		await wrapper.setProps({user: {username: 'missing'}})
		await flushPromises()
		expect(wrapper.find('img').exists()).toBe(false)
	})
})
