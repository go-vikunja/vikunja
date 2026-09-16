import type {ITask} from '@/modelTypes/ITask'
import TaskModel from '@/models/task'
import TaskService from '@/services/task'
import {deleteCachedTask, getCachedTask, setCachedTask} from '@/helpers/taskCache'

function isPermanentError(e: unknown): boolean {
	const status = (e as {response?: {status?: number}})?.response?.status
	return status === 403 || status === 404
}

// One request per task id shared by all callers. 403/404 stay cached; other failures are evicted so a later call can retry.
export function fetchTaskById(id: number): Promise<ITask> {
	const cached = getCachedTask(id)
	if (cached) {
		return cached
	}

	const task: Promise<ITask> = new TaskService().get(new TaskModel({id}))
	task.catch(e => {
		if (!isPermanentError(e)) {
			deleteCachedTask(id, task)
		}
	})
	setCachedTask(id, task)
	return task
}
