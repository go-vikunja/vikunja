import AbstractService from './abstractService'
import TaskDuplicateModel from '@/models/taskDuplicateModel'
import type {ITaskDuplicate} from '@/modelTypes/ITaskDuplicate'

export default class TaskDuplicateService extends AbstractService<ITaskDuplicate> {
	constructor() {
		super({
			create: '/projects/{projectId}/tasks/{taskId}/duplicate',
		})
	}

	beforeCreate(model: ITaskDuplicate) {
		model.duplicatedTask = null
		return model
	}

	modelFactory(data: Partial<ITaskDuplicate>) {
		return new TaskDuplicateModel(data)
	}
}
