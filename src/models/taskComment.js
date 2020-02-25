import AbstractModel from './abstractModel'
import UserModel from './user'

export default class TaskCommentModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.author = new UserModel(this.author)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			task_id: 0,
			comment: '',
			author: UserModel,
			created: null,
			update: null,
		}
	}
}
