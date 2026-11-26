import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import OAuthAuthorize from './OAuthAuthorize.vue'
import en from '@/i18n/lang/en.json'

const {sdk, mockRoute, authStore} = vi.hoisted(() => ({
	sdk: {
		oauthAuthorize: vi.fn(),
		oauthAuthorizePreview: vi.fn(),
	},
	mockRoute: {
		query: {} as Record<string, string>,
	},
	authStore: {
		info: {
			id: 1,
			username: 'alice',
			name: 'Alice',
		},
	},
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('vue-router', () => ({useRoute: () => mockRoute}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => authStore}))

const i18n = createI18n({
	legacy: false,
	locale: 'en',
	messages: {en},
})

const REDIRECT_URI = 'http://127.0.0.1:9999/callback'
const ISSUER = 'https://vikunja.example'
const ISS_PARAM = `iss=${encodeURIComponent(ISSUER)}`
const RESOURCE = 'https://vikunja.example/api/v2/mcp'
const BASE_QUERY = {
	response_type: 'code',
	client_id: 'mcp-client',
	redirect_uri: REDIRECT_URI,
	state: 'st-1',
	code_challenge: 'challenge',
	code_challenge_method: 'S256',
}

const METADATA_CLIENT = {
	id: 'https://app.example.com/oauth/client.json',
	kind: 'metadata_document',
	name: 'Example MCP Client',
	host: 'app.example.com',
}

const originalLocation = window.location

function mountView() {
	return mount(OAuthAuthorize, {
		global: {
			plugins: [i18n],
			stubs: {
				XButton: {
					template: '<button @click="$emit(\'click\')"><slot /></button>',
					emits: ['click'],
				},
			},
		},
	})
}

function preview(scopes: string[], client: Record<string, string> = {
	id: 'mcp-client',
	kind: 'legacy',
}) {
	return {
		data: {
			client,
			redirect_uri: REDIRECT_URI,
			scopes,
			scoped: scopes.length > 0,
			loopback_redirect: true,
			iss: ISSUER,
		},
	}
}

function button(wrapper: ReturnType<typeof mountView>, label: string) {
	const found = wrapper.findAll('button').find(b => b.text() === label)
	if (!found) throw new Error(`no ${label} button`)
	return found
}

beforeEach(() => {
	vi.clearAllMocks()
	mockRoute.query = {...BASE_QUERY}
	Object.defineProperty(window, 'location', {
		configurable: true,
		value: {
			...originalLocation,
			href: 'http://localhost/oauth/authorize',
		},
	})
	sdk.oauthAuthorize.mockResolvedValue({
		data: {
			code: 'the-code',
			redirect_uri: REDIRECT_URI,
			state: 'st-1',
			iss: ISSUER,
		},
	})
})

afterEach(() => {
	vi.restoreAllMocks()
	Object.defineProperty(window, 'location', {
		configurable: true,
		value: originalLocation,
	})
})

describe('OAuthAuthorize', () => {
	it('refuses a scoped request inside a frame', async () => {
		vi.spyOn(window, 'top', 'get').mockReturnValue({} as Window)
		mockRoute.query = {...BASE_QUERY, scope: 'mcp:write', resource: RESOURCE}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read', 'mcp:write']))
		const wrapper = mountView()
		await flushPromises()

		expect(sdk.oauthAuthorizePreview).not.toHaveBeenCalled()
		expect(sdk.oauthAuthorize).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('cannot be shown inside another website')
		expect(wrapper.find('input[value="write"]').exists()).toBe(false)
	})

	it('approves an unscoped request without asking', async () => {
		sdk.oauthAuthorizePreview.mockResolvedValue(preview([]))
		mountView()
		await flushPromises()

		expect(sdk.oauthAuthorizePreview).toHaveBeenCalledOnce()
		expect(sdk.oauthAuthorize).toHaveBeenCalledOnce()
		expect(sdk.oauthAuthorize.mock.calls[0][0].body.scope).toBeUndefined()
		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1&${ISS_PARAM}`)
	})

	it('approves a request without mcp scopes without asking', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'openid offline_access',
		}
		sdk.oauthAuthorizePreview.mockResolvedValue({
			data: {
				client: {id: 'mcp-client'},
				redirect_uri: REDIRECT_URI,
				scopes: [],
				scoped: false,
			},
		})
		mountView()
		await flushPromises()

		expect(sdk.oauthAuthorizePreview).toHaveBeenCalledOnce()
		expect(sdk.oauthAuthorize).toHaveBeenCalledOnce()
		expect(sdk.oauthAuthorize.mock.calls[0][0].body.scope).toBeUndefined()
		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1&${ISS_PARAM}`)
	})

	it('names the scheme of a custom-scheme redirect', async () => {
		mockRoute.query = {...BASE_QUERY, redirect_uri: 'vikunja-flutter://callback', scope: 'mcp:read'}
		const p = preview(['mcp:read'])
		sdk.oauthAuthorizePreview.mockResolvedValue({data: {...p.data, redirect_uri: 'vikunja-flutter://callback'}})
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.get('.host').text()).toBe('vikunja-flutter://callback')
	})

	it('asks for consent when the server scopes a request without scope or resource', async () => {
		const clientID = 'https://app.example.com/oauth/client.json'
		mockRoute.query = {
			...BASE_QUERY,
			client_id: clientID,
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read', 'mcp:write'], {
			id: clientID,
			kind: 'metadata_document',
			name: 'Example MCP Client',
			host: 'app.example.com',
		}))
		const wrapper = mountView()
		await flushPromises()

		expect(sdk.oauthAuthorize).not.toHaveBeenCalled()
		expect(window.location.href).toBe('http://localhost/oauth/authorize')
		expect(wrapper.text()).toContain('Connect Example MCP Client?')
		expect(wrapper.text()).toContain('revoke it earlier in the settings under MCP → Connected apps')

		await button(wrapper, 'Approve').trigger('click')
		await flushPromises()

		expect(sdk.oauthAuthorize).toHaveBeenCalledWith({body: expect.objectContaining({
			client_id: clientID,
			scope: 'mcp:read mcp:write',
		})})
	})

	it('shows the verified domain next to the name of a metadata document client', async () => {
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read'], METADATA_CLIENT))
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.get('.card-header-title').text()).toBe('Connect Example MCP Client?')
		expect(wrapper.text()).toContain('Published by')
		expect(wrapper.get('.publisher').text()).toBe('app.example.com')
		expect(wrapper.text()).not.toContain(METADATA_CLIENT.id)
		expect(wrapper.text()).toContain('127.0.0.1:9999')
		expect(wrapper.text()).toContain('a program running on this device')
	})

	it('warns about loopback redirects only', async () => {
		sdk.oauthAuthorizePreview.mockResolvedValue({
			data: {
				...preview(['mcp:read'], METADATA_CLIENT).data,
				redirect_uri: 'https://app.example.com/callback',
				loopback_redirect: false,
			},
		})
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.get('.publisher').text()).toBe('app.example.com')
		expect(wrapper.text()).not.toContain('a program running on this device')
	})

	it('shows the raw id of a legacy client without claiming a publisher', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:read',
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read']))
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.get('.card-header-title').text()).toBe('Connect mcp-client?')
		expect(wrapper.text()).not.toContain('Published by')
		expect(wrapper.text()).not.toContain('a program running on this device')
	})

	it('grants read-only access when the user downgrades', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:write',
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read', 'mcp:write']))
		const wrapper = mountView()
		await flushPromises()

		await wrapper.get('input[value="read"]').setValue(true)
		await button(wrapper, 'Approve').trigger('click')
		await flushPromises()

		expect(sdk.oauthAuthorize.mock.calls[0][0].body.scope).toBe('mcp:read')
	})

	it('offers no write access when only read was requested', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:read',
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read']))
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.findAll('input[type="radio"]')).toHaveLength(0)
		await button(wrapper, 'Approve').trigger('click')
		await flushPromises()

		expect(sdk.oauthAuthorize.mock.calls[0][0].body.scope).toBe('mcp:read')
	})

	it('omits iss when the API sends none', async () => {
		sdk.oauthAuthorizePreview.mockResolvedValue({data: {...preview([]).data, iss: undefined}})
		sdk.oauthAuthorize.mockResolvedValue({
			data: {
				code: 'the-code',
				redirect_uri: REDIRECT_URI,
				state: 'st-1',
			},
		})
		mountView()
		await flushPromises()

		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1`)
	})

	it('shows a preview error and does not redirect', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			redirect_uri: 'https://evil.example/cb',
			scope: 'mcp:read',
		}
		sdk.oauthAuthorizePreview.mockRejectedValue(new Error('The redirect URI is not allowed for this client.'))
		const wrapper = mountView()
		await flushPromises()

		expect(wrapper.text()).toContain('The redirect URI is not allowed for this client.')
		expect(wrapper.findAll('button')).toHaveLength(0)
		expect(sdk.oauthAuthorize).not.toHaveBeenCalled()
		expect(window.location.href).toBe('http://localhost/oauth/authorize')
	})
})
