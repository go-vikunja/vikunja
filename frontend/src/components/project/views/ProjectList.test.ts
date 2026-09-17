import {shallowMount, flushPromises} from '@vue/test-utils'
import {describe, expect, it, vi, beforeEach} from 'vitest'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
import {createRouter, createMemoryHistory} from 'vue-router'
import {createPinia} from 'pinia'
import draggable from 'zhyswan-vuedraggable'

import type {Task as ITask} from '@/client/generated'

const {updatePosition} = vi.hoisted(() => ({updatePosition: vi.fn()}))

vi.mock('@/client/generated', async importOriginal => ({...await importOriginal<object>(),
 tasksPositionUpdate: updatePosition,
 projectViewTasksList: vi.fn(async () => ({data: {items: [makeTask(1, 100), makeTask(2, 200), makeTask(3, 300)], page: 1, total_pages: 1}})),
}))
vi.mock('@/stores/auth', () => ({useAuthStore: () => ({settings: {timezone: 'UTC'}})}))
vi.mock('@/message', () => ({error: vi.fn()}))
vi.mock('@/composables/useTaskDragToProject', () => ({
	useTaskDragToProject: () => ({
		handleTaskDropToProject: async () => ({moved: false, targetProjectId: null}),
	}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({setHasTasks: vi.fn()}),
}))

vi.mock('@/composables/useCurrentProject', () => ({
	useCurrentProject: () => ({
		currentProject: {value: {id: 1, max_permission: 2}},
		isPending: {value: false},
	}),
}))

vi.mock('@/composables/useTaskActions', () => ({
	useTaskActions: () => ({setDraggedTask: vi.fn()}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectList from './ProjectList.vue'

function makeTask(id: number, position: number): ITask {
	return {id, title: `Task ${id}`, position} as ITask
}

async function mountList() {
	const router = createRouter({history: createMemoryHistory(), routes: [{path: '/', component: {render: () => null}}]})
 await router.push('/')
 const wrapper = shallowMount(ProjectList, {
		props: {
			isLoadingProject: false,
			projectId: 1,
			viewId: 10,
		},
		global: {
 plugins: [router, createPinia(), [VueQueryPlugin, {queryClient: new QueryClient({defaultOptions: {queries: {retry: false}}})}]],
			mocks: {$t: (key: string) => key},
			stubs: {
				ProjectWrapper: {template: '<div><slot name="default"/></div>'},
			},
		},
	})

	await flushPromises()

	return wrapper
}

function dragEndEvent(taskId: string, newIndex: number) {
	const item = document.createElement('li')
	item.dataset.taskId = taskId
	const list = document.createElement('ul')
	return {item, to: list, from: list, newIndex}
}

describe('ProjectList', () => {
	beforeEach(() => {
		updatePosition.mockReset()
		updatePosition.mockResolvedValue({data: {position: 200}})
	})

	it('saves the position of the dropped task', async () => {
		const wrapper = await mountList()

		// The DOM index sortable reports can point past the last task
		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('2', 42))
		await flushPromises()

		expect(updatePosition).toHaveBeenCalledWith(expect.objectContaining({
			path: {task: 2},
			body: {position: 200, project_view_id: 10},
		}))
	})

	it('does nothing when the dropped task is gone', async () => {
		const wrapper = await mountList()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('404', 0))
		await flushPromises()

		expect(updatePosition).not.toHaveBeenCalled()
	})
})
