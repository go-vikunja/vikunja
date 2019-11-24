import AbstractService from './abstractService'
import TaskAssigneeModel from '../models/taskAssignee'

export default class TaskAssigneeService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/assignees',
			delete: '/tasks/{task_id}/assignees/{user_id}',
		})
	}
	
	modelFactory(data) {
		return new TaskAssigneeModel(data)
	}
}