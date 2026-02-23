import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IProject} from './IProject'

export interface ITaskDuplicate extends IAbstract {
	taskId: ITask['id']
	targetProjectId: IProject['id']
	duplicatedTask: ITask | null
}
