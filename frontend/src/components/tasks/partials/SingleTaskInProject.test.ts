import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {mount, shallowMount, flushPromises} from '@vue/test-utils'
import type {Directive} from 'vue'
import {createPinia, setActivePinia} from 'pinia'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {Task as ITask, TaskRepeat} from '@/client/generated'
import {dueDateOf, recurrenceZone} from '@/helpers/rrule'
import {describeRepeatRule} from '@/helpers/rruleDescribe'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/playPop', () => ({playPopSound: vi.fn()}))
vi.mock('@/composables/useProjects', () => ({useProjects: () => ({projects: {}})}))
vi.mock('@/composables/useCurrentProject', () => ({useCurrentProject: () => ({currentProject: {value: null}})}))
// frontend_settings is read by PriorityLabel when the subtask tests fully mount the row.
const auth = vi.hoisted(() => ({settings: {timezone: 'UTC', week_start: 1, frontend_settings: {}}}))
vi.mock('@/stores/auth', async importOriginal => ({
	...(await importOriginal<typeof import('@/stores/auth')>()),
	useAuthStore: () => auth,
}))

// Shows the key, its params and the plural count, so the tooltip can be compared exactly.
const t = vi.hoisted(() => (key: string, params?: Record<string, unknown>, plural?: number) =>
	`${key}|${JSON.stringify(params ?? {})}|${plural ?? ''}`)
vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t, locale: {value: 'en'}}),
}))

import SingleTaskInProject from './SingleTaskInProject.vue'

const theTask = {id: 7, project_id: 1, title: 'Water the plants', done: true, is_favorite: false} as ITask

function deferred() {
	let resolve!: (value: unknown) => void
	const promise = new Promise(done => resolve = done)
	return {promise, resolve}
}

// Records the tooltip text where the test can read it, instead of rendering a floating tooltip.
const tooltip: Directive<HTMLElement, string> = {
	mounted: (el, {value}) => { el.dataset.tooltip = value ?? '' },
	updated: (el, {value}) => { el.dataset.tooltip = value ?? '' },
}

function createTestRouter() {
	return createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}},
			{path: '/projects/:projectId', name: 'project.index', component: {render: () => null}},
		],
	})
}

function globalOptions() {
	return {
		plugins: [createPinia(), createTestRouter(), [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]] as never[],
		directives: {tooltip},
		mocks: {$t: (key: string) => key},
		stubs: {TaskGlanceTooltip: true, CommentCount: true, ChecklistSummary: true, FancyCheckbox: true},
	}
}

function mountRow(row: ITask = theTask) {
	return shallowMount(SingleTaskInProject, {
		props: {theTask: row},
		global: globalOptions(),
	})
}

describe('SingleTaskInProject', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		sdk.patchTasksRead.mockReset()
	})

	it('shows the row as loading while marking it undone is in flight', async () => {
		const update = deferred()
		sdk.patchTasksRead.mockReturnValue(update.promise)
		const wrapper = mountRow()

		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', false)
		await flushPromises()

		await vi.waitFor(() => expect(wrapper.get('.task').classes()).toContain('is-loading'))

		update.resolve({data: {...theTask, done: false}})
		await flushPromises()

		expect(wrapper.get('.task').classes()).not.toContain('is-loading')
	})

	it('shows the row as loading while favouring it is in flight', async () => {
		const favorite = deferred()
		sdk.patchTasksRead.mockReturnValue(favorite.promise)
		const wrapper = mountRow()

		await wrapper.get('.favorite').trigger('click')
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledWith(expect.objectContaining({
			path: {task: 7},
			body: [expect.objectContaining({path: '/is_favorite', value: true})],
		}))
		await vi.waitFor(() => expect(wrapper.get('.task').classes()).toContain('is-loading'))

		favorite.resolve({data: {...theTask, is_favorite: true}})
		await flushPromises()

		expect(wrapper.get('.task').classes()).not.toContain('is-loading')
	})

	describe('subtask rendering', () => {
		function task(id: number, subtaskIds: number[] = []): ITask {
			return {
				id,
				project_id: 1,
				title: `Task ${id}`,
				done: false,
				is_favorite: false,
				related_tasks: {subtask: subtaskIds.map(subId => ({id: subId}))},
			} as ITask
		}

		function mountTree(tasks: ITask[]) {
			return mount(SingleTaskInProject, {
				props: {theTask: tasks[0], allTasks: tasks},
				global: globalOptions(),
			})
		}

		function renderedTaskIds(wrapper: ReturnType<typeof mountTree>) {
			return wrapper.findAll('[data-task-id]').map(el => Number(el.attributes('data-task-id')))
		}

		it('renders each task of a diamond graph a bounded number of times', () => {
			// Every task of a level is a subtask of every task of the level above.
			const levels = 4
			const width = 4
			const ids = Array.from({length: levels}, (_, level) =>
				Array.from({length: width}, (_, i) => 2 + level * width + i))
			const tasks = [
				task(1, ids[0]),
				...ids.flatMap((levelIds, level) => levelIds.map(id => task(id, ids[level + 1] ?? []))),
			]

			const rendered = renderedTaskIds(mountTree(tasks))

			expect(new Set(rendered)).toEqual(new Set(tasks.map(t => t.id)))
			const edgeCount = tasks.reduce((sum, t) => sum + (t.related_tasks?.subtask?.length ?? 0), 0)
			expect(rendered.length).toBeLessThanOrEqual(edgeCount + 1)
		})

		it('renders a plain tree completely', () => {
			const tasks = [
				task(1, [2, 3]),
				task(2, [4, 5]),
				task(3, [6]),
				task(4),
				task(5),
				task(6),
			]

			expect(renderedTaskIds(mountTree(tasks))).toEqual([1, 2, 4, 5, 3, 6])
		})

		it('stops nesting subtasks after five levels', () => {
			const tasks = Array.from({length: 10}, (_, i) => task(i + 1, i < 9 ? [i + 2] : []))

			expect(renderedTaskIds(mountTree(tasks))).toEqual([1, 2, 3, 4, 5, 6])
		})
	})
})

describe('repeat tooltip in the task list', () => {
	const originalTZ = process.env.TZ
	// Fri Oct 20 2090, 23:30 UTC: already Saturday in Tokyo.
	const FRIDAY_LATE_UTC = '2090-10-20T23:30:00Z'

	beforeEach(() => {
		setActivePinia(createPinia())
		process.env.TZ = 'UTC'
	})

	afterEach(() => {
		process.env.TZ = originalTZ
	})

	function listTooltip(row: ITask): string {
		const wrapper = mountRow(row)
		const icon = wrapper.findAll('.project-task-icon').find(span => span.html().includes('history'))
		expect(icon, 'repeat icon').toBeDefined()
		return (icon!.element as HTMLElement).dataset.tooltip ?? ''
	}

	function detailSummary(row: ITask): string {
		return describeRepeatRule(row.repeat, {
			t,
			locale: 'en',
			weekStart: auth.settings.week_start,
			zone: recurrenceZone(auth.settings),
			dueDate: dueDateOf(row),
		})
	}

	it.each([
		['a weekday rule', {freq: 'weekly', by_day: ['mo', 'we', 'fr']}, 'UTC', 'weeklyOn|{"n":1,"days":"Monday, Wednesday, and Friday"}|1'],
		['an implicit weekday, read in the settings zone', {freq: 'weekly'}, 'UTC', 'weeklyOn|{"n":1,"days":"Friday"}|1'],
		['an implicit weekday that is a day later in the settings zone', {freq: 'weekly'}, 'Asia/Tokyo', 'weeklyOn|{"n":1,"days":"Saturday"}|1'],
		['a rule the server marks read-only', {freq: 'yearly', read_only: true}, 'UTC', 'locked|{}|'],
	] as [string, TaskRepeat, string, string][])('matches the task detail summary for %s', async (_name, repeat, zone, expected) => {
		auth.settings = {timezone: zone, week_start: 1, frontend_settings: {}}
		const row = {...theTask, done: false, repeat, due_date: FRIDAY_LATE_UTC} as ITask
		const listed = listTooltip(row)
		await flushPromises()
		expect(listed).toBe(`task.repeat.describe.${expected}`)
		expect(listed).toBe(detailSummary(row))
	})
})

// One overdue repeating row carries both sides of the upstream merge: upstream's overdue icon and
// dash-free due date (627ff3df5, e730665ee) and the RRULE repeat icon with its rule summary.
describe('an overdue repeating task in the list', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		auth.settings = {timezone: 'UTC', week_start: 1, frontend_settings: {}}
	})

	it('shows the overdue icon, the date without a dash and the repeat summary', async () => {
		const row = {
			...theTask,
			done: false,
			due_date: '2020-01-03T09:00:00Z',
			repeat: {freq: 'weekly', by_day: ['fr']},
		} as ITask
		const wrapper = mount(SingleTaskInProject, {
			props: {theTask: row},
			global: globalOptions(),
		})
		await flushPromises()

		expect(wrapper.get('.task').attributes('data-is-overdue')).toBe('true')
		expect(wrapper.find('.dueDate .overdue-indicator').exists(), 'overdue icon').toBe(true)
		const due = wrapper.get('.dueDate time').text()
		expect(due).toBe('task.detail.due')
		expect(due).not.toMatch(/^[–-]/)

		const repeatIcon = wrapper.findAll('.project-task-icon').find(span => span.html().includes('history'))
		expect(repeatIcon, 'repeat icon').toBeDefined()
		expect((repeatIcon!.element as HTMLElement).dataset.tooltip).toBe('task.repeat.describe.weeklyOn|{"n":1,"days":"Friday"}|1')
	})
})

// Undo after checking off a task restores the task as it was before the check: open, with the same
// due date, rule and remaining count. The server keeps a stored COUNT when a save raises it on an
// otherwise unchanged rule, so a lowered count is restored by clearing the rule and then setting it.
describe('undo after completing a repeating task', () => {
	const DUE = '2026-01-01T12:00:00Z'
	const NEXT_DAY = '2026-01-02T12:00:00Z'

	function repeatOps(call: unknown[]) {
		const {body} = call[0] as {body: {path: string, value: unknown}[]}
		return {
			repeat: body.find(op => op.path === '/repeat'),
			done: body.find(op => op.path === '/done')?.value,
			due: body.find(op => op.path === '/due_date')?.value,
		}
	}

	async function checkThenUndo(row: ITask, completed: ITask, restoreResponses: Array<ITask | Error>) {
		const {success} = await import('@/message')
		vi.mocked(success).mockClear()
		sdk.patchTasksRead.mockResolvedValueOnce({data: completed})
		for (const response of restoreResponses) {
			if (response instanceof Error) sdk.patchTasksRead.mockRejectedValueOnce(response)
			else sdk.patchTasksRead.mockResolvedValueOnce({data: response})
		}
		const wrapper = mountRow(row)
		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		await wrapper.setProps({theTask: completed})
		const actions = vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[] | undefined
		expect(actions?.[0], 'undo action').toBeDefined()
		actions![0].callback()
		await flushPromises()
		await vi.advanceTimersByTimeAsync(0)
		await flushPromises()
		return wrapper
	}

	beforeEach(() => {
		setActivePinia(createPinia())
		auth.settings = {timezone: 'UTC', week_start: 1, frontend_settings: {}}
		sdk.patchTasksRead.mockReset()
		vi.useFakeTimers({shouldAdvanceTime: true})
	})

	afterEach(() => {
		vi.useRealTimers()
	})

	it('restores the rule of a task whose last occurrence was completed (ends after 1)', async () => {
		const repeat = {freq: 'daily', interval: 1, count: 1} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: true, due_date: DUE} as ITask
		await checkThenUndo(row, completed, [{...row}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		const undo = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(undo.done).toBe(false)
		expect(undo.due).toBe(DUE)
		expect(undo.repeat?.value).toEqual(expect.objectContaining({freq: 'daily', count: 1}))
	})

	it('clears and then sets the rule to restore a lowered count (ends after 2)', async () => {
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		const cleared = {...theTask, done: false, due_date: DUE} as ITask
		await checkThenUndo(row, completed, [cleared, {...row}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(3)
		const clear = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(clear.repeat).toEqual(expect.objectContaining({value: null}))
		expect(clear.done).toBe(false)
		expect(clear.due).toBe(DUE)
		const restore = repeatOps(sdk.patchTasksRead.mock.calls[2])
		expect(restore.repeat?.value).toEqual(expect.objectContaining({freq: 'daily', count: 2}))
		expect(restore.done).toBe(false)
		expect(restore.due).toBe(DUE)
	})

	it('sends no second save when clearing the rule fails, and shows the error', async () => {
		const {error} = await import('@/message')
		vi.mocked(error).mockClear()
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		await checkThenUndo(row, completed, [new Error('offline')])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		expect(vi.mocked(error)).toHaveBeenCalled()
	})

	it('shows the error when setting the rule again fails', async () => {
		const {error} = await import('@/message')
		vi.mocked(error).mockClear()
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		await checkThenUndo(row, completed, [{...theTask, done: false, due_date: DUE} as ITask, new Error('offline')])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(3)
		expect(vi.mocked(error)).toHaveBeenCalled()
	})

	it('starts setting the rule only after clearing it has finished, judged by the completion response', async () => {
		const {success} = await import('@/message')
		vi.mocked(success).mockClear()
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		const clear = deferred()
		sdk.patchTasksRead.mockResolvedValueOnce({data: completed})
		sdk.patchTasksRead.mockReturnValueOnce(clear.promise)
		sdk.patchTasksRead.mockResolvedValueOnce({data: {...row}})
		// The row is never replaced with the completion response: the branch must come from the response.
		const wrapper = mountRow(row)
		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		const actions = vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[]
		actions[0].callback()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		expect(repeatOps(sdk.patchTasksRead.mock.calls[1]).repeat).toEqual(expect.objectContaining({value: null}))

		clear.resolve({data: {...theTask, done: false, due_date: DUE}})
		await flushPromises()
		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(3)
		expect(repeatOps(sdk.patchTasksRead.mock.calls[2]).repeat?.value).toEqual(expect.objectContaining({count: 2}))
	})

	it('undoing an earlier check restores that check, not a later one', async () => {
		const {success} = await import('@/message')
		vi.mocked(success).mockClear()
		const repeat = {freq: 'daily', interval: 1, count: 3} as TaskRepeat
		const first = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const second = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 2}} as ITask
		const third = {...theTask, done: false, due_date: '2026-01-03T12:00:00Z', repeat: {...repeat, count: 1}} as ITask
		sdk.patchTasksRead.mockResolvedValueOnce({data: second})
		sdk.patchTasksRead.mockResolvedValueOnce({data: third})
		sdk.patchTasksRead.mockResolvedValue({data: {...first}})
		const wrapper = mountRow(first)
		const checkbox = () => wrapper.getComponent({name: 'FancyCheckbox'}).vm
		checkbox().$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		const firstUndo = (vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[])[0]
		await wrapper.setProps({theTask: second})
		checkbox().$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		await wrapper.setProps({theTask: third})

		firstUndo.callback()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(4)
		const clear = repeatOps(sdk.patchTasksRead.mock.calls[2])
		expect(clear.repeat).toEqual(expect.objectContaining({value: null}))
		expect(clear.due).toBe(DUE)
		const restore = repeatOps(sdk.patchTasksRead.mock.calls[3])
		expect(restore.due).toBe(DUE)
		expect(restore.repeat?.value).toEqual(expect.objectContaining({count: 3}))
	})

	it('never rewrites a read-only repeat: one save, no clear, the lowered count stays', async () => {
		const repeat = {freq: 'yearly', interval: 1, count: 3, read_only: true} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 2}} as ITask
		await checkThenUndo(row, completed, [{...completed}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		const undo = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(undo.repeat).toBeUndefined()
		expect(undo.due).toBe(DUE)
	})

	it('does not write a read-only repeat back after its last occurrence, even from a stale row', async () => {
		const {success} = await import('@/message')
		vi.mocked(success).mockClear()
		const repeat = {freq: 'yearly', interval: 1, count: 1, read_only: true} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		sdk.patchTasksRead.mockResolvedValueOnce({data: {...theTask, done: true, due_date: DUE}})
		sdk.patchTasksRead.mockResolvedValueOnce({data: {...theTask, done: false, due_date: DUE}})
		// The row keeps showing the repeat: it is never replaced with the completion response.
		const wrapper = mountRow(row)
		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		const actions = vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[]
		actions[0].callback()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		const undo = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(undo.repeat).toBeUndefined()
		expect(undo.done).toBe(false)
	})

	it('does not write a read-only repeat back after its last occurrence', async () => {
		const repeat = {freq: 'yearly', interval: 1, count: 1, read_only: true} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: true, due_date: DUE} as ITask
		await checkThenUndo(row, completed, [{...completed, done: false}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		expect(repeatOps(sdk.patchTasksRead.mock.calls[1]).repeat).toBeUndefined()
	})

	it('a second Undo restores the same state again', async () => {
		const {success} = await import('@/message')
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		sdk.patchTasksRead.mockResolvedValue({data: {...row}})
		await checkThenUndo(row, completed, [])
		const actions = vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[]
		actions[0].callback()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(5)
		for (const call of sdk.patchTasksRead.mock.calls.slice(1)) {
			expect(repeatOps(call).due).toBe(DUE)
		}
		expect(repeatOps(sdk.patchTasksRead.mock.calls[4]).repeat?.value).toEqual(expect.objectContaining({count: 2}))
	})

	it('finishes restoring the rule when the row goes away while clearing it', async () => {
		const {success} = await import('@/message')
		vi.mocked(success).mockClear()
		const repeat = {freq: 'daily', interval: 1, count: 2} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat: {...repeat, count: 1}} as ITask
		const clear = deferred()
		sdk.patchTasksRead.mockResolvedValueOnce({data: completed})
		sdk.patchTasksRead.mockReturnValueOnce(clear.promise)
		sdk.patchTasksRead.mockResolvedValueOnce({data: {...row}})
		const wrapper = mountRow(row)
		wrapper.getComponent({name: 'FancyCheckbox'}).vm.$emit('update:modelValue', true)
		await flushPromises()
		await vi.advanceTimersByTimeAsync(300)
		await flushPromises()
		const actions = vi.mocked(success).mock.calls.at(-1)?.[1] as {callback: () => void}[]
		actions[0].callback()
		await flushPromises()
		wrapper.unmount()

		clear.resolve({data: {...theTask, done: false, due_date: DUE}})
		await flushPromises()
		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(3)
		expect(repeatOps(sdk.patchTasksRead.mock.calls[2]).repeat?.value).toEqual(expect.objectContaining({count: 2}))
	})

	it('restores a never-ending repeat with one save, as before', async () => {
		const repeat = {freq: 'daily', interval: 1} as TaskRepeat
		const row = {...theTask, done: false, due_date: DUE, repeat} as ITask
		const completed = {...theTask, done: false, due_date: NEXT_DAY, repeat} as ITask
		await checkThenUndo(row, completed, [{...row}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		const undo = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(undo.due).toBe(DUE)
		expect(undo.repeat?.value).toEqual(expect.objectContaining({freq: 'daily'}))
	})

	it('reopens a non-repeating task with one save, as before', async () => {
		const row = {...theTask, done: false, due_date: DUE} as ITask
		const completed = {...theTask, done: true, due_date: DUE} as ITask
		await checkThenUndo(row, completed, [{...row}])

		expect(sdk.patchTasksRead).toHaveBeenCalledTimes(2)
		const undo = repeatOps(sdk.patchTasksRead.mock.calls[1])
		expect(undo.done).toBe(false)
		expect(undo.repeat).toBeUndefined()
	})
})
