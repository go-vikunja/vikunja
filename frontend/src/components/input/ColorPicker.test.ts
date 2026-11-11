import {describe, it, expect} from 'vitest'
import {mount} from '@vue/test-utils'
import {nextTick} from 'vue'
import ColorPicker from './ColorPicker.vue'

function mountPicker(modelValue: string) {
	const emitted: string[] = []
	const wrapper = mount(ColorPicker, {
		props: {
			modelValue,
			'onUpdate:modelValue': (val: string) => emitted.push(val),
		},
		global: {mocks: {$t: (key: string) => key}},
	})
	return {wrapper, emitted}
}

describe('ColorPicker', () => {
	it('emits when the user picks a color', async () => {
		const {wrapper, emitted} = mountPicker('')
		await wrapper.find('input[type="color"]').setValue('#00ff00')
		expect(emitted).toEqual(['#00ff00'])
	})

	it('does not echo external model changes back as updates', async () => {
		const {wrapper, emitted} = mountPicker('ff0000')
		await nextTick()

		// The API returns colors without the leading '#'
		await wrapper.setProps({modelValue: '00ff00'})
		await wrapper.setProps({modelValue: '#0000ff'})
		await nextTick()

		expect(emitted).toEqual([])
	})

	it('emits an empty value on reset', async () => {
		const {wrapper, emitted} = mountPicker('ff0000')
		await wrapper.find('button').trigger('click')
		expect(emitted).toEqual([''])
	})
})
