import {describe, it, expect, beforeEach, vi} from 'vitest'
import {mount, flushPromises} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import ApiTokens from '@/views/user/settings/ApiTokens.vue'
import Modal from '@/components/misc/Modal.vue'
import en from '@/i18n/lang/en.json'

const tokens = [
	{
		id: 1,
		title: 'chrome-quick-add',
		token: '',
		permissions: {tasks: ['create', 'read_all']},
		expiresAt: new Date('2036-01-01'),
		created: new Date('2026-01-01'),
	},
]

const getAll = vi.fn(async () => tokens.slice())
const getAvailableRoutes = vi.fn(async () => ({
	tasks: {
		create: {path: '/t', method: 'PUT'},
		read_all: {path: '/t', method: 'GET'},
	},
	projects: {
		read_all: {path: '/p', method: 'GET'},
	},
}))
const del = vi.fn(async () => ({}))

vi.mock('@/services/apiToken', () => ({
	default: class {
		loading = false
		getAll = getAll
		getAvailableRoutes = getAvailableRoutes
		delete = del
		create = vi.fn(async (token) => ({...token, id: 2, token: 'tk_new'}))
	},
}))

HTMLDialogElement.prototype.showModal = function () {
	this.setAttribute('open', '')
}
HTMLDialogElement.prototype.close = function () {
	this.removeAttribute('open')
}

const i18n = createI18n({legacy: false, locale: 'en', messages: {en}})

async function mountPage() {
	const errors: unknown[] = []
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/user/settings/api-tokens', name: 'user.settings.apiTokens', component: ApiTokens}],
	})
	await router.push('/user/settings/api-tokens')
	await router.isReady()

	const wrapper = mount(ApiTokens, {
		global: {
			plugins: [i18n, router],
			components: {Modal},
			stubs: {
				flatPickr: true,
				Card: {template: '<div><slot /></div>'},
				XButton: {
					template: '<button type="button" v-bind="$attrs" @click="$emit(\'click\', $event)"><slot /></button>',
					emits: ['click'],
				},
				BaseButton: {template: '<a v-bind="$attrs"><slot /></a>'},
				Message: {template: '<div class="message"><slot /></div>'},
				Icon: true,
			},
			config: {
				errorHandler(err) {
					errors.push(err)
				},
			},
		},
		attachTo: document.body,
	})
	await flushPromises()
	return {wrapper, errors}
}

function runtimeErrorMessages(errors: unknown[]) {
	return errors.map(e => (e instanceof Error ? e.message : String(e)))
}

describe('ApiTokens settings page', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		document.body.innerHTML = ''
		getAll.mockImplementation(async () => tokens.slice())
		getAvailableRoutes.mockClear()
		del.mockClear()
	})

	it('opens create form and edits fields without Vue runtime errors', async () => {
		const {wrapper, errors} = await mountPage()

		const createBtn = wrapper.findAll('button').find(b => /create a token/i.test(b.text()))
		expect(createBtn).toBeTruthy()
		await createBtn!.trigger('click')
		await flushPromises()

		const title = wrapper.find('#apiTokenTitle')
		expect(title.exists()).toBe(true)
		await title.setValue('another-token')
		await flushPromises()

		const checkboxes = wrapper.findAll('input[type="checkbox"]')
		expect(checkboxes.length).toBeGreaterThan(0)
		await checkboxes[0].setValue(true)
		await flushPromises()

		expect(runtimeErrorMessages(errors)).toEqual([])
	})

	it('deletes a token without crashing while the modal close animation is in flight', async () => {
		const {wrapper, errors} = await mountPage()

		const deleteBtn = wrapper.findAll('button').find(b => /delete/i.test(b.text()))
		expect(deleteBtn).toBeTruthy()
		await deleteBtn!.trigger('click')
		await flushPromises()

		expect(document.querySelector('dialog.modal-dialog')).toBeTruthy()

		const buttons = Array.from(document.querySelectorAll('dialog button')) as HTMLButtonElement[]
		buttons[buttons.length - 1].click()
		await flushPromises()
		// Modal keeps the dialog mounted for its close transition.
		await new Promise(r => setTimeout(r, 200))
		await flushPromises()

		expect(del).toHaveBeenCalled()
		expect(runtimeErrorMessages(errors)).toEqual([])
	})
})
