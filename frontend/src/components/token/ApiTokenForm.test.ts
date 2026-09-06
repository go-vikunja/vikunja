import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type DOMWrapper, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import ApiTokenForm from './ApiTokenForm.vue'
import en from '@/i18n/lang/en.json'

const getAvailableRoutes = vi.fn(async () => ({
	tasks: {
		create: {path: '/api/v1/projects/:project/tasks', method: 'PUT'},
		read_all: {path: '/api/v1/tasks', method: 'GET'},
	},
	projects: {
		read_all: {path: '/api/v1/projects', method: 'GET'},
		read_one: {path: '/api/v1/projects/:project', method: 'GET'},
	},
}))
const create = vi.fn(async (token: Record<string, unknown>) => ({...token, id: 1, token: 'tk_test'}))

vi.mock('@/services/apiToken', () => ({
	default: class {
		loading = false
		getAvailableRoutes = getAvailableRoutes
		create = create
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountForm({stubDatePicker = true} = {}) {
	const errors: unknown[] = []
	const stubs: Record<string, unknown> = {
		XButton: {template: '<button v-bind="$attrs"><slot /></button>'},
	}
	if (stubDatePicker) {
		stubs.flatPickr = true
	}
	const wrapper = mount(ApiTokenForm, {
		global: {
			plugins: [i18n],
			stubs,
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
		attachTo: document.body,
	})
	return {wrapper, errors}
}

function warningAfterCheckbox(wrapper: DOMWrapper<Element>, label: string) {
	const checkbox = wrapper.findAll('.fancy-checkbox').find(c => c.find('.fancy-checkbox__content').text() === label)
	expect(checkbox).toBeTruthy()
	const next = checkbox!.element.nextElementSibling
	return next?.tagName === 'P' ? next.textContent?.trim() : undefined
}

function setTitleFieldRef(wrapper: VueWrapper, value: unknown) {
	const {setupState} = wrapper.vm.$ as unknown as {setupState: Record<string, unknown>}
	setupState.apiTokenTitle = value
}

describe('ApiTokenForm', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		setActivePinia(createPinia())
		getAvailableRoutes.mockClear()
		create.mockClear()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		document.body.innerHTML = ''
	})

	it('focuses the title field when submitting without a title', async () => {
		const mounted = mountForm()
		wrapper = mounted.wrapper
		await flushPromises()

		;(document.activeElement as HTMLElement | null)?.blur()
		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(create).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('The title is required')
		expect(document.activeElement).toBe(wrapper.find('#apiTokenTitle').element)
		expect(mounted.errors).toEqual([])
	})

	it('does not crash when the title field ref is gone on submit', async () => {
		const mounted = mountForm()
		wrapper = mounted.wrapper
		await flushPromises()

		// Vue nulls template refs when the element unmounts, which is what the
		// submit handler saw in FRONTEND-OSS-2H1.
		setTitleFieldRef(wrapper, null)
		await wrapper.find('form').trigger('submit')
		await flushPromises()

		expect(create).not.toHaveBeenCalled()
		expect(wrapper.text()).toContain('The title is required')
		expect(mounted.errors).toEqual([])
	})

	it('survives switching the expiry between a preset and a custom date', async () => {
		const mounted = mountForm({stubDatePicker: false})
		wrapper = mounted.wrapper
		await flushPromises()

		const expiry = wrapper.find('#apiTokenExpiry')
		for (let i = 0; i < 3; i++) {
			await expiry.setValue('custom')
			await flushPromises()
			expect(wrapper.find('input[readonly]').exists()).toBe(true)

			await expiry.setValue('30')
			await flushPromises()
			expect(wrapper.find('input[readonly]').exists()).toBe(false)
		}

		expect(mounted.errors).toEqual([])
	})

	it('warns about root-equivalent admin scopes only', async () => {
		getAvailableRoutes.mockResolvedValueOnce({
			admin: {
				users_list: {path: '/api/v2/admin/users', method: 'GET'},
				users_set_admin: {path: '/api/v2/admin/users/{id}/admin', method: 'PATCH'},
				users_set_password: {path: '/api/v2/admin/users/{id}/password', method: 'PATCH'},
			},
			users: {
				users_set_admin: {path: '/api/v2/users/{id}/admin', method: 'PATCH'},
			},
		})
		const mounted = mountForm()
		wrapper = mounted.wrapper
		await flushPromises()

		const warning = i18n.global.t('user.settings.apiTokens.escalationWarning')
		const [adminGroup, usersGroup] = wrapper.findAll('.mbe-2')

		expect(adminGroup.text()).toContain('admin')
		expect(warningAfterCheckbox(adminGroup, 'users set admin')).toBe(warning)
		expect(warningAfterCheckbox(adminGroup, 'users set password')).toBe(warning)
		expect(warningAfterCheckbox(adminGroup, 'users list')).toBeUndefined()

		// same key outside the admin group must not warn
		expect(usersGroup.text()).toContain('users')
		expect(warningAfterCheckbox(usersGroup, 'users set admin')).toBeUndefined()
		expect(mounted.errors).toEqual([])
	})
})
