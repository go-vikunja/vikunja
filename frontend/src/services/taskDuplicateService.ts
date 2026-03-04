import AbstractService from './abstractService'
import TaskDuplicateModel from '@/models/taskDuplicateModel'
import type {ITaskDuplicate} from '@/modelTypes/ITaskDuplicate'

export default class TaskDuplicateService extends AbstractService<ITaskDuplicate> {
	constructor() {
		super({
			create: '/tasks/{taskId}/duplicate',
		})
	}

	modelFactory(data: Partial<ITaskDuplicate>) {
		return new TaskDuplicateModel(data)
	}
}
