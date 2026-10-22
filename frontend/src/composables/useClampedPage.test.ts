import {
	effectScope,
	nextTick,
	ref,
} from 'vue'
import {describe, expect, it} from 'vitest'

import {useClampedPage} from './useClampedPage'

function setup(initialPage: number) {
	const page = ref(initialPage)
	const data = ref<{total_pages: number}>()
	const isPlaceholderData = ref(false)
	effectScope().run(() => useClampedPage(page, {
		data,
		isPlaceholderData,
	}))
	return {
		page,
		data,
		isPlaceholderData,
	}
}

describe('useClampedPage', () => {
	it('ignores placeholder data and clamps once the real page arrives', async () => {
		const {page, data, isPlaceholderData} = setup(5)
		data.value = {total_pages: 2}
		isPlaceholderData.value = true
		await nextTick()
		expect(page.value).toBe(5)

		isPlaceholderData.value = false
		await nextTick()
		expect(page.value).toBe(2)
	})

	it('steps back to page 1 when the list is empty', async () => {
		const {page, data} = setup(3)
		data.value = {total_pages: 0}
		await nextTick()
		expect(page.value).toBe(1)
	})

	it('keeps an in-range page', async () => {
		const {page, data} = setup(2)
		data.value = {total_pages: 4}
		await nextTick()
		expect(page.value).toBe(2)
	})
})
