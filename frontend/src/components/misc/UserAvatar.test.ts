import {describe, it, expect, afterEach, beforeEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import UserAvatar from './UserAvatar.vue'

const sdk = vi.hoisted(() => ({avatarGet: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
const wrappers: VueWrapper[] = []
let queryClient: QueryClient
function mountAvatar(props: InstanceType<typeof UserAvatar>['$props']) {
	const wrapper = mount(UserAvatar, {
		props,
		global: {plugins: [[VueQueryPlugin, {queryClient}]]},
	})
	wrappers.push(wrapper)
	return wrapper
}

function taggedBlob(testUrl: string) {
	return Object.assign(new Blob([testUrl]), {testUrl})
}

beforeEach(() => {
	queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
	sdk.avatarGet.mockReset().mockResolvedValue({data: taggedBlob('blob:avatar')})
	vi.spyOn(URL, 'createObjectURL').mockImplementation(blob => (blob as Blob & {testUrl?: string}).testUrl ?? 'blob:untagged')
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
		expect(sdk.avatarGet).toHaveBeenCalledWith(expect.objectContaining({
			path: {username: 'sam'},
			query: {size: 40},
			parseAs: 'blob',
		}))
		const image = first.find('img')
		expect(image.attributes('src')).toBe('blob:avatar')
		expect(image.attributes('width')).toBe('40')
		expect(image.attributes('height')).toBe('40')
		expect(image.attributes('alt')).toBe('')
		expect(URL.createObjectURL).toHaveBeenCalledTimes(2)
		first.unmount()
		expect(URL.revokeObjectURL).toHaveBeenCalledTimes(1)
		expect(URL.revokeObjectURL).toHaveBeenCalledWith('blob:avatar')
	})
	it('uses the alt prop when one is passed', async () => {
		const wrapper = mountAvatar({user: {username: 'sam'}, alt: "sam's profile image"})
		await flushPromises()
		expect(wrapper.find('img').attributes('alt')).toBe("sam's profile image")
	})
	it('does not refetch when rerendered with an equal user object', async () => {
		const wrapper = mountAvatar({user: {username: 'sam'}, size: 40})
		await flushPromises()
		await wrapper.setProps({user: {username: 'sam'}})
		await flushPromises()
		expect(sdk.avatarGet).toHaveBeenCalledTimes(1)
		expect(wrapper.find('img').attributes('src')).toBe('blob:avatar')
	})
	it('does not display a late response for a previous user', async () => {
		let resolveFirst!: (value: {data: Blob}) => void
		sdk.avatarGet.mockReturnValueOnce(new Promise(resolve => {resolveFirst = resolve}))
		sdk.avatarGet.mockResolvedValue({data: taggedBlob('blob:new')})
		const wrapper = mountAvatar({user: {username: 'old'}})
		await wrapper.setProps({user: {username: 'new'}})
		await flushPromises()
		expect(wrapper.find('img').attributes('src')).toBe('blob:new')
		resolveFirst({data: taggedBlob('blob:old')})
		await flushPromises()
		expect(wrapper.find('img').attributes('src')).toBe('blob:new')
		expect(URL.createObjectURL).toHaveBeenCalledTimes(1)
	})
	it('renders a placeholder without a user or when loading fails', async () => {
		const wrapper = mountAvatar({user: null, size: 40})
		await flushPromises()
		expect(sdk.avatarGet).not.toHaveBeenCalled()
		expect(wrapper.find('img').exists()).toBe(false)
		expect(wrapper.attributes('style')).toContain('--user-avatar-size: 40px')
		sdk.avatarGet.mockRejectedValue(new Error('missing'))
		await wrapper.setProps({user: {username: 'missing'}})
		await flushPromises()
		expect(wrapper.find('img').exists()).toBe(false)
	})
})
