import AbstractService from './abstractService'
import TaskCommentModel from '@/models/taskComment'
import type {ITaskComment} from '@/modelTypes/ITaskComment'

export default class TaskCommentService extends AbstractService<ITaskComment> {
	constructor() {
		super({
			create: '/tasks/{taskId}/comments',
			getAll: '/tasks/{taskId}/comments',
			get: '/tasks/{taskId}/comments/{id}',
			update: '/tasks/{taskId}/comments/{id}',
			delete: '/tasks/{taskId}/comments/{id}',
		})
	}

	modelFactory(data) {
		return new TaskCommentModel(data)
	}
}