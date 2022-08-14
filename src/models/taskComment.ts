import AbstractModel, { type IAbstract } from './abstractModel'
import UserModel, { type IUser } from './user'
import type { ITask } from './task'

export interface ITaskComment extends IAbstract {
	id: number
	taskId: ITask['id']
	comment: string
	author: IUser

	created: Date
	updated: Date
}

export default class TaskCommentModel extends AbstractModel implements ITaskComment {
	id!: number
	taskId!: ITask['id']
	comment!: string
	author: IUser

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
