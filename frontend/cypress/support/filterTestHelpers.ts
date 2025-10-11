import {TaskFactory} from '../factories/task'
import {TaskBucketFactory} from '../factories/task_buckets'

export function createTasksWithPriorities(buckets?: any[]) {
	TaskFactory.truncate()

	const highPriorityTask1 = TaskFactory.create(1, {
		id: 1,
		project_id: 1,
		priority: 4,
		title: 'High Priority Task 1',
	})[0]

	const highPriorityTask2 = TaskFactory.create(1, {
		id: 2,
		project_id: 1,
		priority: 4,
		title: 'High Priority Task 2',
	})[0]

	const lowPriorityTask1 = TaskFactory.create(1, {
		id: 3,
		project_id: 1,
		priority: 1,
		title: 'Low Priority Task 1',
	})[0]

	const lowPriorityTask2 = TaskFactory.create(1, {
		id: 4,
		project_id: 1,
		priority: 1,
		title: 'Low Priority Task 2',
	})[0]

	// If buckets are provided (for Kanban), add tasks to buckets
	if (buckets && buckets.length > 0) {
		TaskBucketFactory.truncate()
		TaskBucketFactory.create(1, {
			task_id: highPriorityTask1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: highPriorityTask2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: lowPriorityTask1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: lowPriorityTask2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
	}

	return {
		highPriorityTasks: [highPriorityTask1, highPriorityTask2],
		lowPriorityTasks: [lowPriorityTask1, lowPriorityTask2],
	}
}

export function createTasksWithSearch(buckets?: any[]) {
	TaskFactory.truncate()

	const task1 = TaskFactory.create(1, {
		id: 1,
		project_id: 1,
		title: 'Regular task 1',
	})[0]

	const task2 = TaskFactory.create(1, {
		id: 2,
		project_id: 1,
		title: 'Regular task 2',
	})[0]

	const task3 = TaskFactory.create(1, {
		id: 3,
		project_id: 1,
		title: 'Regular task 3',
	})[0]

	const searchableTask = TaskFactory.create(1, {
		id: 4,
		project_id: 1,
		title: 'Meeting notes for project',
	})[0]

	// If buckets are provided (for Kanban), add tasks to buckets
	if (buckets && buckets.length > 0) {
		TaskBucketFactory.truncate()
		TaskBucketFactory.create(1, {
			task_id: task1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: task2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: task3.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: searchableTask.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
	}

	return { searchableTask }
}

export function createTasksWithPriorityAndSearch(buckets?: any[]) {
	TaskFactory.truncate()

	const matchingTask = TaskFactory.create(1, {
		id: 1,
		project_id: 1,
		priority: 5,
		title: 'Urgent meeting preparation',
	})[0]

	const nonMatchingTask1 = TaskFactory.create(1, {
		id: 2,
		project_id: 1,
		priority: 5,
		title: 'Important task',
	})[0]

	const nonMatchingTask2 = TaskFactory.create(1, {
		id: 3,
		project_id: 1,
		priority: 1,
		title: 'Optional meeting attendance',
	})[0]

	// If buckets are provided (for Kanban), add tasks to buckets
	if (buckets && buckets.length > 0) {
		TaskBucketFactory.truncate()
		TaskBucketFactory.create(1, {
			task_id: matchingTask.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: nonMatchingTask1.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
		TaskBucketFactory.create(1, {
			task_id: nonMatchingTask2.id,
			bucket_id: buckets[0].id,
			project_view_id: buckets[0].project_view_id,
		})
	}

	return {
		matchingTask,
		nonMatchingTask1,
		nonMatchingTask2,
	}
}
