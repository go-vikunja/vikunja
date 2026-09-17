import {beforeEach, describe, expect, it, vi} from 'vitest'
import {defineComponent} from 'vue'
import {mount} from '@vue/test-utils'
import {createPinia} from 'pinia'
import {VueQueryPlugin} from '@tanstack/vue-query'

import {queryClient} from '@/client/queryClient'

const sdk = vi.hoisted(() => ({
	tasksCreate: vi.fn(),
	tasksBulkCreate: vi.fn(),
	labelsList: vi.fn(),
	labelsCreate: vi.fn(),
	taskLabelsCreate: vi.fn(),
}))
vi.mock('@/client/generated', () => sdk)
vi.mock('@/message', () => ({error: vi.fn(), success: vi.fn(), translatedError: (key: string) => new Error(key)}))
vi.mock('vue-router', async importOriginal => ({
	...(await importOriginal<typeof import('vue-router')>()),
	useRouter: () => ({currentRoute: {value: {params: {}}}}),
}))

import {error} from '@/message'
import {useTaskActions} from './useTaskActions'

function taskActions() {
	let actions!: ReturnType<typeof useTaskActions>
	mount(defineComponent({
		setup() {
			actions = useTaskActions()
			return () => null
		},
	}), {global: {plugins: [createPinia(), [VueQueryPlugin, {queryClient}]]}})
	return actions
}

describe('useTaskActions', () => {
	beforeEach(() => {
		queryClient.clear()
		vi.mocked(error).mockClear();
		[sdk.tasksCreate, sdk.tasksBulkCreate, sdk.labelsList, sdk.labelsCreate, sdk.taskLabelsCreate].forEach(mock => mock.mockReset())
		sdk.labelsList.mockResolvedValue({data: {items: [], total_pages: 1}})
	})

	it('keeps the created task when its label cannot be created', async () => {
		sdk.tasksCreate.mockResolvedValue({data: {id: 7, title: 'Buy milk', project_id: 1}})
		sdk.labelsCreate.mockRejectedValue({status: 500})

		const task = await taskActions().createNewTask({title: 'Buy milk *Urgent', project_id: 1})

		expect(task).toMatchObject({id: 7, title: 'Buy milk'})
		expect(task.labels).toEqual([])
		expect(sdk.taskLabelsCreate).not.toHaveBeenCalled()
		expect(error).toHaveBeenCalledTimes(1)
	})

	it('reports a failed label write on a bulk create without failing the create', async () => {
		sdk.labelsList.mockResolvedValue({data: {items: [{id: 9, title: 'Urgent'}], total_pages: 1}})
		sdk.tasksBulkCreate.mockResolvedValue({data: {tasks: [{id: 7, title: 'Buy milk', project_id: 1}]}})
		sdk.taskLabelsCreate.mockRejectedValue({status: 500})

		const result = await taskActions().createNewTasksBulk([{title: 'Buy milk *Urgent', project_id: 1}])

		expect(result.error).toBeNull()
		expect(result.tasks[0]).toMatchObject({id: 7, title: 'Buy milk'})
		expect(sdk.labelsCreate).not.toHaveBeenCalled()
		expect(error).toHaveBeenCalledTimes(1)
	})
})
