import AbstractService from './abstractService'
import TaskAssigneeModel, { type ITaskAssignee } from '../models/taskAssignee'
import {formatISO} from 'date-fns'

export default class TaskAssigneeService extends AbstractService<ITaskAssignee> {
	constructor() {
		super({
			create: '/tasks/{taskId}/assignees',
			delete: '/tasks/{taskId}/assignees/{userId}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		return model
	}

	modelFactory(data) {
		return new TaskAssigneeModel(data)
	}
}