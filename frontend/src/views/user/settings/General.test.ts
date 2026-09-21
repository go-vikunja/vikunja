import {VueQueryPlugin} from '@tanstack/vue-query'
import {queryClient} from '@/client/queryClient'
import {accountKeys} from '@/client/queries/account'
import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import General from './General.vue'
import {AUTH_TYPES} from '@/constants/auth'
import testid from '@/directives/testid'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

vi.mock('@/client/generated', () => ({userTimezones: vi.fn(async () => ({data: []}))}))

vi.mock('@/message', () => ({
	success: vi.fn(),
	error: vi.fn(),
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

let wrapper: VueWrapper | undefined
let errors: unknown[] = []

async function mountComponent() {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/user/settings/general', name: 'user.settings.general', component: General}],
	})
	await router.push('/user/settings/general')
	await router.isReady()

	return mount(General, {
		global: {
			plugins: [i18n, router, [VueQueryPlugin, {queryClient}]],
			directives: {cy: testid, focus: () => {}},
			stubs: {
				Card: {template: '<div><slot /></div>'},
				ProjectSearch: true,
				Multiselect: true,
				FormSelect: true,
				FormCheckbox: true,
				ShortcutRecorder: true,
				Reminders: true,
				CustomTransition: true,
			},
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
	})
}

describe('General user settings', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		errors = []
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
	})

	// Logout cleared the user while this view was still the current route (FRONTEND-OSS-2CJ).
	it('renders without a logged in user', async () => {
		useAuthStore().setSession(null)

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).not.toContain('managed by')
	})

	it('marks a non-local user as external', async () => {
		useAuthStore().setSession({
			id: 1,
			type: AUTH_TYPES.USER,
			exp: 0,
		})
		queryClient.setQueryData(accountKeys.user(1), {
			id: 1,
			username: 'user1',
			is_local_user: false,
			auth_provider: 'keycloak',
		})

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).toContain('keycloak')
	})
})
