import AbstractService from './abstractService'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import AttachmentService from './attachment'
import LabelService from './label'

import {colorFromHex} from '@/helpers/color/colorFromHex'
import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_WEEK} from '@/constants/date'
import {objectToSnakeCase} from '@/helpers/case'
import {AuthenticatedHTTPFactory, apiV2Url} from '@/helpers/fetcher'

// The fields a client may set when creating a task. v2 rejects properties it does not
// know, and a task model carries plenty the api does not accept on write: the empty
// createdBy user it defaults to, maxPermission on every nested model, and frontend-only
// ones like parentTaskId.
const CREATE_FIELDS = [
	'title',
	'description',
	'done',
	'due_date',
	'start_date',
	'end_date',
	'reminders',
	'repeat_after',
	'repeat_mode',
	'priority',
	'hex_color',
	'percent_done',
	'project_id',
	'bucket_id',
	'position',
	'index',
	'is_favorite',
]

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

		model.reminderDates = null
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

		// Do the same for all related tasks
		Object.keys(model.relatedTasks).forEach(relationKind => {
			model.relatedTasks[relationKind] = model.relatedTasks[relationKind].map(t => {
				return this.processModel(t)
			})
		})

		// Process all attachments to prevent parsing errors
		if (model.attachments.length > 0) {
			const attachmentService = new AttachmentService()
			model.attachments.map(a => {
				return attachmentService.processModel(a)
			})
		}

		// Preprocess all labels
		if (model.labels.length > 0) {
			const labelService = new LabelService()
			model.labels = model.labels.map(l => labelService.processModel(l))
		}

		const transformed = objectToSnakeCase(model)

		// We can't convert emojis to skane case, hence we add them back again
		transformed.reactions = {}
		Object.keys(updatedModel.reactions || {}).forEach(reaction => {
			transformed.reactions[reaction] = updatedModel.reactions[reaction].map(u => objectToSnakeCase(u))
		})

		return transformed as ITask
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
				tasks: tasks.map(task => this.createPayload(task)),
			})
			return data.tasks.map((task: Partial<ITask>) => this.modelCreateFactory(task))
		} finally {
			cancel()
		}
	}

	createPayload(task: ITask): Record<string, unknown> {
		const processed = this.processModel(task) as unknown as Record<string, unknown>
		const payload = Object.fromEntries(
			CREATE_FIELDS
				.filter(field => processed[field] !== null && typeof processed[field] !== 'undefined')
				.map(field => [field, processed[field]]),
		)

		// The api resolves assignees by id, and the rest of the user model does not survive
		// validation. Reminders are picked apart for the same reason.
		payload.assignees = (task.assignees ?? []).map(({id}) => ({id}))
		payload.reminders = ((processed.reminders ?? []) as Record<string, unknown>[])
			.map(reminder => ({
				reminder: reminder.reminder,
				relative_period: reminder.relative_period,
				relative_to: reminder.relative_to,
			}))

		return payload
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

