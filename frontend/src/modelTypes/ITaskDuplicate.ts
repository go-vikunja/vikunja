import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'

export interface ITaskDuplicate extends IAbstract {
	taskId: number
	duplicatedTask: ITask | null
}
