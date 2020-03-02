import AbstractService from './abstractService'
import TaskAssigneeModel from '../models/taskAssignee'
import {formatISO} from 'date-fns'

export default class TaskAssigneeService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/assignees',
			delete: '/tasks/{task_id}/assignees/{user_id}',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		return model
	}

	modelFactory(data) {
		return new TaskAssigneeModel(data)
	}
}