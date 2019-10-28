import AbstractService from "./abstractService";
import TaskRelationModel from '../models/taskRelation'

export default class TaskRelationService extends AbstractService {
	constructor() {
		super({
			create: '/tasks/{task_id}/relations',
			delete: '/tasks/{task_id}/relations',
		})
	}
	
	modelFactory(data) {
		return new TaskRelationModel(data)
	}
}