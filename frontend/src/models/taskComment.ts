import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ITaskComment} from '@/modelTypes/ITaskComment'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

export default class TaskCommentModel extends AbstractModel<ITaskComment> implements ITaskComment {
	id = 0
	taskId: ITask['id'] = 0
	comment = ''
	author: IUser = UserModel
	
	reactions = {}

	created: Date = null
	updated: Date = null

	constructor(data: Partial<ITaskComment> = {}) {
		super()
		this.assignData(data)

		this.author = new UserModel(this.author)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
		
		// We can't convert emojis to camel case, hence we do this manually
		this.reactions = {}
		Object.keys(data.reactions || {}).forEach(reaction => {
			this.reactions[reaction] = data.reactions[reaction].map(u => new UserModel(u))
		})
	}
}
