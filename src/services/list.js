import AbstractService from './abstractService'
import ListModel from '../models/list'
import TaskService from './task'

export default class ListService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/lists',
			get: '/lists/{id}',
			update: '/lists/{id}',
			delete: '/lists/{id}',
		})
	}
	
	modelFactory(data) {
		return new ListModel(data)
	}

	beforeUpdate(model) {
		let taskService = new TaskService()
		model.tasks = model.tasks.map(task => {
			return taskService.beforeUpdate(task)
		})
		return model
	}
}