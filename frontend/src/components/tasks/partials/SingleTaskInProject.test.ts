import {describe, it, expect, vi, beforeEach} from 'vitest'
import {mount, shallowMount, flushPromises} from '@vue/test-utils'
import {createPinia, setActivePinia} from 'pinia'
import {createRouter, createMemoryHistory} from 'vue-router'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

import type {Task as ITask} from '@/client/generated'

const sdk = vi.hoisted(() => ({
	patchTasksRead: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/helpers/playPop', () => ({playPopSound: vi.fn()}))
vi.mock('@/composables/useProjects', () => ({useProjects: () => ({projects: {}})}))
vi.mock('@/composables/useCurrentProject', () => ({useCurrentProject: () => ({currentProject: {value: null}})}))
vi.mock('vue-i18n', async importOriginal => ({
	...(await importOriginal<typeof import('vue-i18n')>()),
	useI18n: () => ({t: (key: string) => key}),
}))

import SingleTaskInProject from './SingleTaskInProject.vue'

const theTask = {id: 7, project_id: 1, title: 'Water the plants', done: true, is_favorite: false} as ITask

function deferred() {
	let resolve!: (value: unknown) => void
	const promise = new Promise(done => resolve = done)
	return {promise, resolve}
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
		directives: {tooltip: {}},
		mocks: {$t: (key: string) => key},
		stubs: {TaskGlanceTooltip: true, CommentCount: true, ChecklistSummary: true, FancyCheckbox: true},
	}
}

function mountRow() {
	return shallowMount(SingleTaskInProject, {
		props: {theTask},
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
