import AbstractModel from './abstractModel'
import TaskModel from './task'

import type {ITaskDuplicate} from '@/modelTypes/ITaskDuplicate'
import type {ITask} from '@/modelTypes/ITask'

export default class TaskDuplicateModel extends AbstractModel<ITaskDuplicate> implements ITaskDuplicate {
	taskId = 0
	duplicatedTask: ITask | null = null

	constructor(data: Partial<ITaskDuplicate>) {
		super()
		this.assignData(data)

		this.duplicatedTask = this.duplicatedTask ? new TaskModel(this.duplicatedTask) : null
	}
}
