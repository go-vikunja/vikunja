import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'

export interface ITaskBucket extends IAbstract {
	taskId: ITask['id']
	bucketId: IBucket['id']
	projectViewId: number
	projectId: number
	task: ?ITask
	bucket: ?IBucket
}
