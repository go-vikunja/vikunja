import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import General from './General.vue'
import testid from '@/directives/testid'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

vi.mock('@/helpers/fetcher', () => {
	const httpStub = () => ({
		get: vi.fn(async () => ({data: []})),
		post: vi.fn(async () => ({data: {}})),
		interceptors: {request: {use: vi.fn()}, response: {use: vi.fn()}},
	})
	return {AuthenticatedHTTPFactory: httpStub, HTTPFactory: httpStub}
})

// Avoid the avatar request triggered by setUser.
vi.mock('@/models/user', async (importOriginal) => {
	const original = await importOriginal<typeof import('@/models/user')>()
	return {
		...original,
		fetchAvatarBlobUrl: vi.fn(async () => ''),
		invalidateAvatarCache: vi.fn(),
	}
})

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
			plugins: [i18n, router],
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
		useAuthStore().setUser(null)

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).not.toContain('managed by')
	})

	it('marks a non-local user as external', async () => {
		useAuthStore().setUser({
			id: 1,
			username: 'user1',
			isLocalUser: false,
			authProvider: 'keycloak',
		} as never)

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).toContain('keycloak')
	})

	it('shows job title and department read-only when the user has them', async () => {
		useAuthStore().setUser({
			id: 1,
			username: 'user1',
			isLocalUser: false,
			authProvider: 'Microsoft Entra ID',
			settings: {jobTitle: 'Engineer', department: 'Information Technology'},
		} as never)

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).toContain('Job title')
		expect(wrapper.text()).toContain('Department')
		const values = wrapper.findAll('input').map(input => (input.element as HTMLInputElement).value)
		expect(values).toContain('Engineer')
		expect(values).toContain('Information Technology')
		// The fields are display-only: nobody can type into them.
		const titleInput = wrapper.findAll('input').find(input => (input.element as HTMLInputElement).value === 'Engineer')
		expect(titleInput?.attributes('disabled')).toBeDefined()
		expect(titleInput?.attributes('readonly')).toBeDefined()
	})

	it('hides the job title and department fields when there is nothing to show', async () => {
		useAuthStore().setUser({
			id: 2,
			username: 'user2',
			isLocalUser: true,
			settings: {jobTitle: '', department: ''},
		} as never)

		wrapper = await mountComponent()
		await flushPromises()

		expect(errors).toEqual([])
		expect(wrapper.text()).not.toContain('Job title')
		expect(wrapper.text()).not.toContain('Department')
	})
})
