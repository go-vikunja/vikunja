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

	reactions: Record<string, IUser[]> = {}

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<ITaskComment> = {}) {
		super()
		this.assignData(data)

		this.author = new UserModel(this.author)
		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()

		// We can't convert emojis to camel case, hence we do this manually
		this.reactions = {}
		const dataWithReactions = data as ITaskComment & { reactions?: Record<string, IUser[]> }
		Object.keys(dataWithReactions.reactions || {}).forEach((reaction: string) => {
			if (dataWithReactions.reactions && dataWithReactions.reactions[reaction]) {
				this.reactions[reaction] = dataWithReactions.reactions[reaction].map((u: IUser) => new UserModel(u))
			}
		})
	}
}
