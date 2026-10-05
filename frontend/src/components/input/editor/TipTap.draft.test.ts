import {describe, it, expect, beforeEach} from 'vitest'
import {defineComponent, h, ref} from 'vue'
import {mount, flushPromises} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'

import TipTap from './TipTap.vue'
import en from '@/i18n/lang/en.json'

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountWithDraft(draft: string) {
	localStorage.setItem('editorDraft-task-comment-1', draft)
	const text = ref('')
	const Parent = defineComponent({
		setup: () => () => h(TipTap, {
			modelValue: text.value,
			'onUpdate:modelValue': (v: string) => { text.value = v },
			storageKey: 'task-comment-1',
		}),
	})
	return mount(Parent, {global: {plugins: [i18n]}, attachTo: document.body})
}

describe('TipTap draft restore', () => {
	beforeEach(() => localStorage.clear())

	it.each([
		'<p>hello</p>',
		'<p>hello  world</p>',
		'<p>hello</p><p></p>',
		'<p>a</p><ul><li><p>b</p></li></ul>',
	])('stays in edit mode for %s', async draft => {
		const wrapper = mountWithDraft(draft)
		await flushPromises()
		await flushPromises()
		expect(wrapper.find('.tiptap__editor-is-edit-enabled').exists()).toBe(true)
	})
})
