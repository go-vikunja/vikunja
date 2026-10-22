import {afterEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import McpClientGuide from './McpClientGuide.vue'
import en from '@/i18n/lang/en.json'

const copy = vi.fn()
vi.mock('@/composables/useCopyToClipboard', () => ({useCopyToClipboard: () => copy}))
const endpoint = 'https://tasks.example.com/api/v2/mcp'
const token = 'tk_test_secret'
let wrapper: VueWrapper
function mountGuide() {
	wrapper = mount(McpClientGuide, {
		props: {endpoint, token},
		global: {
			plugins: [createI18n({legacy: false, locale: 'en', messages: {en}})],
			stubs: {XButton: {template: '<button><slot /></button>'}},
		},
	})
	return wrapper
}
afterEach(() => {
	wrapper?.unmount()
	localStorage.clear()
	copy.mockClear()
})

describe('McpClientGuide', () => {
	it('hides instructions until a client is selected', () => {
		mountGuide()
		expect(wrapper.find('ol').exists()).toBe(false)
		expect(wrapper.text()).not.toContain(token)
	})
	it.each(['claudeCode', 'codex', 'claudeDesktop', 'mistral', 'other'])('copies credentials for %s and only persists the client choice', async client => {
		mountGuide()
		await wrapper.get('select').setValue(client)
		await flushPromises()
		for (const button of wrapper.findAll('button')) await button.trigger('click')
		const copied = copy.mock.calls.map(([value]) => value).join('\n')
		expect(copied).toContain(endpoint)
		expect(copied).toContain(token)
		if (client === 'claudeDesktop') expect(copy).toHaveBeenCalledWith(`Bearer ${token}`)
		expect(JSON.stringify(localStorage)).not.toContain(token)
		wrapper.unmount()
		mountGuide()
		expect((wrapper.get('select').element as HTMLSelectElement).value).toBe(client)
	})
	it('explains ChatGPT support without offering credentials', async () => {
		mountGuide()
		await wrapper.get('select').setValue('chatgpt')
		expect(wrapper.text()).toContain('OAuth')
		expect(wrapper.text()).not.toContain(token)
		expect(wrapper.find('button').exists()).toBe(false)
	})
})
