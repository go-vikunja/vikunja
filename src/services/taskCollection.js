import AbstractService from './abstractService'
import TaskModel from '../models/task'
import moment from 'moment'

export default class TaskCollectionService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listID}/tasks',
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new TaskModel(data)
	}
}