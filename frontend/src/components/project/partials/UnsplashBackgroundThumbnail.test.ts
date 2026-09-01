import {defineComponent} from 'vue'
import {shallowMount} from '@vue/test-utils'
import {describe, expect, it, vi} from 'vitest'

const unsplashAuthor = vi.hoisted(() => vi.fn())

vi.mock('@/client/queries/projectBackgrounds', () => ({
	unsplashAuthor,
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

function mountThumbnail() {
	return shallowMount(UnsplashBackgroundThumbnail, {
		props: {image: {id: 'image-1', blur_hash: ''}},
		global: {
			stubs: {BaseButton: BaseButtonStub, CustomTransition: defineComponent({template: '<div><slot /></div>'})},
			mocks: {$t: (key: string) => key},
		},
	})
}

describe('UnsplashBackgroundThumbnail', () => {
	it('links to the encoded unsplash author profile', () => {
		unsplashAuthor.mockReturnValue({author: 'a b', author_name: 'A B'})

		const link = mountThumbnail().find('.unsplash-thumbnail__info')

		expect(link.attributes('href')).toBe('https://unsplash.com/@a%20b?utm_source=vikunja&utm_medium=referral')
		expect(link.text()).toBe('A B')
	})

	it('omits the attribution when the image has no author info', () => {
		unsplashAuthor.mockReturnValue(null)

		expect(mountThumbnail().find('.unsplash-thumbnail__info').exists()).toBe(false)
	})
})
