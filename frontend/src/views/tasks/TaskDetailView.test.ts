import {createPinia} from 'pinia'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {enableAutoUnmount, flushPromises, shallowMount, type VueWrapper} from '@vue/test-utils'
import {createMemoryHistory, createRouter, type RouteLocationRaw} from 'vue-router'

const sdk = vi.hoisted(() => ({
	tasksRead: vi.fn(),
	patchTasksRead: vi.fn(),
	tasksList: vi.fn(async () => ({data: {items: []}})),
	projectTasksList: vi.fn(async () => ({data: {items: []}})),
	projectViewTasksList: vi.fn(async () => ({data: {items: []}})),
	projectViewBucketsTasksList: vi.fn(async () => ({data: {items: []}})),
	bucketsList: vi.fn(async () => ({data: {items: []}})),
}))
vi.mock('@/client/generated', () => sdk)
const message = vi.hoisted(() => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/message', () => message)
const projects = vi.hoisted(() => ({} as Record<number, unknown>))
vi.mock('@/composables/useProjects', () => ({useProjects: () => ({projects})}))
vi.mock('@/composables/useLabels', () => ({useLabels: () => ({filterLabelsByQuery: () => [], isPending: {value: false}, getLabelById: vi.fn()})}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string, named?: Record<string, unknown>) => named ? `${key}:${JSON.stringify(named)}` : key}),
}))

vi.mock('@/i18n/lang/en.json', () => ({default: {}}))

import {ApiError} from '@/client/problem'
import BaseButton from '@/components/base/BaseButton.vue'
import ProjectInTrashNotice from '@/components/project/ProjectInTrashNotice.vue'
import TaskInTrashNotice from '@/components/tasks/partials/TaskInTrashNotice.vue'
import {mapTaskEverywhere} from '@/client/queries/taskCache'
import {taskKeys} from '@/client/queries/tasks'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import RelatedTasks from '@/components/tasks/partials/RelatedTasks.vue'
import TaskDetailView from './TaskDetailView.vue'

enableAutoUnmount(afterEach)

function serverTask(id: number, overrides: Record<string, unknown> = {}) {
	return {
		id,
		title: `Task ${id}`,
		description: '',
		done: false,
		project_id: 1,
		percent_done: 0.25,
		due_date: '2024-01-01T10:00:00Z',
		labels: [{id: 1, title: 'First'}],
		assignees: [],
		max_permission: 2,
		...overrides,
	}
}

let queryClient: QueryClient

function dueDatePicker(wrapper: VueWrapper) {
	return wrapper.findAllComponents(Datepicker)[0]!
}

async function mountDetail(taskId = 1, {visited = [], backdropView, stubs = {}}: {visited?: RouteLocationRaw[], backdropView?: string, stubs?: Record<string, unknown>} = {}) {
	queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [
			{path: '/', component: {render: () => null}},
			{path: '/tasks/:id', name: 'task.detail', component: {render: () => null}},
			{path: '/projects/:projectId', name: 'project.index', component: {render: () => null}},
			{path: '/:pathMatch(.*)*', name: 'not-found', component: {render: () => null}},
		],
	})
	await router.push('/')
	for (const location of visited) {
		await router.push(location)
	}
	const wrapper = shallowMount(TaskDetailView, {
		attachTo: document.body,
		props: {
			taskId,
			backdropView,
		},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient}]],
			mocks: {$t: (key: string) => key},
			stubs: {Icon: true, XButton: true, Modal: true, ...stubs},
			directives: {tooltip: () => {}, cy: () => {}, shortcut: () => {}},
		},
	})
	await flushPromises()
	return wrapper
}

describe('TaskDetailView draft', () => {
	beforeEach(() => {
		sdk.tasksRead.mockReset()
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task)}))
		sdk.patchTasksRead.mockReset()
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1)}))
	})

	it('keeps unsaved edits when the cached task changes and follows mutation-owned fields', async () => {
		const wrapper = await mountDetail()
		expect(dueDatePicker(wrapper).props('modelValue')).toEqual(new Date('2024-01-01T10:00:00Z'))

		dueDatePicker(wrapper).vm.$emit('update:modelValue', new Date('2024-02-02T10:00:00Z'))
		await flushPromises()

		mapTaskEverywhere(queryClient, 1, task => ({...task, is_unread: true, labels: [{id: 1, title: 'First'}, {id: 2, title: 'Second'}]}))
		await flushPromises()

		expect(sdk.patchTasksRead).not.toHaveBeenCalled()
		expect(dueDatePicker(wrapper).props('modelValue')).toEqual(new Date('2024-02-02T10:00:00Z'))
		expect(wrapper.findComponent(EditLabels).props('modelValue')).toEqual([{id: 1, title: 'First'}, {id: 2, title: 'Second'}])
	})

	it('reseeds the draft when the task id changes', async () => {
		const wrapper = await mountDetail()
		dueDatePicker(wrapper).vm.$emit('update:modelValue', new Date('2024-02-02T10:00:00Z'))
		await flushPromises()

		await wrapper.setProps({taskId: 2})
		await flushPromises()

		expect(dueDatePicker(wrapper).props('modelValue')).toEqual(new Date('2024-01-01T10:00:00Z'))
	})

	it('reseeds the draft from the response of a save', async () => {
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1, {percent_done: 0.9, due_date: '2024-03-03T10:00:00Z'})}))
		const wrapper = await mountDetail()

		wrapper.findComponent(PercentDoneSelect).vm.$emit('update:modelValue', 0.9)
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(wrapper.findAllComponents(Datepicker)[0]?.props('modelValue')).toEqual(new Date('2024-03-03T10:00:00Z'))
	})
})

describe('TaskDetailView mark done and go back', () => {
	beforeEach(() => {
		for (const id of Object.keys(projects)) {
			delete projects[Number(id)]
		}
		sdk.tasksRead.mockReset()
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task)}))
		sdk.patchTasksRead.mockReset()
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1, {done: true})}))
	})

	function pressShiftT() {
		document.body.dispatchEvent(new KeyboardEvent('keydown', {code: 'KeyT', key: 'T', shiftKey: true, bubbles: true}))
	}

	it('marks the task done and returns to the project list it was opened from', async () => {
		projects[5] = {id: 5, title: 'Origin'}
		const wrapper = await mountDetail(1, {visited: ['/projects/5', {path: '/tasks/1', state: {back: '/projects/5'}}]})

		pressShiftT()
		await flushPromises()
		await vi.waitFor(() => expect(wrapper.vm.$router.currentRoute.value.path).toBe('/projects/5'))

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(sdk.patchTasksRead.mock.calls[0]![0].body).toContainEqual({op: 'add', path: '/done', value: true})
	})

	it('falls back to the task project when opened directly', async () => {
		const wrapper = await mountDetail(1, {visited: ['/tasks/1']})

		pressShiftT()
		await flushPromises()

		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/projects/1')
	})

	it('stays on the task when saving fails', async () => {
		sdk.patchTasksRead.mockRejectedValue(new Error('nope'))
		const wrapper = await mountDetail(1, {visited: ['/tasks/1']})

		pressShiftT()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(message.error).toHaveBeenCalled()
		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/tasks/1')
	})

	it('does nothing for a task that is already done', async () => {
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task, {done: true})}))
		const wrapper = await mountDetail(1, {visited: ['/tasks/1']})

		pressShiftT()
		await flushPromises()

		expect(sdk.patchTasksRead).not.toHaveBeenCalled()
		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/tasks/1')
	})
})

describe('TaskDetailView in a trashed project', () => {
	it.each([
		{
			mode: 'standalone',
			backdropView: undefined,
		},
		{
			mode: 'modal',
			backdropView: '/projects/2',
		},
	])('shows the project notice in $mode mode without leaving the task route', async ({backdropView}) => {
		message.error.mockClear()
		const inTrash = {
			id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
			can_restore: true,
			can_restore_root: false,
		}
		sdk.tasksRead.mockRejectedValue(new ApiError({
			status: 404,
			code: 3018,
			detail: 'project in trash',
			details: inTrash,
		}))
		const wrapper = await mountDetail(1, {visited: ['/tasks/1'], backdropView})

		expect(message.error).not.toHaveBeenCalled()
		expect(wrapper.findComponent(ProjectInTrashNotice).props('inTrash')).toEqual(inTrash)
		expect(wrapper.findComponent(TaskInTrashNotice).exists()).toBe(false)
		expect(wrapper.classes()).not.toContain('is-loading')
		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/tasks/1')
		if (backdropView) {
			const close = wrapper.findComponent(BaseButton)
			expect(close.attributes('aria-label')).toBe('misc.close')
			close.vm.$emit('click')
			expect(wrapper.emitted('close')).toHaveLength(1)
		}
	})

	it('restores focus when restoring the project reveals a task still in the trash', async () => {
		const projectInTrash = {
			id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
			can_restore: true,
			can_restore_root: false,
		}
		const taskInTrash = {
			id: 1,
			project_id: 1,
			deleted_at: '2026-01-01T00:00:00Z',
			can_restore: true,
		}
		sdk.tasksRead.mockRejectedValueOnce(new ApiError({
			status: 404,
			code: 3018,
			detail: 'project in trash',
			details: projectInTrash,
		})).mockRejectedValue(new ApiError({
			status: 404,
			code: 4040,
			detail: 'task in trash',
			details: taskInTrash,
		}))
		const wrapper = await mountDetail(1, {visited: ['/tasks/1']})
		const projectNotice = wrapper.findComponent(ProjectInTrashNotice).element as HTMLElement
		projectNotice.tabIndex = -1
		projectNotice.focus()
		expect(document.activeElement).toBe(projectNotice)

		await queryClient.invalidateQueries({queryKey: taskKeys.details})
		await flushPromises()

		expect(wrapper.findComponent(TaskInTrashNotice).props('inTrash')).toEqual(taskInTrash)
		expect(document.activeElement).toBe(wrapper.element)
	})
})

describe('TaskDetailView property sidebar', () => {
	beforeEach(() => {
		sdk.tasksRead.mockReset()
		sdk.patchTasksRead.mockReset()
	})

	function serve(overrides: Record<string, unknown> = {}) {
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task, overrides)}))
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1, overrides)}))
	}

	const renderButtons = {BaseButton: {template: '<button><slot/></button>'}}

	it('keeps a section the user opened visible after saving another field', async () => {
		serve()
		const wrapper = await mountDetail()
		expect(wrapper.findComponent(RelatedTasks).exists()).toBe(false)

		document.body.dispatchEvent(new KeyboardEvent('keydown', {code: 'KeyR', key: 'r', bubbles: true}))
		await flushPromises()
		expect(wrapper.findComponent(RelatedTasks).exists()).toBe(true)

		wrapper.findComponent(PercentDoneSelect).vm.$emit('update:modelValue', 0.9)
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(wrapper.findComponent(RelatedTasks).exists()).toBe(true)
	})

	it('shows every property row to users who can edit the task', async () => {
		serve({due_date: '', labels: []})
		const wrapper = await mountDetail()

		expect(wrapper.find('.property-due-date').exists()).toBe(true)
		expect(wrapper.find('.property-labels').exists()).toBe(true)
		expect(wrapper.find('.property-due-date').classes()).not.toContain('is-set')
	})

	it('only shows rows with a value to read-only users', async () => {
		serve({max_permission: 0, labels: []})
		const wrapper = await mountDetail()

		expect(wrapper.find('.property-due-date').classes()).toContain('is-set')
		expect(wrapper.find('.property-percent-done').exists()).toBe(true)
		expect(wrapper.find('.property-labels').exists()).toBe(false)
		expect(wrapper.find('.property-priority').exists()).toBe(false)
		expect(wrapper.find('.sidebar-toolbar').exists()).toBe(false)
	})

	it.each([
		[{repeat_after: 0, repeat_mode: 0}, 'misc.notSet'],
		[{repeat_after: 172800, repeat_mode: 0}, 'task.repeat.summary.days:{"count":2}'],
		[{repeat_after: 0, repeat_mode: 1}, 'task.repeat.monthly'],
		[{repeat_after: 604800, repeat_mode: 2}, 'task.repeat.summary.fromCurrentDate:{"interval":"task.repeat.summary.weeks:{\\"count\\":1}"}'],
		[{repeat_after: 0, repeat_mode: 2}, 'task.repeat.fromCurrentDate'],
	])('summarizes repeat settings %o as %s', async (repeat, summary) => {
		serve(repeat)
		const wrapper = await mountDetail(1, {stubs: renderButtons})

		expect(wrapper.find('.property-repeat .property-value-button').text()).toBe(summary)
	})
})
