import {beforeEach, describe, expect, it, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

vi.mock('vue-router', () => ({
	useRouter: () => ({push: vi.fn()}),
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({t: (key: string) => key}),
	createI18n: () => ({global: {t: (key: string) => key}}),
}))

vi.mock('@/stores/base', () => ({
	useBaseStore: () => ({
		currentProject: null,
		setCurrentProject: vi.fn(),
	}),
}))

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		authUser: true,
		info: null,
	}),
}))

import {useKanbanStore} from './kanban'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'

function makeBucket(id: number, title: string, tasks: ITask[] = []): IBucket {
	return {
		id,
		title,
		projectViewId: 1,
		tasks,
		count: tasks.length,
	} as IBucket
}

function makeTask(id: number, bucketId: number): ITask {
	return {
		id,
		title: `Task ${id}`,
		bucketId,
	} as ITask
}

describe('kanban store: moveTaskToBucket', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('relocates a task from its current bucket into the target bucket', () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'To-Do'), makeBucket(2, 'Done')])

		const task = makeTask(42, 2)
		kanban.addTaskToBucket(task)
		expect(kanban.buckets[1].tasks.map(t => t.id)).toEqual([42])

		kanban.moveTaskToBucket(task, 1)

		expect(kanban.buckets[0].tasks.map(t => t.id)).toEqual([42])
		expect(kanban.buckets[0].tasks[0].bucketId).toBe(1)
		expect(kanban.buckets[1].tasks.map(t => t.id)).toEqual([])
	})

	it('is a no-op when the task is already in the target bucket', () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'To-Do'), makeBucket(2, 'Done')])

		const task = makeTask(42, 1)
		kanban.addTaskToBucket(task)

		kanban.moveTaskToBucket(task, 1)

		expect(kanban.buckets[0].tasks.map(t => t.id)).toEqual([42])
		expect(kanban.buckets[1].tasks.map(t => t.id)).toEqual([])
	})

	it('is a no-op when the task is not present in any bucket', () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'To-Do'), makeBucket(2, 'Done')])

		const strayTask = makeTask(99, 2)
		kanban.moveTaskToBucket(strayTask, 1)

		expect(kanban.buckets[0].tasks).toEqual([])
		expect(kanban.buckets[1].tasks).toEqual([])
	})
})
