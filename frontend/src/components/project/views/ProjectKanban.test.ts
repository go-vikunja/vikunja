import {shallowMount, flushPromises} from '@vue/test-utils'
import {describe, expect, it, vi, beforeEach} from 'vitest'
import {createPinia} from 'pinia'
import {nextTick} from 'vue'
import draggable from 'zhyswan-vuedraggable'
import {QueryClient, VueQueryPlugin} from '@tanstack/vue-query'

const sdk = vi.hoisted(() => ({
	projectViewBucketsTasksList: vi.fn(),
	projectViewTasksList: vi.fn(),
	bucketsUpdate: vi.fn(),
	tasksCreate: vi.fn(),
}))
const {projectViewBucketsTasksList: loadBoard, projectViewTasksList: loadBucketPage, bucketsUpdate: updateBucket} = sdk

// Asymmetric positions: the midpoint between the neighbours must differ from the dragged bucket's own position.
const buckets = [
	{id: 1, title: 'First', position: 100, tasks: [{id: 11, title: 'Task 11'}], count: 1, limit: 0},
	{id: 2, title: 'Second', position: 250, tasks: [], count: 0, limit: 0},
	{id: 3, title: 'Third', position: 300, tasks: [], count: 0, limit: 0},
]

vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn()}))
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

const drop = vi.hoisted(() => ({
	result: {moved: false, targetProjectId: null as number | null},
	task: {id: 11, title: 'Task 11'},
}))
vi.mock('@/composables/useTaskDragToProject', () => ({
	useTaskDragToProject: () => ({
		handleTaskDropToProject: async (
			_e: unknown,
			onSuccess?: (task: unknown, targetProjectId: number) => void,
		) => {
			if (drop.result.moved) onSuccess?.(drop.task, drop.result.targetProjectId!)
			return drop.result
		},
	}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		settings: {
			frontend_settings: {
				always_show_bucket_task_count: false,
				quick_add_magic_mode: 'vikunja',
				quick_add_default_reminders: [],
			},
		},
	}),
}))

vi.mock('@/client/queries/savedFilters', () => ({
	savedFilterQuery: (id: number) => ({
		queryKey: ['saved-filters', 'detail', id],
		queryFn: async () => null,
		enabled: false,
	}),
}))

vi.mock('@vueuse/router', () => ({
	useRouteQuery: () => ({value: undefined}),
}))

vi.mock('vue-router', () => ({
	useRouter: () => ({push: vi.fn(), currentRoute: {value: {fullPath: '/', params: {}}}}),
}))

vi.mock('vue-i18n', async importOriginal => ({
	...await importOriginal<typeof import('vue-i18n')>(),
	useI18n: () => ({t: (key: string) => key}),
}))

import ProjectKanban from './ProjectKanban.vue'

let queryClient: QueryClient

async function mountKanban(stubs: Record<string, unknown> = {}, projectId = 1) {
	queryClient = new QueryClient()
	const wrapper = shallowMount(ProjectKanban, {
		props: {
			isLoadingProject: false,
			projectId,
			viewId: 10,
		},
		global: {
			plugins: [createPinia(), [VueQueryPlugin, {queryClient}]],
			mocks: {$t: (key: string) => key},
			directives: {focus: () => {}, tooltip: () => {}, cy: () => {}},
			stubs: {
				ProjectWrapper: {template: '<div><slot name="default"/></div>'},
				...stubs,
			},
		},
	})

	await flushPromises()
	return wrapper
}

// vuedraggable is stubbed away by shallowMount, but the add task control lives in its footer slot.
const DRAGGABLE_STUB = {
	props: ['modelValue'],
	template: '<ul><template v-for="(element, index) in modelValue"><slot name="item" :element="element" :index="index" /></template><slot name="footer" /></ul>',
}

function dragEndEvent(bucketId: string) {
	const item = document.createElement('li')
	item.dataset.bucketId = bucketId
	// The DOM index sortable reports does not have to be a valid bucket index
	return {item, newIndex: 42}
}

function taskDragEndEvent(taskId: number, bucketIndex = 0) {
	const item = document.createElement('li')
	item.dataset.taskId = String(taskId)
	const list = document.createElement('ul')
	list.dataset.bucketIndex = String(bucketIndex)
	return {item, to: list, from: list, originalEvent: new MouseEvent('mouseup'), newIndex: 0}
}

// The bucket draggable comes first, each bucket renders its own task draggable after it.
function taskDraggable(wrapper: Awaited<ReturnType<typeof mountKanban>>) {
	return wrapper.findAllComponents(draggable)[1]
}

function cachedBoard() {
	const entries = queryClient.getQueriesData<{
		buckets: {id: number, tasks: {id: number}[], count: number}[],
	}>({queryKey: ['kanban']})
	return entries[0][1]
}

describe('ProjectKanban', () => {
	beforeEach(() => {
		updateBucket.mockClear()
		updateBucket.mockResolvedValue({data: {id: 2, position: 200}})
		loadBoard.mockClear()
		loadBoard.mockResolvedValue({data: {items: buckets}})
		loadBucketPage.mockClear()
		loadBucketPage.mockResolvedValue({data: {items: []}})
		sdk.tasksCreate.mockReset()
		drop.result = {moved: false, targetProjectId: null}
	})

	it('saves the position of the dropped bucket', async () => {
		const wrapper = await mountKanban()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('2'))
		await flushPromises()

		expect(updateBucket).toHaveBeenCalledWith(expect.objectContaining({
			path: {project: 1, view: 10, bucket: 2},
			body: expect.objectContaining({position: 200}),
		}))
	})

	it('keeps the add bucket column while the board refetches in the background', async () => {
		const wrapper = await mountKanban()
		expect(wrapper.find('.new-bucket').exists()).toBe(true)

		loadBoard.mockReturnValue(new Promise(() => {}))
		queryClient.refetchQueries()
		await nextTick()

		expect(wrapper.find('.new-bucket').exists()).toBe(true)
	})

	it('does not load another bucket page while a task is dragged out of the bucket', async () => {
		const wrapper = await mountKanban()
		// The scroll handler is wired through vuedraggable's component-data, which the stub does not render.
		const vm = wrapper.vm as unknown as {
			updateTasks: (bucketId: number, tasks: unknown[]) => void,
			handleTaskContainerScroll: (id: number, el: HTMLElement) => void,
		}

		vm.updateTasks(1, [])
		vm.handleTaskContainerScroll(1, document.createElement('div'))
		await flushPromises()

		expect(loadBucketPage).not.toHaveBeenCalled()
	})

	it('does not refetch the board after a task was dropped on another project', async () => {
		const wrapper = await mountKanban({draggable: DRAGGABLE_STUB})
		drop.result = {moved: true, targetProjectId: 2}
		loadBoard.mockClear()

		taskDraggable(wrapper).vm.$emit('end', taskDragEndEvent(11))
		await flushPromises()

		expect(loadBoard).not.toHaveBeenCalled()
	})

	it('drops the card from a saved filter board when the task moved to another project', async () => {
		const wrapper = await mountKanban({draggable: DRAGGABLE_STUB}, -2)
		drop.result = {moved: true, targetProjectId: 2}

		taskDraggable(wrapper).vm.$emit('end', taskDragEndEvent(11))
		await flushPromises()

		expect(cachedBoard()?.buckets[0].tasks).toEqual([])
		expect(cachedBoard()?.buckets[0].count).toBe(0)
	})

	it('leaves a real project board to the mutation when the task moved to another project', async () => {
		const wrapper = await mountKanban({draggable: DRAGGABLE_STUB})
		drop.result = {moved: true, targetProjectId: 2}

		taskDraggable(wrapper).vm.$emit('end', taskDragEndEvent(11))
		await flushPromises()

		expect(cachedBoard()?.buckets[0].tasks).toHaveLength(1)
		expect(cachedBoard()?.buckets[0].count).toBe(1)
	})

	it('re-enables the add task input once the create settles, without a board request', async () => {
		const wrapper = await mountKanban({draggable: DRAGGABLE_STUB})
		let resolveCreate!: (value: unknown) => void
		sdk.tasksCreate.mockImplementation(() => new Promise(resolve => { resolveCreate = resolve }))

		;(wrapper.vm as unknown as {toggleShowNewTaskInput: (id: number) => void}).toggleShowNewTaskInput(1)
		await nextTick()
		await wrapper.find('.bucket-footer input').setValue('New task')
		await wrapper.find('.bucket-footer input').trigger('keyup.enter')
		await flushPromises()
		expect(wrapper.find('.bucket-footer input').attributes('disabled')).toBeDefined()

		loadBoard.mockClear()
		resolveCreate({data: {id: 12, title: 'New task', project_id: 1, bucket_id: 1}})
		await flushPromises()

		expect(wrapper.find('.bucket-footer input').attributes('disabled')).toBeUndefined()
		expect(loadBoard).not.toHaveBeenCalled()
	})

	it('closes the add task input when the created task fills the bucket limit', async () => {
		loadBoard.mockResolvedValue({data: {items: [{...buckets[0], count: 1, limit: 2}, buckets[1], buckets[2]]}})
		sdk.tasksCreate.mockResolvedValue({data: {id: 12, title: 'New task', project_id: 1, bucket_id: 1}})
		const wrapper = await mountKanban({draggable: DRAGGABLE_STUB})

		;(wrapper.vm as unknown as {toggleShowNewTaskInput: (id: number) => void}).toggleShowNewTaskInput(1)
		await nextTick()
		await wrapper.find('.bucket-footer input').setValue('New task')
		await wrapper.find('.bucket-footer input').trigger('keyup.enter')
		await flushPromises()

		expect(wrapper.find('.bucket-footer input').exists()).toBe(false)
	})

	it('does nothing when the dropped bucket is gone', async () => {
		const wrapper = await mountKanban()

		wrapper.findComponent(draggable).vm.$emit('end', dragEndEvent('42'))

		expect(updateBucket).not.toHaveBeenCalled()
	})
})
