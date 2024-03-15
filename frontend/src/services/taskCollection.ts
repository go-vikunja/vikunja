import AbstractService from '@/services/abstractService'
import TaskModel from '@/models/task'

import type {ITask} from '@/modelTypes/ITask'
import BucketModel from '@/models/bucket'

export interface TaskFilterParams {
	sort_by: ('start_date' | 'end_date' | 'due_date' | 'done' | 'id' | 'position' | 'kanban_position')[],
	order_by: ('asc' | 'desc')[],
	filter: string,
	filter_include_nulls: boolean,
	filter_timezone?: string,
	s: string,
	per_page?: number,
}

export function getDefaultTaskFilterParams(): TaskFilterParams {
	return {
		sort_by: ['position', 'id'],
		order_by: ['asc', 'desc'],
		filter: 'done = false',
		filter_include_nulls: false,
		filter_timezone: '',
		s: '',
	}
}

export default class TaskCollectionService extends AbstractService<ITask> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/views/{viewId}/tasks',
		})
	}

	modelFactory(data) {
		// FIXME: There must be a better way for this…
		if (typeof data.project_view_id !== 'undefined') {
			return new BucketModel(data)
		}
		return new TaskModel(data)
	}
}