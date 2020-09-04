import AbstractService from './abstractService'
import TaskCommentModel from '../models/taskComment'
import {formatISO} from 'date-fns'

export default class TaskCommentService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskId}/comments',
			getAll: '/tasks/{taskId}/comments',
			get: '/tasks/{taskId}/comments/{id}',
			update: '/tasks/{taskId}/comments/{id}',
			delete: '/tasks/{taskId}/comments/{id}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		model.updated = formatISO(new Date(model.updated))
		return model
	}

	modelFactory(data) {
		return new TaskCommentModel(data)
	}
}