import AbstractService from './abstractService'
import ListModel from '../models/list'
import TaskService from './task'
import moment from 'moment'

export default class ListService extends AbstractService {
	constructor() {
		super({
			create: '/namespaces/{namespaceID}/lists',
			get: '/lists/{id}',
			update: '/lists/{id}',
			delete: '/lists/{id}',
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
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