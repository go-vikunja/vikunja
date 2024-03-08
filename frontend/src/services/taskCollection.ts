import AbstractService from '@/services/abstractService'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'

export interface TaskFilterParams {
	sort_by: ('start_date' | 'done' | 'id')[],
	order_by: ('asc' | 'desc')[],
	filter: string,
	filter_include_nulls: boolean,
	s: string,
}

export default class TaskCollectionService extends AbstractService<ITask> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/tasks',
		})
	}

	modelFactory(data) {
		return new TaskModel(data)
	}
}