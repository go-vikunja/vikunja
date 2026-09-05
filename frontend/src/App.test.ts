import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import App from '@/App.vue'
import {useAuthStore} from '@/stores/auth'
import {AUTH_TYPES} from '@/modelTypes/IUser'
import en from '@/i18n/lang/en.json'

vi.mock('@/helpers/fetcher', () => {
	const httpStub = () => Object.assign(
		vi.fn(async () => ({data: new Blob()})),
		{
			get: vi.fn(async () => ({data: []})),
			post: vi.fn(async () => ({data: {}})),
			interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}},
		},
	)
	return {AuthenticatedHTTPFactory: httpStub, HTTPFactory: httpStub}
})

vi.mock('@/models/user', async (importOriginal) => {
	const original = await importOriginal<typeof import('@/models/user')>()
	return {
		...original,
		fetchAvatarBlobUrl: vi.fn(async () => ''),
		invalidateAvatarCache: vi.fn(),
	}
})

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

const AppRoute = {template: '<div class="app-route">app route</div>'}
const LoginRoute = {template: '<div class="login-route">login route</div>'}

let wrapper: VueWrapper | undefined

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

	wrapper = mount(App, {
		global: {
			plugins: [i18n, router],
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

describe('App layout', () => {
	beforeEach(() => setActivePinia(createPinia()))

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
	})

	it('renders the login route in the logged out shell', async () => {
		await mountApp('/login')

		expect(wrapper!.find('.login-route').exists()).toBe(true)
	})

	// Logout clears the user before the navigation to /login lands. Rendering the
	// still-current app route in the logged out shell remounts components that
	// dereference authStore.info (FRONTEND-OSS-2CJ, FRONTEND-OSS-2CH).
	it('does not render an app route in the logged out shell after the user is cleared', async () => {
		const authStore = useAuthStore()
		authStore.setAuthenticated(true)
		authStore.setUser({id: 1, username: 'user1', type: AUTH_TYPES.USER} as never)

		await mountApp('/labels')
		expect(wrapper!.findComponent({name: 'ContentAuth'}).exists()).toBe(true)

		authStore.setAuthenticated(false)
		authStore.setUser(null)
		await flushPromises()

		expect(wrapper!.find('.no-auth').exists()).toBe(true)
		expect(wrapper!.find('.app-route').exists()).toBe(false)
	})
})
