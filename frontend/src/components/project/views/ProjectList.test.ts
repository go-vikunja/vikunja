import {shallowMount, flushPromises} from '@vue/test-utils'
import {describe, expect, it, vi, beforeEach} from 'vitest'
import {nextTick, ref} from 'vue'
import draggable from 'zhyswan-vuedraggable'

import type {ITask} from '@/modelTypes/ITask'

const {updatePosition} = vi.hoisted(() => ({updatePosition: vi.fn()}))

const allTasks = ref<ITask[]>([])

vi.mock('@/services/taskPosition', () => ({
	default: class {
		update = updatePosition
	},
}))

vi.mock('@/composables/useTaskList', () => ({
	useTaskList: () => ({
		tasks: allTasks,
		loading: ref(false),
		totalPages: ref(1),
		currentPage: ref(1),
		loadTasks: vi.fn(),
		params: ref({}),
		sortByParam: ref({position: 'asc'}),
	}),
}))

vi.mock('@/composables/useTaskDragToProject', () => ({
	useTaskDragToProject: () => ({
		handleTaskDropToProject: async () => ({moved: false, targetProjectId: null}),
	}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({
		currentProject: {id: 1, maxPermission: 2},
		setHasTasks: vi.fn(),
	}),
}))

vi.mock('@/stores/tasks', () => ({
	useTaskStore: () => ({setDraggedTask: vi.fn()}),
}))

vi.mock('@/services/savedFilter', () => ({
	isSavedFilter: () => false,
	useSavedFilter: () => ({filter: ref(null)}),
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
	const wrapper = shallowMount(ProjectList, {
		props: {
			isLoadingProject: false,
			projectId: 1,
			viewId: 10,
		},
		global: {
			mocks: {$t: (key: string) => key},
			stubs: {
				ProjectWrapper: {template: '<div><slot name="default"/></div>'},
			},
		},
	})

	allTasks.value = [makeTask(1, 100), makeTask(2, 200), makeTask(3, 300)]
	await nextTick()

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
		allTasks.value = []
		updatePosition.mockReset()
		updatePosition.mockResolvedValue(undefined)
	})

	it('saves the position of the dropped task', async () => {
		const wrapper = await mountList()

		// The DOM index sortable reports can point past the last task
		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('2', 42))
		await flushPromises()

		expect(updatePosition).toHaveBeenCalledWith(expect.objectContaining({
			taskId: 2,
			position: 200,
		}))
	})

	it('does nothing when the dropped task is gone', async () => {
		const wrapper = await mountList()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('404', 0))
		await flushPromises()

		expect(updatePosition).not.toHaveBeenCalled()
	})
})
