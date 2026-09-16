import {describe, expect, it, vi} from 'vitest'
import {reactive, ref, type Ref} from 'vue'
import type {RouteLocationNormalized} from 'vue-router'

vi.mock('vue-router', async importOriginal => ({
	...await importOriginal<typeof import('vue-router')>(),
	useRouter: () => ({
		resolve: (to: {params: {projectId: number, viewId: number}}) => ({
			fullPath: `/projects/${to.params.projectId}/${to.params.viewId}`,
		}),
		push: vi.fn(),
	}),
}))

vi.mock('@/stores/viewFilters', () => ({
	useViewFiltersStore: () => ({
		setViewQuery: vi.fn(),
		clearViewQuery: vi.fn(),
	}),
}))

vi.mock('./useGanttTaskList', () => ({
	useGanttTaskList: () => ({
		tasks: ref(new Map()),
		isLoading: ref(false),
		loadTasks: vi.fn(),
		addTask: vi.fn(),
		updateTask: vi.fn(),
	}),
}))

import {useGanttFilters} from './useGanttFilters'

function routeRef(route: Partial<RouteLocationNormalized>): Ref<RouteLocationNormalized> {
	return ref(reactive(route)) as unknown as Ref<RouteLocationNormalized>
}

describe('useGanttFilters', () => {
	// The gantt view renders as the backdrop of the task detail modal, where the current route is
	// the task and has no project params.
	it('takes project and view id from the props, not from the route', () => {
		const route = routeRef({
			name: 'task.detail',
			params: {id: '3619'},
			query: {},
			fullPath: '/tasks/3619',
		})

		const {filters} = useGanttFilters(route, ref(5), ref(2759))

		expect(filters.value.projectId).toBe(5)
		expect(filters.value.viewId).toBe(2759)
	})

	it('reads the date range from the route query', () => {
		const route = routeRef({
			name: 'project.view',
			params: {projectId: '5', viewId: '2759'},
			query: {dateFrom: '2024-01-01', dateTo: '2024-02-01'},
			fullPath: '/projects/5/2759?dateFrom=2024-01-01&dateTo=2024-02-01',
		})

		const {filters} = useGanttFilters(route, ref(5), ref(2759))

		expect(filters.value.dateFrom).toBe(new Date(2024, 0, 1).toISOString())
		expect(filters.value.dateTo).toBe(new Date(2024, 1, 1).toISOString())
	})
})
