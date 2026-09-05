import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
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
const create = vi.fn(async token => ({...token, id: 1, token: 'tk_test'}))

vi.mock('@/services/apiToken', () => ({
	default: class {
		loading = false
		getAvailableRoutes = getAvailableRoutes
		create = create
	},
}))

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

function mountForm() {
	const errors: unknown[] = []
	const wrapper = mount(ApiTokenForm, {
		global: {
			plugins: [i18n],
			stubs: {
				flatPickr: true,
				XButton: {template: '<button v-bind="$attrs"><slot /></button>'},
			},
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
})
