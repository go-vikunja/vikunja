import AbstractService from './abstractService'
import ListModel from '../models/list'
import TaskService from './task'
import {formatISO} from 'date-fns'

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
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
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
		model.hex_color = model.hex_color.substring(1, 7)
		return model
	}

	beforeCreate(list) {
		list.hex_color = list.hex_color.substring(1, 7)
		return list
	}
}