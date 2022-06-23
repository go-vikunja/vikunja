import AbstractModel from './abstractModel'
import type UserModel from './user'
import type TaskModel from './task'

export default class TaskAssigneeModel extends AbstractModel {
	created: Date
	userId: UserModel['id']
	taskId: TaskModel['id']

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
