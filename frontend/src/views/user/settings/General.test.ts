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
import {removeToken, saveToken} from '@/helpers/auth'
import testid from '@/directives/testid'
import {useAuthStore} from '@/stores/auth'
import en from '@/i18n/lang/en.json'

const sdk = vi.hoisted(() => ({
	userTimezones: vi.fn(async () => ({data: []})),
	patchUserSettingsRead: vi.fn(),
	userShow: vi.fn(),
	userGetAvatarProvider: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)

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
				CustomTransition: {template: '<div><slot /></div>'},
				XButton: {template: '<button><slot /></button>'},
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
		vi.clearAllMocks()
		queryClient.clear()
		setActivePinia(createPinia())
		removeToken()
		errors = []
		sdk.userGetAvatarProvider.mockResolvedValue({data: {avatar_provider: 'default'}})
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

	it('keeps edits made while a save is in flight', async () => {
		const {nameInput, release} = await startPendingSave()
		await nameInput.setValue('Newer')
		release()
		await flushPromises()

		expect((nameInput.element as HTMLInputElement).value).toBe('Newer')
		expect(wrapper?.find('.sticky-save button').exists()).toBe(true)
	})

	it('sends only what changed since the last save', async () => {
		const {nameInput, release} = await startPendingSave()
		await wrapper!.find('input[type="time"]').setValue('17:30')
		release()
		await flushPromises()

		await nameInput.trigger('keyup.enter')
		await flushPromises()

		expect(sdk.patchUserSettingsRead).toHaveBeenCalledTimes(2)
		expect(sdk.patchUserSettingsRead.mock.calls[1][0].body).toStrictEqual({frontend_settings: {default_due_time: '17:30'}})
	})

	it('ignores a submit while a save is in flight', async () => {
		const {nameInput, release} = await startPendingSave()
		await nameInput.setValue('Old')
		await nameInput.trigger('keyup.enter')
		await flushPromises()
		release()
		await flushPromises()

		expect(sdk.patchUserSettingsRead).toHaveBeenCalledOnce()
		expect((nameInput.element as HTMLInputElement).value).toBe('Old')
		expect(wrapper?.find('.sticky-save button').exists()).toBe(true)
	})
})

function seedAccount() {
	saveToken(`header.${btoa(JSON.stringify({id: 1, type: AUTH_TYPES.USER}))}.signature`, false)
	useAuthStore().setSession({
		id: 1,
		type: AUTH_TYPES.USER,
		exp: 0,
	})
	queryClient.setQueryData(accountKeys.user(1), {
		id: 1,
		username: 'user1',
		is_local_user: true,
		settings: {
			name: 'Old',
			frontend_settings: {sidebar_width: 376},
		},
	})
	sdk.userShow.mockResolvedValue({
		data: {
			id: 1,
			username: 'user1',
			is_local_user: true,
			settings: {name: 'Old'},
		},
	})
}

async function startPendingSave() {
	seedAccount()
	let release = () => {}
	sdk.patchUserSettingsRead.mockImplementation(() => new Promise(resolve => {
		release = () => resolve({
			data: {},
			response: {status: 200},
		})
	}))
	wrapper = await mountComponent()
	await flushPromises()

	const nameInput = wrapper.find('input[type="text"]')
	await nameInput.setValue('New')
	await nameInput.trigger('keyup.enter')
	await flushPromises()
	expect(sdk.patchUserSettingsRead).toHaveBeenCalledOnce()
	expect(sdk.patchUserSettingsRead.mock.calls[0][0].body).toStrictEqual({name: 'New'})
	return {
		nameInput,
		release: () => release(),
	}
}
