import type {IProjectView} from '@/modelTypes/IProjectView'
import type {IAbstract} from '@/modelTypes/IAbstract'
import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

export interface ITaskBucket extends IAbstract {
	taskId: ITask['id']
	bucketId: IBucket['id']
	projectViewId: IProjectView['id']
	projectId: IProject['id']
	task: ?ITask
	bucket: ?IBucket
}
