import AbstractService from './abstractService'
import TaskCommentModel from '../models/taskComment'
import moment from 'moment'

export default class TaskCommentService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/comments',
			getAll: '/tasks/{task_id}/comments',
			get: '/tasks/{task_id}/comments/{id}',
			update: '/tasks/{task_id}/comments/{id}',
			delete: '/tasks/{task_id}/comments/{id}',
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		model.updated = moment(model.updated).toISOString()
		return model
	}

	modelFactory(data) {
		return new TaskCommentModel(data)
	}
}