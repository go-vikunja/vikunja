import {createPinia} from 'pinia'
import {afterEach, expect, it, vi} from 'vitest'
import {flushPromises, mount, enableAutoUnmount} from '@vue/test-utils'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import LinkPreviewCard from './LinkPreviewCard.vue'

const sdk = vi.hoisted(() => ({linkPreviewRead: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
enableAutoUnmount(afterEach)

function card(href = 'https://example.com/first') {
	const client = new QueryClient({defaultOptions: {queries: {retry: false}}})
	return mount(LinkPreviewCard, {
		props: {href},
		global: {plugins: [createPinia(), [VueQueryPlugin, {queryClient: client}]]},
	})
}

it('discards a late response after the link changes', async () => {
	let resolveFirst!: (value: unknown) => void
	sdk.linkPreviewRead.mockReturnValueOnce(new Promise(resolve => { resolveFirst = resolve }))
		.mockResolvedValueOnce({data: {title: 'Second article', url: 'https://example.com/second'}})
	const wrapper = card()
	await flushPromises()
	await wrapper.setProps({href: 'https://example.com/second'})
	await flushPromises()
	resolveFirst({data: {title: 'Stale article', url: 'https://example.com/first'}})
	await flushPromises()
	expect(wrapper.text()).toContain('Second article')
	expect(wrapper.text()).not.toContain('Stale article')
	expect(wrapper.get('a').attributes('href')).toBe('https://example.com/second')
})

it('renders no card when metadata fetching fails', async () => {
	sdk.linkPreviewRead.mockRejectedValueOnce(new Error('offline'))
	const wrapper = card()
	await flushPromises()
	expect(wrapper.find('a').exists()).toBe(false)
})
