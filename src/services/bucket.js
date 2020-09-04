import AbstractService from './abstractService'
import BucketModel from '../models/bucket'
import TaskService from '@/services/task'

export default class BucketService extends AbstractService {
	constructor() {
		super({
			getAll: '/lists/{listId}/buckets',
			create: '/lists/{listId}/buckets',
			update: '/lists/{listId}/buckets/{id}',
			delete: '/lists/{listId}/buckets/{id}',
		})
	}

	modelFactory(data) {
		return new BucketModel(data)
	}

	beforeUpdate(model) {
		const taskService = new TaskService()
		model.tasks = model.tasks.map(t => taskService.processModel(t))
		return model
	}
}