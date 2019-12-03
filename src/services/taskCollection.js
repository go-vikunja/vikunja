import AbstractService from './abstractService'
import TaskModel from '../models/task'

export default class TaskCollectionService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listID}/tasks',
		})
	}
	
	modelFactory(data) {
		return new TaskModel(data)
	}
}