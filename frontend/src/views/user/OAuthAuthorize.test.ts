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
const RESOURCE = 'https://vikunja.example/api/v2/mcp'
const BASE_QUERY = {
	response_type: 'code',
	client_id: 'mcp-client',
	redirect_uri: REDIRECT_URI,
	state: 'st-1',
	code_challenge: 'challenge',
	code_challenge_method: 'S256',
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

function preview(scopes: string[]) {
	return {
		data: {
			client: {id: 'mcp-client'},
			redirect_uri: REDIRECT_URI,
			scopes,
			scoped: true,
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
		},
	})
})

afterEach(() => {
	Object.defineProperty(window, 'location', {
		configurable: true,
		value: originalLocation,
	})
})

describe('OAuthAuthorize', () => {
	it('approves an unscoped request without asking', async () => {
		mountView()
		await flushPromises()

		expect(sdk.oauthAuthorizePreview).not.toHaveBeenCalled()
		expect(sdk.oauthAuthorize).toHaveBeenCalledOnce()
		expect(sdk.oauthAuthorize.mock.calls[0][0].body.scope).toBeUndefined()
		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1`)
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
		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1`)
	})

	it('asks for consent before authorizing a scoped request', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:write',
			resource: RESOURCE,
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read', 'mcp:write']))
		const wrapper = mountView()
		await flushPromises()

		expect(sdk.oauthAuthorizePreview).toHaveBeenCalledWith({body: expect.objectContaining({
			client_id: 'mcp-client',
			scope: 'mcp:write',
			resource: RESOURCE,
		})})
		expect(sdk.oauthAuthorize).not.toHaveBeenCalled()
		expect(window.location.href).toBe('http://localhost/oauth/authorize')
		expect(wrapper.text()).toContain('mcp-client')
		expect(wrapper.text()).toContain('127.0.0.1:9999')
		expect(wrapper.text()).toContain('Alice (alice)')
		expect(wrapper.get<HTMLInputElement>('input[value="write"]').element.checked).toBe(true)
	})

	it('sends the requested scope and resource on approve', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:write',
			resource: RESOURCE,
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read', 'mcp:write']))
		const wrapper = mountView()
		await flushPromises()

		await button(wrapper, 'Approve').trigger('click')
		await flushPromises()

		expect(sdk.oauthAuthorize).toHaveBeenCalledWith({body: expect.objectContaining({
			scope: 'mcp:read mcp:write',
			resource: RESOURCE,
			state: 'st-1',
		})})
		expect(window.location.href).toBe(`${REDIRECT_URI}?code=the-code&state=st-1`)
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

	it('redirects a denial with access_denied and state without authorizing', async () => {
		mockRoute.query = {
			...BASE_QUERY,
			scope: 'mcp:read',
		}
		sdk.oauthAuthorizePreview.mockResolvedValue(preview(['mcp:read']))
		const wrapper = mountView()
		await flushPromises()

		await button(wrapper, 'Deny').trigger('click')
		await flushPromises()

		expect(sdk.oauthAuthorize).not.toHaveBeenCalled()
		expect(window.location.href).toBe(`${REDIRECT_URI}?error=access_denied&state=st-1`)
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
