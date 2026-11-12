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
