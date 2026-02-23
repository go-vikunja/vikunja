import AbstractService from './abstractService'
import TaskDuplicateModel from '@/models/taskDuplicate'
import type {ITaskDuplicate} from '@/modelTypes/ITaskDuplicate'

export default class TaskDuplicateService extends AbstractService<ITaskDuplicate> {
	constructor() {
		super({
			create: '/tasks/{taskId}/duplicate',
		})
	}

	modelFactory(data) {
		return new TaskDuplicateModel(data)
	}
}
