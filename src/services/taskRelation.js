import AbstractService from './abstractService'
import TaskRelationModel from '../models/taskRelation'
import {formatISO} from 'date-fns'

export default class TaskRelationService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/relations',
			delete: '/tasks/{task_id}/relations',
		})
	}

	processModel(model) {
		model.created = formatISO(model.created)
		return model
	}

	modelFactory(data) {
		return new TaskRelationModel(data)
	}
}