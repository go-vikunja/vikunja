import {describe, it, expect} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'

import TipTap from './TipTap.vue'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

async function editButtonShown(isEditEnabled: boolean, bottomActions: {title: string, action: () => void}[]) {
	const wrapper = mount(TipTap, {
		props: {modelValue: '<p>hello</p>', initialMode: 'preview', isEditEnabled, bottomActions},
		global: {plugins: [i18n]},
		attachTo: document.body,
	})
	await flushPromises()
	const shown = wrapper.findAll('button').some(b => b.text() === 'Edit')
	wrapper.unmount()
	return shown
}

describe('TipTap edit button', () => {
	const reply = [{title: 'Reply', action: () => {}}]

	it.each([
		[true, [], true],
		[false, [], false],
		[true, reply, true],
		[false, reply, false],
	])('isEditEnabled=%s with %j bottom actions shows edit: %s', async (isEditEnabled, actions, expected) => {
		expect(await editButtonShown(isEditEnabled, actions)).toBe(expected)
	})
})
