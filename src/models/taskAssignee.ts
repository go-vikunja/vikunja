import AbstractModel, { type IAbstract } from './abstractModel'
import type { ITask } from './task'
import type { IUser } from './user'

export interface ITaskAssignee extends IAbstract {
	created: Date
	userId: IUser['id']
	taskId: ITask['id']
}

export default class TaskAssigneeModel extends AbstractModel implements ITaskAssignee {
	created: Date = null
	userId: IUser['id'] = 0
	taskId: ITask['id'] = 0

	constructor(data: Partial<ITaskAssignee>) {
		super()
		this.assignData(data)
		this.created = new Date(this.created)
	}
}
