import AbstractModel from './abstractModel'

import type {ITaskAssignee} from '@/modelTypes/ITaskAssignee'
import type {IUser} from '@/modelTypes/IUser'
import type {ITask} from '@/modelTypes/ITask'

export default class TaskAssigneeModel extends AbstractModel<ITaskAssignee> implements ITaskAssignee {
	created: Date = null
	userId: IUser['id'] = 0
	taskId: ITask['id'] = 0

	constructor(data: Partial<ITaskAssignee>) {
		super()
		this.assignData(data)
		this.created = new Date(this.created)
	}
}
