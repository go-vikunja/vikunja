import AbstractService from './abstractService'
import BucketModel from '../models/bucket'
import TaskService from '@/services/task'
import type { IBucket } from '@/modelTypes/IBucket'

export default class BucketService extends AbstractService<IBucket> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/buckets',
			create: '/projects/{projectId}/buckets',
			update: '/projects/{projectId}/buckets/{id}',
			delete: '/projects/{projectId}/buckets/{id}',
		})
	}

	modelFactory(data: Partial<IBucket>) {
		return new BucketModel(data)
	}

	beforeUpdate(model) {
		const taskService = new TaskService()
		model.tasks = model.tasks?.map(t => taskService.processModel(t))
		return model
	}
}