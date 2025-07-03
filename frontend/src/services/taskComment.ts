import AbstractService from './abstractService'
import TaskCommentModel from '@/models/taskComment'
import type {ITaskComment} from '@/modelTypes/ITaskComment'
import type {IUser} from '@/modelTypes/IUser'
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

	modelFactory(data: Partial<ITaskComment>): ITaskComment {
		return new TaskCommentModel(data)
	}

	autoTransformBeforePost(): boolean {
		return false
	}
	
	beforeUpdate(model: ITaskComment) {
		const transformed = objectToSnakeCase({...model})

		// We can't convert emojis to snake case, hence we add them back again
		const transformedTyped = transformed as Record<string, unknown> & { reactions: Record<string, unknown> }
		transformedTyped.reactions = {}
		Object.keys(model.reactions || {}).forEach(reaction => {
			transformedTyped.reactions[reaction] = model.reactions![reaction].map((u: IUser) => objectToSnakeCase(u as unknown as Record<string, unknown>))
		})
		
		console.log()

		return transformedTyped as unknown as ITaskComment
	}
}
