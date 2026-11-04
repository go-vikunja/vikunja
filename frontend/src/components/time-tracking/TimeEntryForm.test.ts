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
import {ref, type Directive} from 'vue'

import en from '@/i18n/lang/en.json'

const sdk = vi.hoisted(() => ({
	projectsList: vi.fn(async () => ({data: {
		items: [],
		total_pages: 1,
	}})),
	timeEntriesCreate: vi.fn(async ({body}: {body: Record<string, unknown>}) => ({data: {
		...body,
		id: 1,
		user: {id: 1},
	}})),
}))

vi.mock('@/client/generated', () => sdk)

const activeTimer = vi.hoisted(() => ({value: null as {start_time: string} | null}))
const spanning = vi.hoisted(() => ({value: [] as unknown[]}))
vi.mock('@/composables/useTimeTracking', () => ({
	useTimeTracking: () => ({activeTimer: ref(activeTimer.value)}),
	useTimeEntries: () => ({entries: ref(spanning.value)}),
}))
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

describe('TimeEntryForm backdated stop hint', () => {
	const hint = '[data-cy="backdatedStopHint"]'

	beforeEach(() => {
		vi.useFakeTimers({toFake: ['Date']})
		vi.setSystemTime(CLICKED)
	})

	afterEach(() => {
		activeTimer.value = null
		spanning.value = []
		vi.useRealTimers()
	})

	async function chooseStart(wrapper: ReturnType<typeof mountForm>, date: Date) {
		wrapper.findAllComponents(Datepicker)[0]!.vm.$emit('update:modelValue', date)
		await flushPromises()
	}

	it('shows when the chosen start backdates the running timer', async () => {
		activeTimer.value = {start_time: '2026-06-07T08:00:00Z'}
		const wrapper = mountForm()

		await chooseStart(wrapper, new Date('2026-06-07T09:15:00Z'))

		expect(wrapper.find(hint).exists()).toBe(true)
	})

	it('is hidden while the start is untouched', () => {
		activeTimer.value = {start_time: '2026-06-07T08:00:00Z'}
		const wrapper = mountForm()

		expect(wrapper.find(hint).exists()).toBe(false)
	})

	it('is hidden without a running timer', async () => {
		const wrapper = mountForm()

		await chooseStart(wrapper, new Date('2026-06-07T09:15:00Z'))

		expect(wrapper.find(hint).exists()).toBe(false)
	})

	it('shows when a completed entry spans the chosen start', async () => {
		spanning.value = [{id: 2}]
		const wrapper = mountForm()

		await chooseStart(wrapper, new Date('2026-06-07T09:15:00Z'))

		expect(wrapper.find(hint).exists()).toBe(true)
	})

	it('is hidden for a start in the future', async () => {
		activeTimer.value = {start_time: '2026-06-07T08:00:00Z'}
		spanning.value = [{id: 2}]
		const wrapper = mountForm()

		await chooseStart(wrapper, new Date('2026-06-07T10:00:00Z'))

		expect(wrapper.find(hint).exists()).toBe(false)
	})

	it('is hidden when the chosen start is before the running timer', async () => {
		activeTimer.value = {start_time: '2026-06-07T08:00:00Z'}
		const wrapper = mountForm()

		await chooseStart(wrapper, new Date('2026-06-07T07:00:00Z'))

		expect(wrapper.find(hint).exists()).toBe(false)
	})
})
