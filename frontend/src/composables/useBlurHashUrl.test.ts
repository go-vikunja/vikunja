import {beforeEach, describe, expect, it, vi} from 'vitest'
import {defineComponent, h, nextTick, ref} from 'vue'
import {flushPromises, mount} from '@vue/test-utils'

const getBlobFromBlurHash = vi.hoisted(() => vi.fn())
vi.mock('@/helpers/getBlobFromBlurHash', () => ({getBlobFromBlurHash}))

import {useBlurHashUrl} from './useBlurHashUrl'

function mountBlurHash(hash: ReturnType<typeof ref<string>>) {
	let url: ReturnType<typeof useBlurHashUrl> | undefined
	const wrapper = mount(defineComponent({
		setup() {
			url = useBlurHashUrl(() => hash.value ?? '')
			return () => h('div')
		},
	}))
	return {wrapper, url: url!}
}

describe('useBlurHashUrl', () => {
	beforeEach(() => {
		getBlobFromBlurHash.mockReset()
		let objectUrl = 0
		window.URL.createObjectURL = vi.fn(() => `blob:${++objectUrl}`)
		window.URL.revokeObjectURL = vi.fn()
	})

	it('ignores a decode which resolves after the hash changed', async () => {
		let resolveFirst: (blob: Blob) => void = () => {}
		getBlobFromBlurHash
			.mockReturnValueOnce(new Promise<Blob>(resolve => {
				resolveFirst = resolve
			}))
			.mockResolvedValueOnce(new Blob(['second']))
		const hash = ref('one')
		const {url} = mountBlurHash(hash)

		hash.value = 'two'
		await nextTick()
		resolveFirst(new Blob(['first']))
		await flushPromises()

		expect(url.value).toBe('blob:1')
		expect(window.URL.createObjectURL).toHaveBeenCalledTimes(1)
	})

	it('does not decode an empty hash', async () => {
		const {url} = mountBlurHash(ref(''))
		await flushPromises()

		expect(url.value).toBeUndefined()
		expect(getBlobFromBlurHash).not.toHaveBeenCalled()
	})
})
