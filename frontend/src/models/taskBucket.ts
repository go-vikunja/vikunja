import AbstractModel from '@/models/abstractModel'
import type {ITaskBucket} from '@/modelTypes/ITaskBucket'
import TaskModel from '@/models/task.ts'
import BucketModel from '@/models/bucket.ts'

export default class TaskBucketModel extends AbstractModel<ITaskBucket> implements ITaskBucket {
	taskId = 0
	bucketId = 0
	projectViewId = 0
	projectId = 0
	task = undefined
	bucket = undefined

	constructor(data: Partial<ITaskBucket>) {
		super()
		this.assignData(data)
		
		if (data.task) {
			this.task = new TaskModel(data.task)
		}
		
		if (data.bucket) {
			this.bucket = new BucketModel(data.bucket)
		}
	}
}
