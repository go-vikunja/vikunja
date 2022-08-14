import AbstractModel, { type IAbstract } from './abstractModel'
import UserModel, { type IUser } from './user'
import TaskModel, { type ITask } from './task'

export interface IBucket extends IAbstract {
	id: number
	title: string
	listId: number
	limit: number
	tasks: ITask[]
	isDoneBucket: boolean
	position: number
	
	createdBy: IUser
	created: Date
	updated: Date
}

export default class BucketModel extends AbstractModel implements IBucket {
	id = 0
	title = ''
	listId = ''
	limit = 0
	tasks: ITask[] = []
	isDoneBucket: false
	position: 0
	
	createdBy: IUser = null
	created: Date = null
	updated: Date = null

	constructor(data: Partial<IBucket>) {
		super()
		this.assignData(data)

		this.tasks = this.tasks.map(t => new TaskModel(t))

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}