import AbstractModel from './abstractModel'
import UserModel from './user'

import type {ITaskRelation} from '@/modelTypes/ITaskRelation'
import type {ITask} from '@/modelTypes/ITask'
import type {IUser} from '@/modelTypes/IUser'

import type {IRelationKind} from '@/types/IRelationKind'
export default class TaskRelationModel extends AbstractModel implements ITaskRelation {
	id = 0
	otherTaskId: ITask['id'] = 0
	taskId: ITask['id'] = 0
	relationKind: IRelationKind = ''

	createdBy: IUser = UserModel
	created: Date = null

	constructor(data: Partial<ITaskRelation>) {
		super()
		this.assignData(data)

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
	}
}