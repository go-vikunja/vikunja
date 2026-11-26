import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {flushPromises, mount, type VueWrapper} from '@vue/test-utils'
import {createI18n} from 'vue-i18n'
import {createMemoryHistory, createRouter} from 'vue-router'
import {VueQueryPlugin, QueryClient} from '@tanstack/vue-query'
import {createPinia, setActivePinia} from 'pinia'
import ConnectedApps from './ConnectedApps.vue'
import Modal from '@/components/misc/Modal.vue'
import en from '@/i18n/lang/en.json'
import type {ApiToken} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	tokensList: vi.fn(),
	tokensDelete: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

const metadataApp: ApiToken = {
	id: 40,
	title: 'Claude Code',
	oauth_client_id: 'https://claude.ai/oauth/claude-code-client-metadata',
	oauth_access: 'write',
	permissions: {mcp: ['access']},
	created: '2026-11-01T10:00:00Z',
	expires_at: '2027-01-30T10:00:00Z',
}
const legacyApp: ApiToken = {
	id: 41,
	title: 'vikunja-desktop',
	oauth_client_id: 'vikunja-desktop',
	oauth_access: 'read',
	permissions: {mcp: ['access']},
	created: '2026-11-02T10:00:00Z',
	expires_at: '2020-01-01T00:00:00Z',
}
let stored: ApiToken[]
let wrapper: VueWrapper

beforeEach(() => {
	setActivePinia(createPinia())
	stored = [
		metadataApp,
		legacyApp,
	]
	sdk.tokensList.mockReset()
	sdk.tokensList.mockImplementation(async () => ({data: {
		items: stored,
		page: 1,
		per_page: 50,
		total: stored.length,
		total_pages: 1,
	}}))
	sdk.tokensDelete.mockReset()
	sdk.tokensDelete.mockImplementation(async ({path}: {path: {id: number}}) => {
		stored = stored.filter(token => token.id !== path.id)
		return {}
	})
})
afterEach(() => wrapper?.unmount())

async function mountList() {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{
			path: '/user/settings/mcp',
			name: 'user.settings.mcp',
			component: ConnectedApps,
		}],
	})
	await router.push('/user/settings/mcp')
	await router.isReady()
	wrapper = mount(ConnectedApps, {
		global: {
			plugins: [
				createI18n({
					legacy: false,
					locale: 'en',
					messages: {en},
				}),
				router,
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			components: {Modal},
			stubs: {
				Card: {template: '<div><slot /></div>'},
				XButton: {
					template: '<button type="button" @click="$emit(\'click\', $event)"><slot /></button>',
					emits: ['click'],
				},
				Icon: true,
			},
		},
		attachTo: document.body,
	})
	await flushPromises()
}

function dialogButton(text: string) {
	const button = [...document.querySelectorAll<HTMLButtonElement>('dialog button')].find(b => b.textContent?.trim() === text)
	expect(button).toBeTruthy()
	return button!
}

function rowTexts() {
	return wrapper.findAll('tbody tr').map(row => row.findAll('td').map(cell => cell.text()))
}

describe('ConnectedApps', () => {
	it('marks an expired grant', async () => {
		await mountList()
		expect(rowTexts()[1][3]).toContain('This token has expired')
	})

	it('keeps the grant when the confirmation is dismissed', async () => {
		await mountList()
		await wrapper.findAll('button').find(button => button.text() === 'Revoke')!.trigger('click')
		await flushPromises()
		dialogButton('Cancel').click()
		await flushPromises()
		expect(sdk.tokensDelete).not.toHaveBeenCalled()
		expect(rowTexts()).toHaveLength(2)
	})
})
