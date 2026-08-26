import AbstractService from '@/services/abstractService'
import type {ITaskBucket} from '@/modelTypes/ITaskBucket'
import TaskBucketModel from '@/models/taskBucket'
import {invalidateCachedTask} from '@/helpers/taskCache'

export default class TaskBucketService extends AbstractService<ITaskBucket> {
	constructor() {
		super({
			update: '/projects/{projectId}/views/{projectViewId}/buckets/{bucketId}/tasks',
		})
	}
	
	modelFactory(data: Partial<ITaskBucket>) {
		return new TaskBucketModel(data)
	}

	// Moving a task into the done bucket marks it done without going through TaskService.
	async update(model: ITaskBucket) {
		const updated = await super.update(model)
		invalidateCachedTask(model.taskId)
		return updated
	}
}
