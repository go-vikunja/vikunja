import {
	afterEach,
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {flushPromises, shallowMount} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {createI18n} from 'vue-i18n'
import type {Directive} from 'vue'

import en from '@/i18n/lang/en.json'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(async () => ({data: {
		items: [],
		total_pages: 1,
	}})),
	timeEntriesCreate: vi.fn(async ({body}: {body: Record<string, unknown>}) => ({data: {
		...body,
		id: 1,
		user_id: 1,
	}})),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({
	error: vi.fn(),
	success: vi.fn(),
}))

import TimeEntryForm from './TimeEntryForm.vue'
import Datepicker from '@/components/input/Datepicker.vue'

const FORM_OPENED = new Date('2026-06-07T09:00:00Z')
const CLICKED = new Date('2026-06-07T09:30:00Z')

const cy: Directive<HTMLElement, string> = {
	mounted(el, {value}) {
		el.setAttribute('data-cy', value)
	},
}

function mountForm() {
	const pinia = createPinia()
	setActivePinia(pinia)
	return shallowMount(TimeEntryForm, {
		props: {taskId: 1},
		global: {
			plugins: [
				pinia,
				createI18n({
					legacy: false,
					locale: 'en',
					messages: {en},
				}),
				[VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}],
			],
			stubs: {
				XButton: {template: '<button v-bind="$attrs"><slot /></button>'},
				Icon: true,
			},
			directives: {
				tooltip: {},
				cy,
			},
		},
	})
}

async function startTimerAt(wrapper: ReturnType<typeof mountForm>, now: Date) {
	vi.setSystemTime(now)
	await wrapper.find('[data-cy="startTimer"]').trigger('click')
	await flushPromises()
	return sdk.timeEntriesCreate.mock.calls.at(-1)?.[0].body
}

describe('TimeEntryForm start timer', () => {
	beforeEach(() => {
		sdk.timeEntriesCreate.mockClear()
		vi.useFakeTimers({toFake: ['Date']})
		vi.setSystemTime(FORM_OPENED)
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('starts at click time when the start was left untouched', async () => {
		const wrapper = mountForm()

		const body = await startTimerAt(wrapper, CLICKED)

		expect(body?.start_time).toBe(CLICKED.toISOString())
		expect(body?.end_time).toBeUndefined()
	})

	it('starts at the chosen start time', async () => {
		const wrapper = mountForm()
		const chosen = new Date('2026-06-07T08:15:00Z')
		wrapper.findAllComponents(Datepicker)[0]!.vm.$emit('update:modelValue', chosen)

		const body = await startTimerAt(wrapper, CLICKED)

		expect(body?.start_time).toBe(chosen.toISOString())
	})
})
