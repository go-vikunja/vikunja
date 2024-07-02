import AbstractModel from '@/models/abstractModel'
import type {ITaskBucket} from '@/modelTypes/ITaskBucket'

export default class TaskBucketModel extends AbstractModel<ITaskBucket> implements ITaskBucket {
	taskId = 0
	bucketId = 0
	projectViewId = 0
	projectId = 0

	constructor(data: Partial<ITaskBucket>) {
		super()
		this.assignData(data)
	}
}
