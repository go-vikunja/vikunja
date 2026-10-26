import {describe, it, expect, vi, beforeEach, afterEach} from 'vitest'
import {shallowMount, flushPromises} from '@vue/test-utils'
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
const auth = vi.hoisted(() => ({settings: {timezone: 'UTC', week_start: 1}}))
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

function mountRow(row: ITask = theTask) {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}},
			{path: '/projects/:projectId', name: 'project.index', component: {render: () => null}},
		],
	})
	return shallowMount(SingleTaskInProject, {
		props: {theTask: row},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			directives: {tooltip},
			mocks: {$t: (key: string) => key},
		},
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
		auth.settings = {timezone: zone, week_start: 1}
		const row = {...theTask, done: false, repeat, due_date: FRIDAY_LATE_UTC} as ITask
		const listed = listTooltip(row)
		await flushPromises()
		expect(listed).toBe(`task.repeat.describe.${expected}`)
		expect(listed).toBe(detailSummary(row))
	})
})
