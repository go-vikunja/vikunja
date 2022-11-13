import AbstractService from '@/services/abstractService'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'

// FIXME: unite with other filter params types
export interface GetAllTasksParams {
	sort_by: ('start_date' | 'done' | 'id')[],
	order_by: ('asc' | 'asc' | 'desc')[],
	filter_by: 'start_date'[],
	filter_comparator: ('greater_equals' | 'less_equals')[],
	filter_value: [string, string] // [dateFrom, dateTo],
	filter_concat: 'and',
	filter_include_nulls: boolean,
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