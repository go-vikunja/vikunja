import {formatISO} from 'date-fns'

import AbstractService from './abstractService'
import TaskRelationModel, {type ITaskRelation} from '@/models/taskRelation'

export default class TaskRelationService extends AbstractService<ITaskRelation> {
	constructor() {
		super({
			create: '/tasks/{taskId}/relations',
			delete: '/tasks/{taskId}/relations/{relationKind}/{otherTaskId}',
		})
	}

	processModel(model) {
		model.created = formatISO(new Date(model.created))
		return model
	}

	modelFactory(data) {
		return new TaskRelationModel(data)
	}
}