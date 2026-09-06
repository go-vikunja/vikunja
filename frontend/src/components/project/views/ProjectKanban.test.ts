import {shallowMount} from '@vue/test-utils'
import {describe, expect, it, vi, beforeEach} from 'vitest'
import draggable from 'zhyswan-vuedraggable'

const updateBucket = vi.fn()

const buckets = [
	{id: 1, title: 'First', position: 100, tasks: [], count: 0, limit: 0},
	{id: 2, title: 'Second', position: 200, tasks: [], count: 0, limit: 0},
	{id: 3, title: 'Third', position: 300, tasks: [], count: 0, limit: 0},
]

vi.mock('@/stores/kanban', () => ({
	useKanbanStore: () => ({
		buckets,
		isLoading: false,
		updateBucket,
		getBucketById: (id: number) => buckets.find(b => b.id === id),
		setBucketById: vi.fn(),
		loadBucketsForProject: vi.fn(),
		loadNextTasksForBucket: vi.fn(),
	}),
}))

vi.mock('@/composables/useCurrentProject', () => ({
	useCurrentProject: () => ({
		currentProject: {
			value: {
				id: 1,
				title: 'Test',
				max_permission: 2,
				views: [{id: 10, view_kind: 'kanban', bucket_configuration_mode: 'manual'}],
			},
		},
		isPending: {value: false},
	}),
}))

vi.mock('@/composables/useTaskDragToProject', () => ({
	useTaskDragToProject: () => ({
		handleTaskDropToProject: async () => ({moved: false, targetProjectId: null}),
	}),
}))

vi.mock('@/stores/tasks', () => ({
	useTaskStore: () => ({isLoading: false, setDraggedTask: vi.fn()}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({settings: {frontendSettings: {alwaysShowBucketTaskCount: false}}}),
}))

vi.mock('@/services/savedFilter', () => ({
	isSavedFilter: () => false,
	useSavedFilter: () => ({filter: {value: null}}),
}))

vi.mock('@vueuse/router', () => ({
	useRouteQuery: () => ({value: undefined}),
}))

vi.mock('vue-router', () => ({
	useRouter: () => ({push: vi.fn(), currentRoute: {value: {fullPath: '/'}}}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectKanban from './ProjectKanban.vue'

function mountKanban() {
	return shallowMount(ProjectKanban, {
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
}

function dragEndEvent(bucketId: string) {
	const item = document.createElement('li')
	item.dataset.bucketId = bucketId
	// The DOM index sortable reports does not have to be a valid bucket index
	return {item, newIndex: 42}
}

describe('ProjectKanban', () => {
	beforeEach(() => updateBucket.mockClear())

	it('saves the position of the dropped bucket', () => {
		const wrapper = mountKanban()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('2'))

		expect(updateBucket).toHaveBeenCalledWith(expect.objectContaining({
			id: 2,
			position: 200,
		}))
	})

	it('does nothing when the dropped bucket is gone', () => {
		const wrapper = mountKanban()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('42'))

		expect(updateBucket).not.toHaveBeenCalled()
	})
})
