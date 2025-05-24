import AbstractModel from './abstractModel'
import UserModel from './user'
import TaskModel from './task'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

export default class BucketModel extends AbstractModel<IBucket> implements IBucket {
	id = 0
	title = ''
	projectId = ''
	limit = 0
	tasks: ITask[] = []
	position = 0
	count = 0
	
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
