import AbstractModel from './abstractModel'
import UserModel from './user'
import TaskModel from './task'

import type {IBucket} from '@/modelTypes/IBucket'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

export default class BucketModel extends AbstractModel<IBucket> implements IBucket {
	id = 0
	title = ''
	projectId = 0
	limit = 0
	tasks: ITask[] = []
	position = 0
	count = 0
	projectViewId = 0

	createdBy: IUser = {} as IUser
	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IBucket>) {
		super()
		this.assignData(data)

		this.tasks = this.tasks.map(t => new TaskModel(t))

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
