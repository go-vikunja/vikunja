import {afterEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import McpClientGuide, {type McpGuideAccess} from './McpClientGuide.vue'
import en from '@/i18n/lang/en.json'

const copy = vi.fn()
vi.mock('@/composables/useCopyToClipboard', () => ({useCopyToClipboard: () => copy}))
const endpoint = 'https://tasks.example.com/api/v2/mcp'
const token = 'tk_test_secret'
const tokenAccess: McpGuideAccess = {
	kind: 'token',
	token,
	oauthAvailable: true,
}
let wrapper: VueWrapper
function mountGuide(access: McpGuideAccess = tokenAccess) {
	wrapper = mount(McpClientGuide, {
		props: {
			endpoint,
			access,
		},
		global: {
			plugins: [createI18n({legacy: false, locale: 'en', messages: {en}})],
			stubs: {XButton: {template: '<button><slot /></button>'}},
		},
	})
	return wrapper
}
async function copyAll() {
	for (const button of wrapper.findAll('button')) await button.trigger('click')
	return copy.mock.calls.map(([value]) => value).join('\n')
}
afterEach(() => {
	wrapper?.unmount()
	localStorage.clear()
	copy.mockClear()
})

describe('McpClientGuide with a token', () => {
	it('hides instructions until a client is selected', () => {
		mountGuide()
		expect(wrapper.find('ol').exists()).toBe(false)
		expect(wrapper.text()).not.toContain(token)
	})
	it.each(['claudeCode', 'codex', 'claudeDesktop', 'mistral', 'other'])('copies credentials for %s and only persists the client choice', async client => {
		mountGuide()
		await wrapper.get('select').setValue(client)
		await flushPromises()
		const copied = await copyAll()
		expect(copied).toContain(endpoint)
		expect(copied).toContain(token)
		if (client === 'claudeDesktop') expect(copy).toHaveBeenCalledWith(`Bearer ${token}`)
		expect(JSON.stringify(localStorage)).not.toContain(token)
		wrapper.unmount()
		mountGuide()
		expect((wrapper.get('select').element as HTMLSelectElement).value).toBe(client)
	})
	it.each([
		[true, 'Add Vikunja with just the endpoint URL instead.'],
		[false, 'this Vikunja instance has no public URL configured'],
	])('explains that ChatGPT cannot use the token (oauth available: %s)', async (oauthAvailable, explanation) => {
		mountGuide({
			...tokenAccess,
			oauthAvailable,
		})
		await wrapper.get('select').setValue('chatgpt')
		expect(wrapper.text()).toContain(explanation)
		expect(wrapper.text()).not.toContain(token)
		expect(wrapper.find('button').exists()).toBe(false)
	})
})

describe('McpClientGuide with oauth', () => {
	it.each(['claudeCode', 'codex', 'claudeDesktop', 'chatgpt', 'other'])('only needs the endpoint for %s', async client => {
		mountGuide({kind: 'oauth'})
		await wrapper.get('select').setValue(client)
		await flushPromises()
		expect(wrapper.findAll('ol li').length).toBeGreaterThan(1)
		const copied = await copyAll()
		expect(copied).toContain(endpoint)
		expect(copied).not.toMatch(/bearer|token/i)
		expect(wrapper.text()).not.toMatch(/bearer/i)
	})
	it('signs in with the client after adding the endpoint', async () => {
		mountGuide({kind: 'oauth'})
		await wrapper.get('select').setValue('codex')
		await copyAll()
		expect(copy.mock.calls.map(([value]) => value)).toEqual([
			`codex mcp add vikunja --url '${endpoint}'`,
			'codex mcp login vikunja',
		])
	})
	it('sends Mistral Vibe to the token path', async () => {
		mountGuide({kind: 'oauth'})
		await wrapper.get('select').setValue('mistral')
		expect(wrapper.text()).toContain('Use an MCP token instead.')
		expect(wrapper.find('ol').exists()).toBe(false)
		expect(wrapper.find('button').exists()).toBe(false)
	})
})
