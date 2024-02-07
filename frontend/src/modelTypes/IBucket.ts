import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'

export interface IBucket extends IAbstract {
	id: number
	title: string
	projectId: number
	limit: number
	tasks: ITask[]
	position: number
	count: number
	
	createdBy: IUser
	created: Date
	updated: Date
}