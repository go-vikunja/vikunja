import {defineComponent, h, nextTick, reactive, ref, type Ref} from 'vue'
import {mount} from '@vue/test-utils'
import {
	afterEach,
	beforeEach,
	describe,
	expect,
	it,
	vi,
} from 'vitest'
import {
	onlineManager,
	QueryClient,
	VueQueryPlugin,
} from '@tanstack/vue-query'

import type {PreviewAnswer, TaskRepeatWritable} from '@/client/generated'
import {ApiError} from '@/client/problem'
import {
	isRefusal,
	previewFixtures,
} from '@/client/queries/occurrencePreview.fixtures'
import {
	PREVIEW_DEBOUNCE_MS,
	previewInputFor,
	useOccurrencePreview,
	type OccurrencePreviewSource,
} from './useOccurrencePreview'

const sdk = vi.hoisted(() => ({
	tasksOccurrencesPreview: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
const message = vi.hoisted(() => ({
	error: vi.fn(),
	success: vi.fn(),
}))
vi.mock('@/message', () => message)

function answerOf(row: string): PreviewAnswer {
	const answer = previewFixtures.find(fixture => fixture.row === row)?.answer
	if (!answer || isRefusal(answer)) throw new Error(`no answer for ${row}`)
	return answer
}

type Pending = {
	body: Record<string, unknown>,
	signal: AbortSignal,
	resolve: (answer: PreviewAnswer) => void,
	reject: (cause: unknown) => void,
}

let requests: Pending[]
let client: QueryClient

beforeEach(() => {
	vi.useFakeTimers()
	vi.clearAllMocks()
	requests = []
	sdk.tasksOccurrencesPreview.mockImplementation(({body, signal}: {body: Record<string, unknown>, signal: AbortSignal}) =>
		new Promise((resolve, reject) => {
			requests.push({
				body,
				signal,
				resolve: answer => resolve({data: answer}),
				reject,
			})
		}))
	client = new QueryClient()
})

afterEach(() => {
	client.clear()
	onlineManager.setOnline(true)
	vi.useRealTimers()
})

function source(overrides: Partial<OccurrencePreviewSource> = {}): OccurrencePreviewSource {
	return {
		repeat: {freq: 'daily'},
		repeatClass: 'editable',
		dueDate: null,
		initialDueTime: '09:00',
		taskId: 7,
		done: false,
		repeatsFromCurrentDate: false,
		...overrides,
	}
}

async function settle(ms = 0) {
	await vi.advanceTimersByTimeAsync(ms)
	await nextTick()
}

function mountPreview(input: Ref<OccurrencePreviewSource | null>) {
	let preview: ReturnType<typeof useOccurrencePreview> | undefined
	const wrapper = mount(defineComponent({
		setup() {
			preview = useOccurrencePreview(input)
			return () => h('div')
		},
	}), {global: {plugins: [[VueQueryPlugin, {queryClient: client}]]}})
	return {wrapper, preview: preview!}
}

describe('previewInputFor', () => {
	it('sends the due date when the task is dated, done or not', () => {
		const dueDate = new Date('2026-10-29T08:00:00Z')
		for (const done of [false, true]) {
			expect(previewInputFor(source({dueDate, done}))).toEqual({
				repeat: {freq: 'daily'},
				due_date: '2026-10-29T08:00:00.000Z',
				task_id: 7,
				repeats_from_current_date: false,
				count: 10,
			})
		}
	})

	it('sends the fixed time when the task is undated and not done', () => {
		expect(previewInputFor(source({repeatsFromCurrentDate: true}))).toEqual({
			repeat: {freq: 'daily'},
			initial_due_time: '09:00',
			task_id: 7,
			repeats_from_current_date: true,
			count: 10,
		})
	})

	it.each([
		['a locked rule', source({repeatClass: 'locked'})],
		['a foreign rule', source({repeatClass: 'foreign'})],
		['no rule', source({repeat: null, repeatClass: 'none'})],
		['an undated done task', source({done: true})],
		['an undated task without a fixed time', source({initialDueTime: null})],
		['a task that is not saved yet', source({taskId: 0})],
		['no source', null],
	])('asks nothing for %s', (_name, value) => {
		expect(previewInputFor(value)).toBeNull()
	})
})

describe('useOccurrencePreview', () => {
	it('asks once on open and once per settled change', async () => {
		const input = ref(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		expect(requests).toHaveLength(1)
		requests[0].resolve(answerOf('P-1'))
		await settle()
		expect(preview.state.value).toBe('ready')

		for (const time of ['10:00', '11:00', '12:00']) {
			input.value = source({initialDueTime: time})
			await settle(PREVIEW_DEBOUNCE_MS - 50)
		}
		expect(requests).toHaveLength(1)
		await settle(50)

		expect(requests).toHaveLength(2)
		expect(requests[1].body.initial_due_time).toBe('12:00')
		expect(requests[1].body).not.toHaveProperty('timezone')
		wrapper.unmount()
	})

	it('keeps the previous answer as stale while a newer request loads', async () => {
		const input = ref(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		requests[0].resolve(answerOf('P-1'))
		await settle()
		const first = preview.view.value
		expect(first?.dates[0]).toBe('2026-10-29T09:00:00+01:00')

		input.value = source({initialDueTime: '18:00', repeatsFromCurrentDate: true, repeat: {freq: 'hourly'}})
		await settle()
		expect(preview.state.value).toBe('stale')
		expect(preview.view.value).toEqual(first)

		await settle(PREVIEW_DEBOUNCE_MS)
		expect(requests).toHaveLength(2)
		expect(preview.state.value).toBe('stale')
		expect(preview.view.value).toEqual(first)

		requests[1].resolve(answerOf('P-13'))
		await settle()
		expect(preview.state.value).toBe('ready')
		expect(preview.view.value?.dates[0]).toBe('2026-10-28T18:00:00+01:00')
		expect(preview.view.value?.forecast).toBe(true)
		wrapper.unmount()
	})

	it('shows a cached answer as stale while it is asked again', async () => {
		const input = ref(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		requests[0].resolve(answerOf('P-1'))
		await settle()

		input.value = source({initialDueTime: '18:00', repeatsFromCurrentDate: true, repeat: {freq: 'hourly'}})
		await settle(PREVIEW_DEBOUNCE_MS)
		requests[1].resolve(answerOf('P-13'))
		await settle()
		expect(preview.state.value).toBe('ready')

		input.value = source()
		await settle(PREVIEW_DEBOUNCE_MS)
		expect(requests).toHaveLength(3)
		expect(preview.view.value?.dates[0]).toBe('2026-10-29T09:00:00+01:00')
		expect(preview.state.value).toBe('stale')

		requests[2].resolve(answerOf('P-1'))
		await settle()
		expect(preview.state.value).toBe('ready')
		wrapper.unmount()
	})

	it('never shows a late answer for an older request', async () => {
		const input = ref(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		requests[0].resolve(answerOf('P-1'))
		await settle()

		input.value = source({initialDueTime: '10:00'})
		await settle(PREVIEW_DEBOUNCE_MS)
		input.value = source({initialDueTime: '18:00', repeatsFromCurrentDate: true, repeat: {freq: 'hourly'}})
		await settle(PREVIEW_DEBOUNCE_MS)
		expect(requests).toHaveLength(3)
		expect(requests[1].signal.aborted).toBe(true)

		requests[1].resolve(answerOf('P-5'))
		await settle()
		expect(preview.state.value).toBe('stale')
		expect(preview.view.value?.reason).toBeUndefined()
		expect(preview.view.value?.dates).toHaveLength(3)

		requests[2].resolve(answerOf('P-13'))
		await settle()
		expect(preview.state.value).toBe('ready')
		expect(preview.view.value?.dates[0]).toBe('2026-10-28T18:00:00+01:00')
		wrapper.unmount()
	})

	it('does not wait or keep the answer when the task changes', async () => {
		const input = ref(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		requests[0].resolve(answerOf('P-1'))
		await settle()

		input.value = source({taskId: 8})
		await settle()

		expect(requests).toHaveLength(2)
		expect(requests[1].body.task_id).toBe(8)
		expect(preview.state.value).toBe('loading')
		expect(preview.view.value).toBeNull()
		wrapper.unmount()
	})

	async function rapidEdits(input: Ref<OccurrencePreviewSource | null>, taskId: number) {
		for (const time of ['10:00', '11:00', '12:00']) {
			input.value = source({taskId, initialDueTime: time})
			await settle(PREVIEW_DEBOUNCE_MS - 50)
		}
	}

	it('debounces edits made right after a task switch', async () => {
		const input = ref<OccurrencePreviewSource | null>(source())
		const {wrapper} = mountPreview(input)
		await settle()
		input.value = source({taskId: 8})
		await settle()
		expect(requests).toHaveLength(2)

		await rapidEdits(input, 8)
		expect(requests).toHaveLength(2)
		await settle(50)

		expect(requests).toHaveLength(3)
		expect(requests[2].body).toMatchObject({task_id: 8, initial_due_time: '12:00'})
		wrapper.unmount()
	})

	it('debounces edits made right after the preview becomes enabled', async () => {
		const input = ref<OccurrencePreviewSource | null>(null)
		const {wrapper} = mountPreview(input)
		await settle()
		input.value = source()
		await settle()
		expect(requests).toHaveLength(1)

		await rapidEdits(input, 7)
		expect(requests).toHaveLength(1)
		await settle(50)

		expect(requests).toHaveLength(2)
		expect(requests[1].body.initial_due_time).toBe('12:00')
		wrapper.unmount()
	})

	it('debounces in-place edits of nested rule fields and arrays', async () => {
		const rule = reactive<TaskRepeatWritable>({freq: 'weekly', by_day: ['mo']})
		const {wrapper} = mountPreview(ref(source({repeat: rule})))
		await settle()
		expect(requests).toHaveLength(1)

		rule.interval = 2
		await settle(PREVIEW_DEBOUNCE_MS - 50)
		rule.by_day!.push('th')
		await settle(PREVIEW_DEBOUNCE_MS - 50)
		expect(requests).toHaveLength(1)
		await settle(50)

		expect(requests).toHaveLength(2)
		expect(requests[1].body.repeat).toEqual({freq: 'weekly', interval: 2, by_day: ['mo', 'th']})
		expect(requests[0].body.repeat).toEqual({freq: 'weekly', by_day: ['mo']})
		wrapper.unmount()
	})

	it('debounces an in-place array edit on its own', async () => {
		const rule = reactive<TaskRepeatWritable>({freq: 'weekly', by_day: ['mo']})
		const {wrapper} = mountPreview(ref(source({repeat: rule})))
		await settle()

		rule.by_day!.push('fr')
		await settle()
		expect(requests).toHaveLength(1)
		await settle(PREVIEW_DEBOUNCE_MS)

		expect(requests).toHaveLength(2)
		expect(requests[1].body.repeat).toEqual({freq: 'weekly', by_day: ['mo', 'fr']})
		wrapper.unmount()
	})

	it.each([
		['a locked rule', source({repeatClass: 'locked'})],
		['an undated done task', source({done: true})],
	])('makes no request for %s', async (_name, value) => {
		const {wrapper, preview} = mountPreview(ref(value))
		await settle(PREVIEW_DEBOUNCE_MS * 2)

		expect(sdk.tasksOccurrencesPreview).not.toHaveBeenCalled()
		expect(preview.state.value).toBe('disabled')
		expect(preview.view.value).toBeNull()
		wrapper.unmount()
	})

	it('stops asking at once when the rule becomes locked', async () => {
		const input = ref<OccurrencePreviewSource | null>(source())
		const {wrapper, preview} = mountPreview(input)
		await settle()
		requests[0].resolve(answerOf('P-1'))
		await settle()

		input.value = source({repeatClass: 'locked', initialDueTime: '10:00'})
		await settle(PREVIEW_DEBOUNCE_MS * 2)

		expect(requests).toHaveLength(1)
		expect(preview.state.value).toBe('disabled')
		expect(preview.view.value).toBeNull()
		wrapper.unmount()
	})

	it('shows a rule hint only for a refused rule', async () => {
		const {wrapper, preview} = mountPreview(ref(source()))
		await settle()
		requests[0].reject(new ApiError({status: 400, code: 2002, detail: 'invalid'}))
		await settle()

		expect(preview.state.value).toBe('ruleHint')
		expect(preview.view.value).toBeNull()
		expect(message.error).not.toHaveBeenCalled()
		wrapper.unmount()
	})

	it('hands a 401 to the app error handling and shows the preview as unavailable', async () => {
		const {wrapper, preview} = mountPreview(ref(source()))
		await settle()
		const unauthorized = new ApiError({status: 401, code: 11, detail: 'expired'})
		requests[0].reject(unauthorized)
		await settle()

		expect(preview.state.value).toBe('unavailable')
		expect(message.error).toHaveBeenCalledExactlyOnceWith(unauthorized)
		wrapper.unmount()
	})

	it('is unavailable offline instead of waiting for the network', async () => {
		onlineManager.setOnline(false)
		const {wrapper, preview} = mountPreview(ref(source()))
		await settle()

		expect(requests).toHaveLength(1)
		requests[0].reject(new TypeError('Failed to fetch'))
		await settle()

		expect(preview.state.value).toBe('unavailable')
		expect(message.error).not.toHaveBeenCalled()
		wrapper.unmount()
	})

	it('cancels the task\'s in-flight preview on request', async () => {
		const {wrapper, preview} = mountPreview(ref(source()))
		await settle()

		await preview.cancel()

		expect(requests[0].signal.aborted).toBe(true)
		wrapper.unmount()
	})
})
