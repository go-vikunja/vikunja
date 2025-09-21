import AbstractModel from '@/models/abstractModel'
import type {ITaskBucket} from '@/modelTypes/ITaskBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IBucket} from '@/modelTypes/IBucket'
import TaskModel from '@/models/task.ts'
import BucketModel from '@/models/bucket.ts'

export default class TaskBucketModel extends AbstractModel<ITaskBucket> implements ITaskBucket {
	taskId = 0
	bucketId = 0
	projectViewId = 0
	projectId = 0
	task: ITask | null = null
	bucket: IBucket | null = null

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
