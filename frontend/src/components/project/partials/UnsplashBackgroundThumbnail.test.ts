import {defineComponent} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {describe, expect, it, vi} from 'vitest'

vi.mock('@/client/queries/projectBackgrounds', async importOriginal => ({
	...await importOriginal<typeof import('@/client/queries/projectBackgrounds')>(),
	unsplashBackgroundThumbnailQuery: (imageId: string) => ({queryKey: ['thumbnail', imageId]}),
}))
vi.mock('@tanstack/vue-query', async importOriginal => {
	const {ref} = await import('vue')
	return {
		...await importOriginal<typeof import('@tanstack/vue-query')>(),
		useQuery: () => ({data: ref(undefined), error: ref(null)}),
	}
})
vi.mock('@/helpers/getBlobFromBlurHash', () => ({getBlobFromBlurHash: vi.fn().mockResolvedValue(null)}))

import UnsplashBackgroundThumbnail from './UnsplashBackgroundThumbnail.vue'

const BaseButtonStub = defineComponent({props: ['href'], template: '<a :href="href"><slot /></a>'})

function mountThumbnail(info: unknown) {
	return shallowMount(UnsplashBackgroundThumbnail, {
		props: {image: {id: 'image-1', blur_hash: '', info}},
		global: {
			stubs: {BaseButton: BaseButtonStub, CustomTransition: defineComponent({template: '<div><slot /></div>'})},
			mocks: {$t: (key: string, params?: Record<string, string>) => params ? `${key}:${params.author}` : key},
		},
	})
}

describe('UnsplashBackgroundThumbnail', () => {
	it('links to the encoded unsplash author profile', () => {
		const wrapper = mountThumbnail({author: 'a b', author_name: 'A B'})

		const link = wrapper.find('.unsplash-thumbnail__info')

		expect(link.attributes('href')).toBe('https://unsplash.com/@a%20b?utm_source=vikunja&utm_medium=referral')
		expect(link.text()).toBe('A B')
		expect(wrapper.find('.unsplash-thumbnail__button').attributes('aria-label')).toContain('A B')
	})

	it('omits the attribution when the image has no author info', () => {
		expect(mountThumbnail(undefined).find('.unsplash-thumbnail__info').exists()).toBe(false)
	})
})
