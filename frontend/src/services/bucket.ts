import AbstractService from './abstractService'
import BucketModel from '../models/bucket'
import TaskService from '@/services/task'
import type { IBucket } from '@/modelTypes/IBucket'
import type { ITask } from '@/modelTypes/ITask'

export default class BucketService extends AbstractService<IBucket> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/views/{projectViewId}/buckets',
			create: '/projects/{projectId}/views/{projectViewId}/buckets',
			update: '/projects/{projectId}/views/{projectViewId}/buckets/{id}',
			delete: '/projects/{projectId}/views/{projectViewId}/buckets/{id}',
		})
	}

	modelFactory(data: Partial<IBucket>) {
		return new BucketModel(data)
	}

	beforeUpdate(model: Partial<IBucket>): Partial<IBucket> {
		const taskService = new TaskService()
		model.tasks = model.tasks?.map((t: Partial<ITask>) => taskService.processModel(t))
		return model
	}
}
