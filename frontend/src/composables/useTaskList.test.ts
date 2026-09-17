import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {defineComponent, h, nextTick} from 'vue'
import {mount, flushPromises, enableAutoUnmount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createRouter, createMemoryHistory, RouterView, type Router} from 'vue-router'

import type {Task as ITask} from '@/client/generated'
import {createTaskDraft} from '@/helpers/task'
import {normalizeTask} from '@/client/queries/tasks'

import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

type TaskListRequest = {
	path: {project: number, view: number},
	query: Record<string, unknown> & {page: number},
}

const sdk = vi.hoisted(() => ({
	projectViewTasksList: vi.fn(),
}))

vi.mock('@/client/generated', () => sdk)

function taskListResponse(items: ITask[], page: number) {
	return {
		data: {
			items,
			page,
			total_pages: 5,
		},
	}
}

vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
import {error} from '@/message'
const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})

beforeEach(() => {
	queryClient.clear()
	localStorage.clear()
	setActivePinia(createPinia())
	vi.mocked(error).mockClear()
	sdk.projectViewTasksList.mockReset()
	sdk.projectViewTasksList.mockImplementation(
		async ({query}: TaskListRequest) => taskListResponse([], query.page),
	)
})

import {useTaskList, buildStoredQuery} from './useTaskList'
import {useViewFiltersStore} from '@/stores/viewFilters'

enableAutoUnmount(afterEach)

describe('buildStoredQuery', () => {
	it('includes sort when set', () => {
		expect(buildStoredQuery({sort: 'due_date:asc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'due_date:asc'})
	})

	it('includes filter and search when set', () => {
		expect(buildStoredQuery({sort: undefined, filter: 'done = false', s: 'foo', page: 1}))
			.toEqual({filter: 'done = false', s: 'foo'})
	})

	it('omits page when it equals the default of 1', () => {
		expect(buildStoredQuery({sort: 'id:desc', filter: undefined, s: undefined, page: 1}))
			.toEqual({sort: 'id:desc'})
	})

	it('includes page when greater than 1', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 3}))
			.toEqual({page: '3'})
	})

	it('returns an empty object when nothing is set', () => {
		expect(buildStoredQuery({sort: undefined, filter: undefined, s: undefined, page: 1}))
			.toEqual({})
	})

	it('skips empty strings', () => {
		expect(buildStoredQuery({sort: '', filter: '', s: '', page: 1}))
			.toEqual({})
	})
})

function lastQuery(): Record<string, unknown> {
	return (sdk.projectViewTasksList.mock.lastCall?.[0] as TaskListRequest).query
}

function listRequests() {
	return sdk.projectViewTasksList.mock.calls.map(([request]) => ({
		path: (request as TaskListRequest).path,
		query: (request as TaskListRequest).query,
	}))
}

async function mountTaskList(query: Record<string, string>): Promise<Router> {
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{path: '/', name: 'home', component: {render: () => null}}],
	})
	await router.push({path: '/', query})
	await router.isReady()

	const TestComponent = defineComponent({
		setup() {
			useTaskList(() => 1, () => 1)
			return () => h('div')
		},
	})

	mount(TestComponent, {global: {plugins: [router, [VueQueryPlugin, {queryClient}]]}})
	await flushPromises()
	await nextTick()
	return router
}

describe('useTaskList sort handling for relevance ranking', () => {
	it('omits the sort while searching with the default sort so the backend ranks by relevance', async () => {
		await mountTaskList({s: 'find me'})

		const query = lastQuery()
		expect(query.q).toBe('find me')
		expect(query.sort_by).toEqual([])
		expect(query.order_by).toEqual([])
	})

	it('keeps an explicit user sort while searching so the user sort is respected', async () => {
		await mountTaskList({s: 'find me', sort: 'title:asc'})

		const query = lastQuery()
		expect(query.q).toBe('find me')
		expect(query.sort_by).toEqual(['title'])
		expect(query.order_by).toEqual(['asc'])
	})

	it.each(['abc', '0'])('loads the first page when the url asks for page %s', async (page) => {
		await mountTaskList({page})

		expect(lastQuery().page).toBe(1)
	})

	it('sends the default sort when not searching', async () => {
		await mountTaskList({})

		const query = lastQuery()
		expect(query.q).toBe('')
		// id always sorts last so other sort columns take precedence.
		expect(query.sort_by).toEqual(['id'])
		expect(query.order_by).toEqual(['desc'])
	})
})

describe('useTaskList error reporting', () => {
	it('toasts a failing load instead of showing an empty list', async () => {
		sdk.projectViewTasksList.mockRejectedValueOnce(new Error('Server is on fire'))

		await mountTaskList({})
		await flushPromises()

		expect(error).toHaveBeenCalledWith(expect.objectContaining({message: 'Server is on fire'}))
	})

	it('stays silent when the request was aborted', async () => {
		sdk.projectViewTasksList.mockRejectedValueOnce(new DOMException('aborted', 'AbortError'))

		await mountTaskList({})
		await flushPromises()

		expect(error).not.toHaveBeenCalled()
	})
})

async function mountRoutedTaskList() {
	let taskList: ReturnType<typeof useTaskList>
	const List = defineComponent({
		props: {projectId: {type: Number, required: true}, viewId: {type: Number, required: true}},
		setup(props) {
			taskList = useTaskList(() => props.projectId, () => props.viewId, {position: 'asc'})
			return () => h('div')
		},
	})
	const View = defineComponent({
		props: {projectId: Number, viewId: Number},
		setup: props => () => h(List, {projectId: props.projectId!, viewId: props.viewId!}),
	})
	const router = createRouter({
		history: createMemoryHistory(),
		routes: [{
			path: '/projects/:projectId/:viewId',
			component: View,
			props: route => ({projectId: Number(route.params.projectId), viewId: Number(route.params.viewId)}),
		}],
	})
	await router.push('/projects/1/11')
	mount(defineComponent({render: () => h(RouterView)}), {global: {plugins: [router, [VueQueryPlugin, {queryClient}]]}})
	await flushPromises()
	return {router, taskList: taskList!}
}

describe('useTaskList navigation and pagination', () => {
	it.each([1, 3])('restores the sort and page %i when returning to a project', async (page) => {
		const {router, taskList} = await mountRoutedTaskList()
		taskList.sortByParam.value = {due_date: 'asc'}
		await flushPromises()
		taskList.currentPage.value = page
		await flushPromises()

		sdk.projectViewTasksList.mockClear()
		await router.push('/projects/2/21')
		await flushPromises()
		expect(listRequests()).toEqual([{
			path: {project: 2, view: 21},
			query: expect.objectContaining({sort_by: ['position'], order_by: ['asc'], page: 1}),
		}])
		expect(taskList.sortByParam.value).toEqual({position: 'asc'})

		sdk.projectViewTasksList.mockClear()
		await router.push('/projects/1/11')
		await flushPromises()
		expect(listRequests()).toEqual([{
			path: {project: 1, view: 11},
			query: expect.objectContaining({sort_by: ['due_date'], order_by: ['asc'], page}),
		}])
		expect(router.currentRoute.value.query.sort).toBe('due_date:asc')
		expect(taskList.sortByParam.value).toEqual({due_date: 'asc'})
		expect(taskList.currentPage.value).toBe(page)
	})

	it('loads a restored saved query once and never with the pre-restore params', async () => {
		const {router} = await mountRoutedTaskList()
		useViewFiltersStore().setViewQuery(21, {sort: 'due_date:asc', page: '2'})

		sdk.projectViewTasksList.mockClear()
		await router.push('/projects/2/21')
		await flushPromises()

		expect(listRequests()).toEqual([{
			path: {project: 2, view: 21},
			query: expect.objectContaining({sort_by: ['due_date'], order_by: ['asc'], page: 2}),
		}])
	})

	it('ignores a response from a project the user has left', async () => {
		const {router, taskList} = await mountRoutedTaskList()
		let resolveOld!: (response: unknown) => void
		let resolveCurrent!: (response: unknown) => void
		sdk.projectViewTasksList.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
		const previousLoad = taskList.loadTasks()

		sdk.projectViewTasksList.mockImplementationOnce(() => new Promise(resolve => { resolveCurrent = resolve }))
		await router.push('/projects/2/21')
		await flushPromises()
		expect(taskList.tasks.value).toEqual([])

		const currentTasks = [normalizeTask(createTaskDraft({id: 2, project_id: 2, title: 'Current project task'}))]
		resolveCurrent(taskListResponse(currentTasks, 1))
		await flushPromises()
		expect(taskList.tasks.value).toEqual(currentTasks)

		resolveOld(taskListResponse([createTaskDraft({id: 1, project_id: 1, title: 'Previous project task'})], 1))
		await previousLoad
		expect(taskList.tasks.value).toEqual(currentTasks)
	})

	it('resets pagination when the user changes the sort', async () => {
		const {taskList} = await mountRoutedTaskList()
		taskList.currentPage.value = 3
		await flushPromises()
		expect(taskList.currentPage.value).toBe(3)

		taskList.sortByParam.value = {due_date: 'asc'}
		await flushPromises()
		expect(taskList.currentPage.value).toBe(1)
		expect(lastQuery().sort_by).toEqual(['due_date'])
	})
})
