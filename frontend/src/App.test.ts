import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia, type Pinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import App from '@/App.vue'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {useAuthStore} from '@/stores/auth'
import {AUTH_TYPES} from '@/constants/auth'
import en from '@/i18n/lang/en.json'

const sdk = vi.hoisted(() => ({
	userDeletionConfirm: vi.fn(),
	userShow: vi.fn(),
}))

vi.mock('@/client/generated', async importOriginal => ({
	...await importOriginal<typeof import('@/client/generated')>(),
	userDeletionConfirm: sdk.userDeletionConfirm,
	userShow: sdk.userShow,
}))

vi.mock('@/message', () => ({success: vi.fn(), error: vi.fn()}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

const AppRoute = {template: '<div class="app-route">app route</div>'}
const LoginRoute = {template: '<div class="login-route">login route</div>'}

let wrapper: VueWrapper | undefined
let queryClient: QueryClient | undefined
// Store actions left running by an earlier test re-activate that test's pinia.
let pinia: Pinia

async function mountApp(path: string) {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/labels', name: 'labels.index', component: AppRoute},
			{path: '/login', name: 'user.login', component: LoginRoute},
		],
	})
	await router.push(path)
	await router.isReady()

	queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
	wrapper = mount(App, {
		global: {
			plugins: [pinia, i18n, router, [VueQueryPlugin, {queryClient}]],
			stubs: {
				Ready: {template: '<div><slot /></div>'},
				NoAuthWrapper: {template: '<div class="no-auth"><slot /></div>'},
				AppHeader: true,
				ContentAuth: true,
				ContentLinkShare: true,
				KeyboardShortcuts: true,
				Notification: true,
				UpdateNotification: true,
				AddToHomeScreen: true,
				DemoMode: true,
				QuickAddOverlay: true,
			},
		},
	})
	await flushPromises()
	return {router}
}

function login() {
	const authStore = useAuthStore(pinia)
	authStore.setAuthenticated(true)
	authStore.setSession({
		id: 1,
		type: AUTH_TYPES.USER,
		exp: 0,
	})
	return authStore
}

describe('App layout', () => {
	beforeEach(() => {
		pinia = createPinia()
		setActivePinia(pinia)
		sdk.userDeletionConfirm.mockResolvedValue({})
		sdk.userShow.mockResolvedValue({
			data: {
				id: 1,
				deletion_scheduled_at: '2030-01-01T00:00:00Z',
			},
		})
	})

	afterEach(() => {
		vi.clearAllMocks()
		wrapper?.unmount()
		wrapper = undefined
		queryClient?.clear()
	})

	// Logout clears the user before the navigation to /login lands. Rendering the
	// still-current app route in the logged out shell remounts components that
	// dereference authStore.info (FRONTEND-OSS-2CJ, FRONTEND-OSS-2CH).
	it('does not render an app route in the logged out shell after the user is cleared', async () => {
		const authStore = login()

		await mountApp('/labels')
		expect(wrapper!.findComponent({name: 'ContentAuth'}).exists()).toBe(true)

		authStore.setAuthenticated(false)
		authStore.setSession(null)
		await flushPromises()

		expect(wrapper!.find('.no-auth').exists()).toBe(true)
		expect(wrapper!.find('.app-route').exists()).toBe(false)
	})

	it('keeps the deletion confirmation token in the url while logged out', async () => {
		const {router} = await mountApp('/login?accountDeletionConfirm=token-1')

		expect(sdk.userDeletionConfirm).not.toHaveBeenCalled()
		expect(router.currentRoute.value.query.accountDeletionConfirm).toBe('token-1')
	})

	it('spends the deletion confirmation token once when the user logs in', async () => {
		const {router} = await mountApp('/login?accountDeletionConfirm=token-1')
		expect(sdk.userDeletionConfirm).not.toHaveBeenCalled()

		login()
		await flushPromises()
		await flushPromises()

		expect(sdk.userDeletionConfirm).toHaveBeenCalledTimes(1)
		expect(sdk.userDeletionConfirm).toHaveBeenCalledWith({body: {token: 'token-1'}})
		expect(router.currentRoute.value.query.accountDeletionConfirm).toBeUndefined()
	})
})
