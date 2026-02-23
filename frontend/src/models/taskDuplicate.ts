import AbstractModel from './abstractModel'
import TaskModel from './task'

import type {ITaskDuplicate} from '@/modelTypes/ITaskDuplicate'
import type {ITask} from '@/modelTypes/ITask'
import type {IProject} from '@/modelTypes/IProject'

export default class TaskDuplicateModel extends AbstractModel<ITaskDuplicate> implements ITaskDuplicate {
	taskId: ITask['id'] = 0
	targetProjectId: IProject['id'] = 0
	duplicatedTask: ITask | null = null

	constructor(data: Partial<ITaskDuplicate> = {}) {
		super()
		this.assignData(data)

		this.duplicatedTask = this.duplicatedTask ? new TaskModel(this.duplicatedTask) : null
	}
}
