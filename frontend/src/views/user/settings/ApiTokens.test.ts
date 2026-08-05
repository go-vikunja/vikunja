import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {mount, flushPromises, type VueWrapper} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createI18n} from 'vue-i18n'
import {createRouter, createMemoryHistory} from 'vue-router'
import ApiTokens from '@/views/user/settings/ApiTokens.vue'
import Modal from '@/components/misc/Modal.vue'
import testid from '@/directives/testid'
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
	{
		id: 2,
		title: 'backup-sync',
		token: '',
		permissions: {tasks: ['read_all']},
		expiresAt: new Date('2036-01-01'),
		created: new Date('2026-01-01'),
	},
]

const getAll = vi.fn(async () => tokens.slice())
const del = vi.fn(async () => ({}))

vi.mock('@/services/apiToken', () => ({
	default: class {
		loading = false
		getAll = getAll
		delete = del
	},
}))

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
			directives: {cy: testid},
			stubs: {
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

async function openDeleteModalForFirstToken(wrapper: VueWrapper) {
	const deleteBtn = wrapper.findAll('button').find(b => /delete/i.test(b.text()))
	expect(deleteBtn).toBeTruthy()
	await deleteBtn!.trigger('click')
	await flushPromises()

	const confirmBtn = document.querySelector<HTMLButtonElement>('dialog [data-cy="modalPrimary"]')
	expect(confirmBtn).toBeTruthy()
	return confirmBtn!
}

// Modal keeps the dialog mounted for its 150ms close transition.
async function settleCloseTransition() {
	await flushPromises()
	await new Promise(r => setTimeout(r, 200))
	await flushPromises()
}

function runtimeErrorMessages(errors: unknown[]) {
	return errors.map(e => (e instanceof Error ? e.message : String(e)))
}

describe('ApiTokens settings page', () => {
	let wrapper: VueWrapper | undefined

	beforeEach(() => {
		// Makes v-cy emit data-cy attributes so the modal's primary button is selectable.
		vi.stubGlobal('TESTING', true)
		setActivePinia(createPinia())
		document.body.innerHTML = ''
		del.mockClear()
	})

	afterEach(() => {
		wrapper?.unmount()
		wrapper = undefined
		vi.unstubAllGlobals()
		document.body.innerHTML = ''
	})

	it('deletes a token without crashing while the modal close animation is in flight', async () => {
		const mounted = await mountPage()
		wrapper = mounted.wrapper
		const {errors} = mounted

		const confirmBtn = await openDeleteModalForFirstToken(wrapper)
		confirmBtn.click()
		await settleCloseTransition()

		expect(del).toHaveBeenCalledTimes(1)
		expect(del).toHaveBeenCalledWith(expect.objectContaining({id: 1}))
		expect(document.querySelector('dialog.modal-dialog')).toBeNull()

		const rows = wrapper.findAll('tbody tr')
		expect(rows).toHaveLength(1)
		expect(rows[0].text()).toContain('backup-sync')
		expect(wrapper.text()).not.toContain('chrome-quick-add')
		expect(runtimeErrorMessages(errors)).toEqual([])
	})

	it('deletes only once when the confirm button is double clicked', async () => {
		const mounted = await mountPage()
		wrapper = mounted.wrapper
		const {errors} = mounted

		const confirmBtn = await openDeleteModalForFirstToken(wrapper)
		confirmBtn.click()
		confirmBtn.click()
		await settleCloseTransition()

		expect(del).toHaveBeenCalledTimes(1)
		expect(del).toHaveBeenCalledWith(expect.objectContaining({id: 1}))
		expect(runtimeErrorMessages(errors)).toEqual([])
	})
})
