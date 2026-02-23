import AbstractModel from './abstractModel'
import TaskModel from './task'

import type {ITaskFromTemplate} from '@/modelTypes/ITaskFromTemplate'
import type {ITask} from '@/modelTypes/ITask'

export default class TaskFromTemplateModel extends AbstractModel<ITaskFromTemplate> implements ITaskFromTemplate {
	templateId = 0
	targetProjectId = 0
	title = ''
	createdTask: ITask | null = null

	constructor(data: Partial<ITaskFromTemplate> = {}) {
		super()
		this.assignData(data)

		this.createdTask = this.createdTask ? new TaskModel(this.createdTask) : null
	}
}
