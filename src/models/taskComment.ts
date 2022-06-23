import AbstractModel from './abstractModel'
import UserModel from './user'
import type TaskModel from './task'

export default class TaskCommentModel extends AbstractModel {
	id: number
	taskId: TaskModel['id']
	comment: string
	author: UserModel

	created: Date
	updated: Date

	constructor(data) {
		super(data)
		this.author = new UserModel(this.author)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			taskId: 0,
			comment: '',
			author: UserModel,
			created: null,
			updated: null,
		}
	}
}
