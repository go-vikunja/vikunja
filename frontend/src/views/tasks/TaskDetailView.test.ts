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
vi.mock('vuemoji-picker', () => ({VuemojiPicker: {render: () => null}}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import {mapTaskEverywhere} from '@/client/queries/taskCache'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import Heading from '@/components/tasks/partials/Heading.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import Datepicker from '@/components/input/Datepicker.vue'
import RepeatAfter from '@/components/tasks/partials/RepeatAfter.vue'
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

async function mountDetail(taskId = 1, visited: RouteLocationRaw[] = [], stubs: Record<string, unknown> = {}) {
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
		props: {taskId},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient}]],
			mocks: {$t: (key: string) => key},
			stubs: {Icon: true, XButton: true, Modal: true, CustomTransition: {template: '<div><slot/></div>'}, ...stubs},
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

	it('never shows the response of a save on another task opened meanwhile', async () => {
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task, path.task === 2
			? {due_date: '2025-05-05T10:00:00Z', percent_done: 0.5}
			: {})}))
		let respond: (value: unknown) => void = () => {}
		sdk.patchTasksRead.mockImplementation(() => new Promise(resolve => {
			respond = resolve
		}))
		const wrapper = await mountDetail()

		wrapper.findComponent(PercentDoneSelect).vm.$emit('update:modelValue', 0.9)
		await flushPromises()
		await wrapper.setProps({taskId: 2})
		await flushPromises()
		expect(dueDatePicker(wrapper).props('modelValue')).toEqual(new Date('2025-05-05T10:00:00Z'))

		respond({data: serverTask(1, {percent_done: 0.9, due_date: '2024-03-03T10:00:00Z'})})
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(dueDatePicker(wrapper).props('modelValue')).toEqual(new Date('2025-05-05T10:00:00Z'))
		expect(wrapper.findComponent(PercentDoneSelect).props('modelValue')).toBe(0.5)
	})
})

describe('TaskDetailView repeat', () => {
	const rule = {
		freq: 'weekly',
		interval: 2,
	}

	beforeEach(() => {
		sdk.tasksRead.mockReset()
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task, {
			repeat: rule,
			repeats_from_current_date: true,
		})}))
		sdk.patchTasksRead.mockReset()
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1)}))
	})

	function removeButton(wrapper: VueWrapper) {
		return wrapper.findAllComponents({name: 'BaseButton'})
			.find(button => button.attributes('aria-label') === 'task.detail.removeRepeat')
	}

	it('shows the editor for a task with a rule', async () => {
		const wrapper = await mountDetail()

		expect(wrapper.findComponent(RepeatAfter).props('modelValue')).toMatchObject({repeat: rule})
		expect(wrapper.findComponent(RepeatAfter).props('startOpen')).toBe(false)
	})

	it('mounts a fresh editor for each task', async () => {
		const wrapper = await mountDetail()
		const first = wrapper.findComponent(RepeatAfter).vm

		await wrapper.setProps({taskId: 2})
		await flushPromises()

		expect(wrapper.findComponent(RepeatAfter).props('modelValue')).toMatchObject({id: 2})
		expect(wrapper.findComponent(RepeatAfter).vm).not.toBe(first)
	})

	it('clears the rule with an explicit null and hides the editor', async () => {
		const wrapper = await mountDetail()

		removeButton(wrapper)!.vm.$emit('click')
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(sdk.patchTasksRead.mock.calls[0]![0].body).toEqual(expect.arrayContaining([
			{op: 'add', path: '/repeat', value: null},
			{op: 'add', path: '/repeats_from_current_date', value: false},
		]))
		expect(wrapper.findComponent(RepeatAfter).exists()).toBe(false)
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
		const wrapper = await mountDetail(1, ['/projects/5', {path: '/tasks/1', state: {back: '/projects/5'}}])

		pressShiftT()
		await flushPromises()
		await vi.waitFor(() => expect(wrapper.vm.$router.currentRoute.value.path).toBe('/projects/5'))

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(sdk.patchTasksRead.mock.calls[0]![0].body).toContainEqual({op: 'add', path: '/done', value: true})
	})

	it('falls back to the task project when opened directly', async () => {
		const wrapper = await mountDetail(1, ['/tasks/1'])

		pressShiftT()
		await flushPromises()

		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/projects/1')
	})

	it('stays on the task when saving fails', async () => {
		sdk.patchTasksRead.mockRejectedValue(new Error('nope'))
		const wrapper = await mountDetail(1, ['/tasks/1'])

		pressShiftT()
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(message.error).toHaveBeenCalled()
		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/tasks/1')
	})

	it('does nothing for a task that is already done', async () => {
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task, {done: true})}))
		const wrapper = await mountDetail(1, ['/tasks/1'])

		pressShiftT()
		await flushPromises()

		expect(sdk.patchTasksRead).not.toHaveBeenCalled()
		expect(wrapper.vm.$router.currentRoute.value.path).toBe('/tasks/1')
	})
})

describe('TaskDetailView empty repeat field', () => {
	beforeEach(() => {
		sdk.tasksRead.mockReset()
		sdk.tasksRead.mockImplementation(async ({path}: {path: {task: number}}) => ({data: serverTask(path.task)}))
		sdk.patchTasksRead.mockReset()
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1, {percent_done: 0.9})}))
	})

	async function setRepeat(wrapper: VueWrapper) {
		// XButton is registered globally in the app, so here it is a stub found by name.
		wrapper.findAllComponents({name: 'XButton'}).find(button => button.attributes('icon') === 'history')!.vm.$emit('click')
		await flushPromises()
	}

	function repeatOps() {
		return sdk.patchTasksRead.mock.calls
			.flatMap(([options]) => (options as {body: {path: string, value: unknown}[]}).body)
			.filter(op => op.path === '/repeat')
	}

	it('opens the editor that "Set Repeat" puts on screen', async () => {
		// startOpen only matters when the editor mounts, so record it there.
		const mountedOpen: boolean[] = []
		const RepeatAfterStub = {
			name: 'RepeatAfter',
			props: ['modelValue', 'disabled', 'startOpen'],
			setup(props: {startOpen: boolean}) {
				mountedOpen.push(props.startOpen)
				return () => null
			},
		}
		const wrapper = await mountDetail(1, [], {RepeatAfter: RepeatAfterStub})
		expect(mountedOpen).toEqual([])

		await setRepeat(wrapper)

		expect(mountedOpen).toEqual([true])
	})

	it('stays while the editor is open across a save of another field, which carries no repeat', async () => {
		const wrapper = await mountDetail()
		await setRepeat(wrapper)
		wrapper.findComponent(RepeatAfter).vm.$emit('update:open', true)

		wrapper.findComponent(PercentDoneSelect).vm.$emit('update:modelValue', 0.9)
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(repeatOps().every(op => op.value === null)).toBe(true)
		expect(wrapper.findComponent(RepeatAfter).exists()).toBe(true)
	})

	it('goes away on the next save once the editor closed without a repeat', async () => {
		const wrapper = await mountDetail()
		await setRepeat(wrapper)
		wrapper.findComponent(RepeatAfter).vm.$emit('update:open', true)
		wrapper.findComponent(RepeatAfter).vm.$emit('update:open', false)

		wrapper.findComponent(PercentDoneSelect).vm.$emit('update:modelValue', 0.9)
		await flushPromises()

		expect(wrapper.findComponent(RepeatAfter).exists()).toBe(false)
	})

	it('stays while the editor is open across a title save, which carries no repeat', async () => {
		sdk.patchTasksRead.mockImplementation(async () => ({data: serverTask(1, {title: 'Renamed'})}))
		const wrapper = await mountDetail(1, [], {Heading: false, BaseButton: false})
		await setRepeat(wrapper)
		wrapper.findComponent(RepeatAfter).vm.$emit('update:open', true)

		const title = wrapper.findComponent(Heading).find('h1')
		title.element.textContent = 'Renamed'
		await title.trigger('blur')
		await flushPromises()

		expect(sdk.patchTasksRead).toHaveBeenCalledOnce()
		expect(sdk.patchTasksRead.mock.calls[0]![0].body).toContainEqual({op: 'add', path: '/title', value: 'Renamed'})
		expect(repeatOps().every(op => op.value === null)).toBe(true)
		expect(wrapper.findComponent(RepeatAfter).exists()).toBe(true)
	})

	it('does not carry an open editor over to the next task', async () => {
		const wrapper = await mountDetail()
		await setRepeat(wrapper)
		wrapper.findComponent(RepeatAfter).vm.$emit('update:open', true)

		await wrapper.setProps({taskId: 2})
		await flushPromises()

		expect(wrapper.findComponent(RepeatAfter).exists()).toBe(false)
	})
})
