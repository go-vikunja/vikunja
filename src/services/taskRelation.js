import AbstractService from './abstractService'
import TaskRelationModel from '../models/taskRelation'
import moment from 'moment'

export default class TaskRelationService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/relations',
			delete: '/tasks/{task_id}/relations',
		})
	}

	processModel(model) {
		model.created = moment(model.created).toISOString()
		return model
	}

	modelFactory(data) {
		return new TaskRelationModel(data)
	}
}