import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createPinia, setActivePinia} from 'pinia'
import Mcp from './Mcp.vue'
import en from '@/i18n/lang/en.json'
import type {ConnectionSettings} from '@/client/generated'
import type {ApiTokenPreset} from '@/modelTypes/IApiTokenSettings'

const sdk = vi.hoisted(() => ({mcpInfo: vi.fn()}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/composables/useTitle', () => ({useTitle: vi.fn()}))

const getAll = vi.fn(async () => [
	{id: 1, title: 'My assistant', permissions: {mcp: ['access']}, expiresAt: new Date('2036-01-01'), created: new Date()},
	{id: 2, title: 'Private API token', permissions: {tasks: ['read_all']}, expiresAt: new Date('2036-01-01'), created: new Date()},
])
vi.mock('@/services/apiToken', () => ({default: class {getAll = getAll}}))
let wrapper: VueWrapper
beforeEach(() => {
	setActivePinia(createPinia())
	getAll.mockClear()
	sdk.mcpInfo.mockReset()
	sdk.mcpInfo.mockResolvedValue({data: {
		endpoint: 'https://example.com/api/v2/mcp',
		routes: {},
		presets: {read_only: {}, typed: {tasks: ['update']}, full: {'*': '*'}},
	} satisfies ConnectionSettings})
})
afterEach(() => wrapper?.unmount())

function mountSettings() {
	wrapper = mount(Mcp, {
		global: {
			plugins: [createI18n({legacy: false, locale: 'en', messages: {en}})],
			stubs: {
				Card: {template: '<div><slot /></div>'},
				ApiTokenForm: true,
				McpClientGuide: true,
				Modal: true,
				RouterLink: true,
				Icon: true,
			},
		},
	})
}

describe('MCP settings', () => {
	it('only shows the client guide after creation and forgets the secret on Done', async () => {
		mountSettings()
		await flushPromises()
		expect(sdk.mcpInfo).toHaveBeenCalledOnce()
		expect(wrapper.text()).toContain('My assistant')
		expect(wrapper.text()).not.toContain('Private API token')
		expect(wrapper.findComponent({name: 'McpClientGuide'}).exists()).toBe(false)
		await wrapper.findAll('button').find(b => b.text() === 'Create a token')!.trigger('click')
		const form = wrapper.findComponent({name: 'ApiTokenForm'})
		expect(form.attributes('initialscopes')).toBe('tasks:update')
		expect(form.props('presets')).toContainEqual({id: 'fullAccess', groups: {'*': '*'}})
		form.vm.$emit('created', {id: 3, token: 'tk_secret'})
		await flushPromises()
		expect(wrapper.findComponent({name: 'McpClientGuide'}).attributes('token')).toBe('tk_secret')
		await wrapper.findAll('button').find(b => b.text() === 'Done')!.trigger('click')
		await flushPromises()
		expect(wrapper.findComponent({name: 'McpClientGuide'}).exists()).toBe(false)
		expect(wrapper.html()).not.toContain('tk_secret')
		expect(getAll).toHaveBeenCalledTimes(2)
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
