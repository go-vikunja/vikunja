import AbstractService from './abstractService'
import TaskCommentModel from '@/models/taskComment'
import type {ITaskComment} from '@/modelTypes/ITaskComment'
import {objectToSnakeCase} from '@/helpers/case'

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

	autoTransformBeforePost(): boolean {
		return false
	}
	
	beforeUpdate(model: ITaskComment) {
		const transformed = objectToSnakeCase({...model})

		// We can't convert emojis to skane case, hence we add them back again
		transformed.reactions = {}
		Object.keys(model.reactions || {}).forEach(reaction => {
			transformed.reactions[reaction] = model.reactions[reaction].map(u => objectToSnakeCase(u))
		})
		
		console.log()

		return transformed as ITaskComment
	}
}
