import AbstractService from './abstractService'
import TaskCommentModel from '../models/taskComment'
import {formatISO} from 'date-fns'

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
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)
		return model
	}

	modelFactory(data) {
		return new TaskCommentModel(data)
	}
}