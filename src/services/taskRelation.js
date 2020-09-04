import AbstractService from './abstractService'
import TaskRelationModel from '../models/taskRelation'
import {formatISO} from 'date-fns'

export default class TaskRelationService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{taskId}/relations',
			delete: '/tasks/{taskId}/relations',
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