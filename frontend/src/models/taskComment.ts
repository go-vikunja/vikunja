import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ITaskComment} from '@/modelTypes/ITaskComment'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

export default class TaskCommentModel extends AbstractModel<ITaskComment> implements ITaskComment {
	id = 0
	taskId: ITask['id'] = 0
	comment = ''
	author: IUser = new UserModel()
	
	reactions = {}

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITaskComment> = {}) {
		super()
		this.assignData(data)

		this.author = new UserModel(this.author)
		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
		
		// We can't convert emojis to camel case, hence we do this manually
		this.reactions = {}
		Object.keys(data.reactions || {}).forEach(reaction => {
			this.reactions[reaction] = (data.reactions as any)[reaction].map((u: any) => new UserModel(u))
		})
	}
}
