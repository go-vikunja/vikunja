import AbstractService from '@/services/abstractService'
import type {ITaskBucket} from '@/modelTypes/ITaskBucket'
import TaskBucketModel from '@/models/taskBucket'

export default class TaskBucketService extends AbstractService<ITaskBucket> {
	constructor() {
		super({
			update: '/projects/{projectId}/views/{projectViewId}/buckets/{bucketId}/tasks',
		})
	}
	
	modelFactory(data: Partial<ITaskBucket>) {
		return new TaskBucketModel(data)
	}
}
