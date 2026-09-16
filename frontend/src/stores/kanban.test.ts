import {beforeEach, describe, expect, it, vi} from 'vitest'
import {createPinia, setActivePinia} from 'pinia'

const {bucketUpdate, bucketDelete, loadBuckets, refreshProjects, refreshProject} = vi.hoisted(() => ({
	bucketUpdate: vi.fn(),
	bucketDelete: vi.fn(),
	loadBuckets: vi.fn(),
	refreshProjects: vi.fn(),
	refreshProject: vi.fn(),
}))

vi.mock('@/services/bucket', () => ({
	default: class {
		update = bucketUpdate
		delete = bucketDelete
	},
}))

vi.mock('@/services/taskCollection', () => ({default: class {getAll = loadBuckets}}))
vi.mock('@/client/queries/projects', async importOriginal => ({
	...await importOriginal<typeof import('@/client/queries/projects')>(),
	refreshProjects,
	refreshProject,
}))

vi.mock('vue-router', () => ({
	useRouter: () => ({push: vi.fn()}),
}))

vi.mock('vue-i18n', () => ({
	useI18n: () => ({t: (key: string) => key}),
	createI18n: () => ({global: {t: (key: string) => key}}),
}))

vi.mock('@/stores/base', async () => {
	const {reactive, ref} = await import('vue')
	const baseStore = reactive({
		currentProjectId: ref(0),
		currentProject: null,
		setCurrentProject: vi.fn(),
	})
	return {useBaseStore: () => baseStore}
})

vi.mock('@/stores/auth', () => ({
	useAuthStore: () => ({
		authUser: true,
		info: null,
	}),
}))

import {useKanbanStore} from './kanban'
import {useBaseStore} from '@/stores/base'

import {queryClient} from '@/client/queryClient'
import {normalizeProject, projectKeys, type ProjectListResult} from '@/client/queries/projects'
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

describe('kanban store: synchronous project reads', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.clear()
		useBaseStore().currentProjectId = 0
		useBaseStore().currentProjectViewId = 2
	})

	it('does not subscribe to project queries during store setup', () => {
		useKanbanStore()
		expect(queryClient.getQueryCache().getAll()).toEqual([])
	})

	it('uses the current project cache on each synchronous bucket-placement call', () => {
		const kanban = useKanbanStore()
		useBaseStore().currentProjectId = 1
		queryClient.setQueryData(projectKeys.detail(1), normalizeProject({id: 1, views: [{id: 2, default_bucket_id: 3, done_bucket_id: 4}]}))
		kanban.setBuckets([makeBucket(3, 'Default'), makeBucket(4, 'Done')])
		const task = {...makeTask(9, 3), done: true}
		kanban.addTaskToBucket(task)

		expect(kanban.ensureTaskIsInCorrectBucket(task)).toBeUndefined()
		expect(kanban.buckets[1].tasks[0].id).toBe(9)

		queryClient.setQueryData(projectKeys.detail(1), normalizeProject({id: 1, views: [{id: 2, default_bucket_id: 3, done_bucket_id: 4}]}))
		task.done = false
		kanban.ensureTaskIsInCorrectBucket(task)
		expect(kanban.buckets[0].tasks[0].id).toBe(9)
	})
})

describe('kanban store: deleteBucket', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.clear()
		bucketDelete.mockReset().mockResolvedValue(undefined)
		loadBuckets.mockReset().mockResolvedValue([])
		refreshProject.mockReset().mockResolvedValue(undefined)
		refreshProjects.mockReset().mockResolvedValue(undefined)
	})

	it.each([
		{defaultBucket: 5, doneBucket: 0, refresh: true},
		{defaultBucket: 0, doneBucket: 5, refresh: true},
		{defaultBucket: 1, doneBucket: 2, refresh: false},
	])('refreshes the project only when deleting a configured bucket: %j', async ({defaultBucket, doneBucket, refresh}) => {
		const project = normalizeProject({id: 1, views: [{
			id: 2,
			project_id: 1,
			default_bucket_id: defaultBucket,
			done_bucket_id: doneBucket,
		}]})
		queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [project], favoriteProject: null, savedFilterProjects: [],
		})
		const kanban = useKanbanStore()
		const bucket = {...makeBucket(5, 'Delete'), projectId: 1, projectViewId: 2}
		kanban.setBuckets([bucket])

		await kanban.deleteBucket({bucket, params: {}})

		expect(bucketDelete).toHaveBeenCalledWith(bucket)
		expect(refreshProject).toHaveBeenCalledTimes(Number(refresh))
		expect(refreshProjects).toHaveBeenCalledTimes(Number(refresh))
		if (refresh) {
			expect(refreshProject).toHaveBeenCalledWith(1)
		}
		expect(loadBuckets).toHaveBeenCalledOnce()
	})
})

describe('kanban store: moveTaskToBucket', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		queryClient.setQueryData<ProjectListResult>(projectKeys.list(), {
			projects: [],
			favoriteProject: null,
			savedFilterProjects: [],
		})
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

	it('keeps the task where it is when the target bucket is not loaded', () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'To-Do'), makeBucket(2, 'Done')])

		const task = makeTask(42, 1)
		kanban.addTaskToBucket(task)

		expect(() => kanban.moveTaskToBucket(task, 404)).not.toThrow()

		expect(kanban.buckets[0].tasks.map(t => t.id)).toEqual([42])
		expect(task.bucketId).toBe(1)
	})
})

describe('kanban store: updateBucket', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
		bucketUpdate.mockReset()
	})

	it('does not touch the board when the update response arrives after the board was replaced', async () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'First'), makeBucket(2, 'Second'), makeBucket(3, 'Third')])

		let respond: (bucket: IBucket) => void = () => {}
		bucketUpdate.mockReturnValue(new Promise<IBucket>(resolve => respond = resolve))

		const updating = kanban.updateBucket({id: 3, title: 'Renamed'})

		// Navigating to another view clears the board while the request is in flight
		kanban.setBuckets([])

		respond(makeBucket(3, 'Renamed'))
		await updating

		expect(kanban.buckets).toEqual([])
	})

	it('does not send an update for a bucket which is not on the board', async () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'First')])

		await kanban.updateBucket({id: 404, title: 'Renamed'})

		expect(bucketUpdate).not.toHaveBeenCalled()
		expect(kanban.buckets).toHaveLength(1)
	})

	it('keeps the loaded tasks when the api returns the bucket without them', async () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'First', [makeTask(42, 1)])])

		bucketUpdate.mockResolvedValue(makeBucket(1, 'Renamed'))

		await kanban.updateBucket({id: 1, title: 'Renamed'})

		expect(kanban.buckets[0].title).toBe('Renamed')
		expect(kanban.buckets[0].tasks.map(t => t.id)).toEqual([42])
	})
})

describe('kanban store: addTaskToBucket', () => {
	beforeEach(() => {
		setActivePinia(createPinia())
	})

	it('does nothing when the bucket is not loaded', () => {
		const kanban = useKanbanStore()
		kanban.setBuckets([makeBucket(1, 'To-Do')])

		expect(() => kanban.addTaskToBucket(makeTask(42, 404))).not.toThrow()

		expect(kanban.buckets[0].tasks).toEqual([])
		expect(kanban.buckets).toHaveLength(1)
	})
})
