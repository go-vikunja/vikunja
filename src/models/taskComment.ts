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
	id = 0
	taskId: ITask['id'] = 0
	comment = ''
	author: IUser = UserModel

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITaskComment>) {
		super()
		this.assignData(data)

		this.author = new UserModel(this.author)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
