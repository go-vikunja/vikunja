import {createPinia} from 'pinia'
import {afterEach, beforeEach, describe, expect, it, vi} from 'vitest'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {enableAutoUnmount, flushPromises, shallowMount, type VueWrapper} from '@vue/test-utils'
import {createMemoryHistory, createRouter} from 'vue-router'

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
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
vi.mock('@/composables/useProjects', () => ({useProjects: () => ({projects: {}})}))
vi.mock('@/composables/useLabels', () => ({useLabels: () => ({filterLabelsByQuery: () => [], isPending: {value: false}, getLabelById: vi.fn()})}))
vi.mock('vuemoji-picker', () => ({VuemojiPicker: {render: () => null}}))
vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import {mapTaskEverywhere} from '@/client/queries/taskCache'
import EditLabels from '@/components/tasks/partials/EditLabels.vue'
import PercentDoneSelect from '@/components/tasks/partials/PercentDoneSelect.vue'
import Datepicker from '@/components/input/Datepicker.vue'
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

async function mountDetail(taskId = 1) {
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
	const wrapper = shallowMount(TaskDetailView, {
		props: {taskId},
		global: {
			plugins: [createPinia(), router, [VueQueryPlugin, {queryClient}]],
			mocks: {$t: (key: string) => key},
			stubs: {Icon: true, XButton: true, Modal: true, CustomTransition: {template: '<div><slot/></div>'}},
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
