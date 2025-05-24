import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'
import type {IProjectView} from '@/modelTypes/IProjectView'

export interface IBucket extends IAbstract {
	id: number
	title: string
	projectId: number
	limit: number
	tasks: ITask[]
	position: number
	count: number
	projectViewId: IProjectView['id']
	
	createdBy: IUser
	created: Date
	updated: Date
}
