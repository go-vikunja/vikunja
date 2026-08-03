import {describe, it, expect, beforeEach, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import ApiTokenForm from './ApiTokenForm.vue'
import en from '@/i18n/lang/en.json'

const getAvailableRoutes = vi.fn(async () => ({
	tasks: {
		create: {path: '/api/v1/projects/:project/tasks', method: 'PUT'},
		read_all: {path: '/api/v1/tasks', method: 'GET'},
		read_one: {path: '/api/v1/tasks/:projecttask', method: 'GET'},
		update: {path: '/api/v1/tasks/:projecttask', method: 'POST'},
	},
	projects: {
		read_all: {path: '/api/v1/projects', method: 'GET'},
		read_one: {path: '/api/v1/projects/:project', method: 'GET'},
	},
	labels: {
		create: {path: '/api/v1/labels', method: 'PUT'},
		read_all: {path: '/api/v1/labels', method: 'GET'},
		read_one: {path: '/api/v1/labels/:label', method: 'GET'},
	},
}))

vi.mock('@/services/apiToken', () => ({
	default: class {
		loading = false
		getAvailableRoutes = getAvailableRoutes
		create = vi.fn(async (token) => ({...token, id: 1, token: 'tk_test'}))
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
	})
	return {wrapper, errors}
}

describe('ApiTokenForm', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		getAvailableRoutes.mockClear()
	})

	it('toggles permission checkboxes without Vue runtime errors', async () => {
		const {wrapper, errors} = mountForm()
		await flushPromises()

		const title = wrapper.find('#apiTokenTitle')
		expect(title.exists()).toBe(true)
		await title.setValue('chrome-quick-add')

		const checkboxes = wrapper.findAll('input[type="checkbox"]')
		expect(checkboxes.length).toBeGreaterThan(3)

		await checkboxes[0].setValue(true)
		await flushPromises()
		await checkboxes[2].setValue(true)
		await flushPromises()

		const expiry = wrapper.find('#apiTokenExpiry')
		await expiry.setValue('custom')
		await flushPromises()
		await expiry.setValue('60')
		await flushPromises()

		const presetButtons = wrapper.findAll('button').filter(b =>
			/read only|tasks|projects|full access/i.test(b.text()),
		)
		if (presetButtons.length > 0) {
			await presetButtons[0].trigger('click')
			await flushPromises()
		}

		const runtimeErrors = errors.map(String)
		expect(runtimeErrors.filter(e => e.includes('emitsOptions'))).toEqual([])
		expect(runtimeErrors).toEqual([])
	})
})
