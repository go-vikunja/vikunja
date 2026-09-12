import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createPinia, setActivePinia} from 'pinia'
import Mcp from './Mcp.vue'
import en from '@/i18n/lang/en.json'

const getAll = vi.fn(async () => [
	{id: 1, title: 'My assistant', permissions: {mcp: ['access']}, expiresAt: new Date('2036-01-01'), created: new Date()},
	{id: 2, title: 'Private API token', permissions: {tasks: ['read_all']}, expiresAt: new Date('2036-01-01'), created: new Date()},
])
vi.mock('@/services/apiToken', () => ({default: class {getAll = getAll}}))
vi.mock('@/services/mcp', () => ({getMcpInfo: async () => ({endpoint: 'https://example.com/api/v2/mcp', routes: {}, presets: {read_only: {}, typed: {tasks: ['update']}, full: {}}})}))
let wrapper: VueWrapper
beforeEach(() => {setActivePinia(createPinia()); getAll.mockClear()})
afterEach(() => wrapper?.unmount())

describe('MCP settings', () => {
	it('only shows the client guide after creation and forgets the secret on Done', async () => {
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
		await flushPromises()
		expect(wrapper.text()).toContain('My assistant')
		expect(wrapper.text()).not.toContain('Private API token')
		expect(wrapper.findComponent({name: 'McpClientGuide'}).exists()).toBe(false)
		await wrapper.findAll('button').find(b => b.text() === 'Create a token')!.trigger('click')
		const form = wrapper.findComponent({name: 'ApiTokenForm'})
		expect(form.attributes('initialscopes')).toBe('tasks:update')
		form.vm.$emit('created', {id: 3, token: 'tk_secret'})
		await flushPromises()
		expect(wrapper.findComponent({name: 'McpClientGuide'}).attributes('token')).toBe('tk_secret')
		await wrapper.findAll('button').find(b => b.text() === 'Done')!.trigger('click')
		await flushPromises()
		expect(wrapper.findComponent({name: 'McpClientGuide'}).exists()).toBe(false)
		expect(wrapper.html()).not.toContain('tk_secret')
		expect(getAll).toHaveBeenCalledTimes(2)
	})
})
