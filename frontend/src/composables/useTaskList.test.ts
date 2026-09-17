import {describe, it, expect, beforeEach, afterEach, vi} from 'vitest'
import {defineComponent, h, nextTick} from 'vue'
import {mount, flushPromises, enableAutoUnmount} from '@vue/test-utils'
import {setActivePinia, createPinia} from 'pinia'
import {createRouter, createMemoryHistory, RouterView, type Router} from 'vue-router'

import type {Task as ITask} from '@/client/generated'
import {createTaskDraft} from '@/helpers/task'

import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'
const getAll = vi.fn(async (..._args: unknown[]) => [] as ITask[])
vi.mock('@/client/generated', () => ({
 projectViewTasksList: async ({path, query}: {path: {project: number, view: number}, query: {q?: string, page: number}}) => ({data: {items: await getAll({projectId: path.project, viewId: path.view}, {...query, s: query.q}, query.page), page: query.page, total_pages: 5}}),
}))
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
import {error} from '@/message'
const queryClient = new QueryClient({defaultOptions: {queries: {retry: false}}})
beforeEach(() => queryClient.clear())
import {useTaskList, buildStoredQuery, parsePageQuery} from './useTaskList'
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

describe('parsePageQuery', () => {
	it.each([
		['2', 2],
		['abc', 1],
		['0', 1],
		['-3', 1],
		['1.5', 1],
		[undefined, 1],
	])('turns %s into page %i', (raw, expected) => {
		expect(parsePageQuery(raw)).toBe(expected)
	})
})

function lastRequestParams(): Record<string, unknown> {
	return getAll.mock.calls.at(-1)?.[1] as Record<string, unknown>
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
	beforeEach(() => {
		localStorage.clear()
		setActivePinia(createPinia())
		getAll.mockClear()
	})

	it('omits the sort while searching with the default sort so the backend ranks by relevance', async () => {
		await mountTaskList({s: 'find me'})

		const params = lastRequestParams()
		expect(params.s).toBe('find me')
		expect(params.sort_by).toEqual([])
		expect(params.order_by).toEqual([])
	})

	it('keeps an explicit user sort while searching so the user sort is respected', async () => {
		await mountTaskList({s: 'find me', sort: 'title:asc'})

		const params = lastRequestParams()
		expect(params.s).toBe('find me')
		expect(params.sort_by).toEqual(['title'])
		expect(params.order_by).toEqual(['asc'])
	})

	it.each(['abc', '0'])('loads the first page when the url asks for page %s', async (page) => {
		await mountTaskList({page})

		expect(getAll.mock.calls.at(-1)?.[2]).toBe(1)
	})

	it('sends the default sort when not searching', async () => {
		await mountTaskList({})

		const params = lastRequestParams()
		expect(params.s).toBe('')
		expect(params.sort_by).not.toHaveLength(0)
		// id always sorts last so other sort columns take precedence.
		expect(params.sort_by).toEqual(['id'])
		expect(params.order_by).toEqual(['desc'])
	})
})

describe('useTaskList error reporting', () => {
	beforeEach(() => {
		localStorage.clear()
		setActivePinia(createPinia())
		getAll.mockReset()
		vi.mocked(error).mockClear()
	})

	it('toasts a failing load instead of showing an empty list', async () => {
		getAll.mockRejectedValueOnce(new Error('Server is on fire'))

		await mountTaskList({})
		await flushPromises()

		expect(error).toHaveBeenCalledWith(expect.objectContaining({message: 'Server is on fire'}))
	})

	it('stays silent when the request was aborted', async () => {
		getAll.mockRejectedValueOnce(new DOMException('aborted', 'AbortError'))

		await mountTaskList({})
		await flushPromises()

		expect(error).not.toHaveBeenCalled()
	})
})

describe('useTaskList restoring stored query into the url', () => {
	beforeEach(() => {
		localStorage.clear()
		setActivePinia(createPinia())
		getAll.mockClear()
	})

	it('writes the persisted sort into the url when the url has none', async () => {
		localStorage.setItem('viewFilters', JSON.stringify({1: {sort: 'due_date:asc'}}))

		const router = await mountTaskList({})
		await flushPromises()

		expect(getAll).toHaveBeenCalledTimes(1)
		expect(router.currentRoute.value.query.sort).toBe('due_date:asc')
		expect(lastRequestParams().sort_by).toEqual(['due_date'])
	})

	it('keeps an explicit url sort over the persisted one', async () => {
		localStorage.setItem('viewFilters', JSON.stringify({1: {sort: 'due_date:asc'}}))

		const router = await mountTaskList({sort: 'title:desc'})

		expect(router.currentRoute.value.query.sort).toBe('title:desc')
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
	beforeEach(() => {
		localStorage.clear()
		setActivePinia(createPinia())
		getAll.mockReset()
	})

	it.each([1, 3])('restores the sort and page %i when returning to a project', async (page) => {
		const {router, taskList} = await mountRoutedTaskList()
		taskList.sortByParam.value = {due_date: 'asc'}
		await flushPromises()
		taskList.currentPage.value = page
		await flushPromises()

		getAll.mockClear()
		await router.push('/projects/2/21')
		await flushPromises()
		expect(getAll.mock.calls).toEqual([[
			{projectId: 2, viewId: 21},
			expect.objectContaining({sort_by: ['position'], order_by: ['asc']}),
			1,
		]])
		expect(taskList.sortByParam.value).toEqual({position: 'asc'})

		getAll.mockClear()
		await router.push('/projects/1/11')
		await flushPromises()
		expect(getAll.mock.calls).toEqual([[
			{projectId: 1, viewId: 11},
			expect.objectContaining({sort_by: ['due_date'], order_by: ['asc']}),
			page,
		]])
		expect(router.currentRoute.value.query.sort).toBe('due_date:asc')
		expect(taskList.sortByParam.value).toEqual({due_date: 'asc'})
		expect(taskList.currentPage.value).toBe(page)
		expect(lastRequestParams().sort_by).toEqual(['due_date'])
	})

	it('keeps an explicit sort when navigating to a project with a saved sort', async () => {
		const {router, taskList} = await mountRoutedTaskList()
		useViewFiltersStore().setViewQuery(21, {sort: 'due_date:asc'})

		await router.push('/projects/2/21?sort=title:desc')
		await flushPromises()
		expect(taskList.sortByParam.value).toEqual({title: 'desc'})
		expect(lastRequestParams().sort_by).toEqual(['title'])
	})

	it('ignores a response from a project the user has left', async () => {
		const {router, taskList} = await mountRoutedTaskList()
		let resolveOld!: (tasks: ITask[]) => void
		let resolveCurrent!: (tasks: ITask[]) => void
		getAll.mockImplementationOnce(() => new Promise(resolve => { resolveOld = resolve }))
		const previousLoad = taskList.loadTasks()

		getAll.mockImplementationOnce(() => new Promise(resolve => { resolveCurrent = resolve }))
		await router.push('/projects/2/21')
		await flushPromises()
		expect(taskList.tasks.value).toEqual([])

		const currentTasks = [createTaskDraft({id: 2, project_id: 2, title: 'Current project task'})]
		resolveCurrent(currentTasks)
		await flushPromises()
		expect(taskList.tasks.value).toEqual(currentTasks)

		resolveOld([createTaskDraft({id: 1, project_id: 1, title: 'Previous project task'})])
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
		expect(lastRequestParams().sort_by).toEqual(['due_date'])
	})
})
