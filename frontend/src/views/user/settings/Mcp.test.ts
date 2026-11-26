import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createPinia, setActivePinia} from 'pinia'
import Mcp from './Mcp.vue'
import {VueQueryPlugin, QueryClient} from '@tanstack/vue-query'
import en from '@/i18n/lang/en.json'
import type {ConnectionSettings} from '@/client/generated'
import type {ApiTokenPreset} from '@/helpers/apiToken'

const sdk = vi.hoisted(() => ({
	mcpInfo: vi.fn(),
	tokensList: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))

const getAll = sdk.tokensList
const tokenRows = [
	{
		id: 1,
		title: 'My assistant',
		permissions: {mcp: ['access']},
		expires_at: '2036-01-01T00:00:00Z',
		created: '2026-01-01T00:00:00Z',
	},
	{
		id: 2,
		title: 'Private API token',
		permissions: {tasks: ['read_all']},
		expires_at: '2036-01-01T00:00:00Z',
		created: '2026-01-01T00:00:00Z',
	},
]
let wrapper: VueWrapper
beforeEach(() => {
	setActivePinia(createPinia())
	getAll.mockClear()
	getAll.mockResolvedValue({data: {items: tokenRows, total_pages: 1}})
	sdk.mcpInfo.mockReset()
	sdk.mcpInfo.mockResolvedValue({data: {
		endpoint: 'https://example.com/api/v2/mcp',
		oauth_available: true,
		routes: {},
		presets: {read_only: {}, typed: {tasks: ['update']}, full: {'*': '*'}},
	} satisfies ConnectionSettings})
})
afterEach(() => wrapper?.unmount())

function mountSettings() {
	wrapper = mount(Mcp, {
		global: {
			plugins: [
				createI18n({legacy: false, locale: 'en', messages: {en}}),
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			stubs: {
				Card: {template: '<div><slot /></div>'},
				ApiTokenForm: true,
				McpClientGuide: true,
				ConnectedApps: true,
				Modal: true,
				RouterLink: true,
				Icon: true,
			},
		},
	})
}

function guides() {
	return wrapper.findAllComponents({name: 'McpClientGuide'}).map(guide => guide.props('access'))
}

describe('MCP settings', () => {
	it('offers the oauth guide before the token path when oauth is available', async () => {
		mountSettings()
		await flushPromises()
		expect(guides()).toEqual([{kind: 'oauth'}])
		const text = wrapper.text()
		expect(text.indexOf('Connect with OAuth')).toBeLessThan(text.indexOf('MCP tokens'))
		expect(text).toContain('For clients and automation without OAuth, create an MCP token instead.')
	})

	it('only offers tokens when oauth is unavailable', async () => {
		sdk.mcpInfo.mockResolvedValueOnce({data: {
			endpoint: 'https://example.com/api/v2/mcp',
			oauth_available: false,
		} satisfies ConnectionSettings})
		mountSettings()
		await flushPromises()
		expect(guides()).toEqual([])
		expect(wrapper.text()).not.toContain('OAuth')
		await wrapper.findAll('button').find(b => b.text() === 'Create a token')!.trigger('click')
		wrapper.findComponent({name: 'ApiTokenForm'}).vm.$emit('created', {id: 3, token: 'tk_secret'})
		await flushPromises()
		expect(guides()).toEqual([{
			kind: 'token',
			token: 'tk_secret',
			oauthAvailable: false,
		}])
	})

	it('only shows the client guide after creation and forgets the secret on Done', async () => {
		mountSettings()
		await flushPromises()
		expect(sdk.mcpInfo).toHaveBeenCalledOnce()
		expect(getAll).toHaveBeenCalledWith(expect.objectContaining({query: {
			page: 1,
			kind: 'manual',
		}}))
		expect(wrapper.text()).toContain('My assistant')
		expect(wrapper.text()).not.toContain('Private API token')
		expect(wrapper.findComponent({name: 'ConnectedApps'}).exists()).toBe(true)
		await wrapper.findAll('button').find(b => b.text() === 'Create a token')!.trigger('click')
		const form = wrapper.findComponent({name: 'ApiTokenForm'})
		expect(form.attributes('initialscopes')).toBe('tasks:update')
		expect(form.props('presets')).toContainEqual({id: 'fullAccess', groups: {'*': '*'}})
		form.vm.$emit('created', {id: 3, token: 'tk_secret'})
		await flushPromises()
		expect(guides()).toEqual([{
			kind: 'token',
			token: 'tk_secret',
			oauthAvailable: true,
		}])
		await wrapper.findAll('button').find(b => b.text() === 'Done')!.trigger('click')
		await flushPromises()
		expect(guides()).toEqual([{kind: 'oauth'}])
		expect(wrapper.html()).not.toContain('tk_secret')
		expect(getAll).toHaveBeenCalledTimes(1)
	})

	it.each<ConnectionSettings>([
		{},
		{presets: {typed: {tasks: null}, read_only: {tasks: null}, full: {'*': 'unexpected'}}},
	])('handles omitted settings and empty permission groups: %j', async settings => {
		sdk.mcpInfo.mockResolvedValueOnce({data: settings})
		mountSettings()
		await flushPromises()
		expect(wrapper.find<HTMLInputElement>('input#mcp-endpoint').element.value).toBe('')
		await wrapper.findAll('button').find(b => b.text() === 'Create a token')!.trigger('click')
		const form = wrapper.findComponent({name: 'ApiTokenForm'})
		expect(form.props('routes')).toEqual({})
		expect(form.props('initialScopes')).toBe('')
		const presets: ApiTokenPreset[] = form.props('presets')
		for (const {groups} of presets) {
			expect(Object.values(groups).flat()).toEqual([])
		}
	})
})
