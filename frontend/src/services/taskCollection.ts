import AbstractService from '@/services/abstractService'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'

export interface TaskFilterParams {
	sort_by: ('start_date' | 'done' | 'id' | 'position')[],
	order_by: ('asc' | 'desc')[],
	filter: string,
	filter_include_nulls: boolean,
	s: string,
}

export function getDefaultTaskFilterParams(): TaskFilterParams {
	return {
		sort_by: ['position', 'id'],
		order_by: ['asc', 'desc'],
		filter: '',
		filter_include_nulls: false,
		s: '',
	}
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