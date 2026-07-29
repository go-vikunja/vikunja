import AbstractService from './abstractService'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'

import {colorFromHex} from '@/helpers/color/colorFromHex'
import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_WEEK} from '@/constants/date'
import {objectToSnakeCase} from '@/helpers/case'
import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'

// Fields a task model carries that the api never reads on write, either because the
// server owns them or because they only exist on the client. Stripping these rather than
// listing what to send means a new writable field is sent by default; v2 would reject the
// ones it does not know, and v1 ignores them.
const READ_ONLY_FIELDS = new Set([
	'attachments',
	'buckets',
	'comments',
	'created_by',
	'identifier',
	'index',
	'labels',
	'max_permission',
	'parent_task_id',
	'position',
	'reactions',
	'related_tasks',
	'reminder_dates',
	'repeat_from_current_date',
	'subscription',
])

// Nested models (assignees, reminders) carry maxPermission of their own, hence recursing.
function withoutReadOnlyFields<T>(value: T): T {
	if (Array.isArray(value)) {
		return value.map(withoutReadOnlyFields) as T
	}

	if (value === null || typeof value !== 'object' || value instanceof Date) {
		return value
	}

	return Object.fromEntries(
		Object.entries(value)
			.filter(([key]) => !READ_ONLY_FIELDS.has(key))
			.map(([key, nested]) => [key, withoutReadOnlyFields(nested)]),
	) as T
}

const parseDate = date => {
	if (date) {
		return new Date(date).toISOString()
	}

	return null
}

export default class TaskService extends AbstractService<ITask> {
	constructor() {
		super({
			create: '/projects/{projectId}/tasks',
			getAll: '/tasks',
			get: '/tasks/{id}',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		})
	}

	modelFactory(data) {
		return new TaskModel(data)
	}

	beforeUpdate(model) {
		return this.processModel(model)
	}

	beforeCreate(model) {
		return this.processModel(model)
	}

	autoTransformBeforePost(): boolean {
		return false
	}

	processModel(updatedModel) {
		const model = {...updatedModel}

		model.title = model.title?.trim()

		// Ensure that projectId is an int
		model.projectId = Number(model.projectId)

		// Convert dates into an iso string
		model.dueDate = parseDate(model.dueDate)
		model.startDate = parseDate(model.startDate)
		model.endDate = parseDate(model.endDate)
		model.doneAt = parseDate(model.doneAt)
		model.deletedAt = parseDate(model.deletedAt)
		model.created = new Date(model.created).toISOString()
		model.updated = new Date(model.updated).toISOString()

		// remove all nulls, these would create empty reminders
		model.reminders = model.reminders.filter(r => r !== null)
		// Make normal timestamps from js dates
		if (model.reminders.length > 0) {
			model.reminders.forEach(r => {
				r.reminder = new Date(r.reminder).toISOString()
			})
		}

		// Make the repeating amount to seconds
		let repeatAfterSeconds = 0
		if (model.repeatAfter !== null && (model.repeatAfter.amount !== null || model.repeatAfter.amount !== 0)) {
			switch (model.repeatAfter.type) {
				case 'hours':
					repeatAfterSeconds = model.repeatAfter.amount * SECONDS_A_HOUR
					break
				case 'days':
					repeatAfterSeconds = model.repeatAfter.amount * SECONDS_A_DAY
					break
				case 'weeks':
					repeatAfterSeconds = model.repeatAfter.amount * SECONDS_A_WEEK
					break
			}
		}
		model.repeatAfter = repeatAfterSeconds

		model.hexColor = colorFromHex(model.hexColor)

		// The api only ever reads the ids, and the rest of a user model does not pass v2
		// validation
		model.assignees = (model.assignees ?? []).map((assignee: {id: number}) => ({id: assignee.id}))

		return withoutReadOnlyFields(objectToSnakeCase(model)) as ITask
	}

	/**
	 * Creates multiple tasks in one request. The api spreads them out over the top of the
	 * views they land in, so a batch keeps the order it is passed in without the client
	 * having to compute positions. Fails as a whole: either all tasks are created or none.
	 */
	async createMultiple(tasks: ITask[]): Promise<ITask[]> {
		const cancel = this.setLoading()

		try {
			// v2 only, so this bypasses the v1 baseURL the rest of this service uses
			const {data} = await AuthenticatedHTTPFactory().post(apiV2Url('tasks/bulk'), {
				tasks: tasks.map(task => this.processModel(task)),
			})
			return data.tasks.map((task: Partial<ITask>) => this.modelCreateFactory(task))
		} finally {
			cancel()
		}
	}

	async markTaskAsRead(taskId: ITask['id']): Promise<void> {
		const cancel = this.setLoading()
	
		try {
			await AuthenticatedHTTPFactory().post(`/tasks/${taskId}/read`, {} as ITask)
		} finally {
			cancel()
		}
	}
}

