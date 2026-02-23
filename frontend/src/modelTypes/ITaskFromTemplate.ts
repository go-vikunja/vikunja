import type {IAbstract} from './IAbstract'
import type {ITask} from './ITask'
import type {IProject} from './IProject'
import type {ITaskTemplate} from './ITaskTemplate'

export interface ITaskFromTemplate extends IAbstract {
	templateId: ITaskTemplate['id']
	targetProjectId: IProject['id']
	title: string
	createdTask: ITask | null
}
