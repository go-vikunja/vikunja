import AbstractService from './abstractService'
import TaskRelationModel from '@/models/taskRelation'
import type {ITaskRelation} from '@/modelTypes/ITaskRelation'

export default class TaskRelationService extends AbstractService<ITaskRelation> {
	constructor() {
		super({
			create: '/tasks/{taskId}/relations',
			delete: '/tasks/{taskId}/relations/{relationKind}/{otherTaskId}',
		})
	}

	modelFactory(data) {
		return new TaskRelationModel(data)
	}
}
