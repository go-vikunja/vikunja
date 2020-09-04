import AbstractService from './abstractService'
import TaskModel from '../models/task'
import {formatISO} from 'date-fns'

export default class TaskCollectionService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listId}/tasks',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new TaskModel(data)
	}
}