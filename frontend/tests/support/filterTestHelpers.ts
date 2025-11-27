import {TaskFactory} from '../factories/task'
import {TaskBucketFactory} from '../factories/task_buckets'

export async function createTasksWithPriorities(buckets?: Array<{id: number; project_view_id: number}>) {
	await TaskFactory.truncate()

	const highPriorityTask1 = (await TaskFactory.create(1, {
		id: 1,
		project_id: 1,
		priority: 4,
		title: 'High Priority Task 1',
	}, false))[0]

	const highPriorityTask2 = (await TaskFactory.create(1, {
		id: 2,
		project_id: 1,
		priority: 4,
		title: 'High Priority Task 2',
	}, false))[0]

	const lowPriorityTask1 = (await TaskFactory.create(1, {
		id: 3,
		project_id: 1,
		priority: 1,
		title: 'Low Priority Task 1',
	}, false))[0]

	const lowPriorityTask2 = (await TaskFactory.create(1, {
		id: 4,
		project_id: 1,
		priority: 1,
		title: 'Low Priority Task 2',
	}, false))[0]

	// If buckets are provided (for Kanban), add tasks to buckets
	if (buckets && buckets.length > 0) {
		await TaskBucketFactory.truncate()
		await TaskBucketFactory.create(1, {
			task_id: highPriorityTask1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: highPriorityTask2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: lowPriorityTask1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: lowPriorityTask2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
	}

	return {
		highPriorityTasks: [highPriorityTask1, highPriorityTask2],
		lowPriorityTasks: [lowPriorityTask1, lowPriorityTask2],
	}
}

export async function createTasksWithSearch(buckets?: Array<{id: number; project_view_id: number}>) {
	await TaskFactory.truncate()

	const task1 = (await TaskFactory.create(1, {
		id: 1,
		project_id: 1,
		title: 'Regular task 1',
	}, false))[0]

	const task2 = (await TaskFactory.create(1, {
		id: 2,
		project_id: 1,
		title: 'Regular task 2',
	}, false))[0]

	const task3 = (await TaskFactory.create(1, {
		id: 3,
		project_id: 1,
		title: 'Regular task 3',
	}, false))[0]

	const searchableTask = (await TaskFactory.create(1, {
		id: 4,
		project_id: 1,
		title: 'Meeting notes for project',
	}, false))[0]

	// If buckets are provided (for Kanban), add tasks to buckets
	if (buckets && buckets.length > 0) {
		await TaskBucketFactory.truncate()
		await TaskBucketFactory.create(1, {
			task_id: task1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: task2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: task3.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: searchableTask.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		}, false)
	}

	return { searchableTask }
}
