import AbstractModel from './abstractModel'
import type { ITask } from './task'
import type { IUser } from './user'

export interface ITaskAssignee {
	created: Date
	userId: IUser['id']
	taskId: ITask['id']
}

export default class TaskAssigneeModel extends AbstractModel implements ITaskAssignee {
	created: Date
	declare userId: IUser['id']
	declare taskId: ITask['id']

	constructor(data) {
		super(data)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			created: null,
			userId: 0,
			taskId: 0,
		}
	}
}
