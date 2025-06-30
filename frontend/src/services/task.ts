import AbstractService from './abstractService'
import TaskModel from '@/models/task'
import type {ITask} from '@/modelTypes/ITask'
import type {ITaskReminder} from '@/modelTypes/ITaskReminder'
import type {IAttachment} from '@/modelTypes/IAttachment'
import type {ILabel} from '@/modelTypes/ILabel'
import type {IUser} from '@/modelTypes/IUser'
import AttachmentService from './attachment'
import LabelService from './label'

import {colorFromHex} from '@/helpers/color/colorFromHex'
import {SECONDS_A_DAY, SECONDS_A_HOUR, SECONDS_A_WEEK} from '@/constants/date'
import {objectToSnakeCase} from '@/helpers/case'

const parseDate = (date: string | Date | null | undefined): string | null => {
	if (date) {
		return new Date(date).toISOString()
	}

	return null
}

export default class TaskService extends AbstractService<ITask> {
	constructor() {
		super({
			create: '/projects/{projectId}/tasks',
			getAll: '/tasks/all',
			get: '/tasks/{id}',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		})
	}

	modelFactory(data: Partial<ITask>) {
		return new TaskModel(data)
	}

	beforeUpdate(model: ITask) {
		return this.processModel(model)
	}

	beforeCreate(model: ITask) {
		return this.processModel(model)
	}

	autoTransformBeforePost(): boolean {
		return false
	}

	processModel(updatedModel: ITask) {
		const model = {...updatedModel}

		if (model.title) {
			model.title = model.title.trim()
		}

		// Ensure that projectId is an int
		if (model.projectId !== null && model.projectId !== undefined) {
			model.projectId = Number(model.projectId)
		}

		// Convert dates into an iso string
		(model as any).dueDate = parseDate(model.dueDate)
		(model as any).startDate = parseDate(model.startDate)
		(model as any).endDate = parseDate(model.endDate)
		(model as any).doneAt = parseDate(model.doneAt)
		(model as any).created = new Date(model.created || Date.now()).toISOString()
		(model as any).updated = new Date(model.updated || Date.now()).toISOString()

		(model as any).reminderDates = null
		// remove all nulls, these would create empty reminders
		if (model.reminders) {
			for (const index in model.reminders) {
				if (model.reminders[index] === null) {
					model.reminders.splice(Number(index), 1)
				}
			}
		}
		// Make normal timestamps from js dates
		if (model.reminders && model.reminders.length > 0) {
			model.reminders.forEach((r: ITaskReminder) => {
				if (r && r.reminder) {
					(r as any).reminder = new Date(r.reminder).toISOString()
				}
			})
		}

		// Make the repeating amount to seconds
		let repeatAfterSeconds = 0
		if (model.repeatAfter !== null && typeof model.repeatAfter === 'object' && model.repeatAfter.amount !== null && model.repeatAfter.amount !== 0) {
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
		(model as any).repeatAfter = repeatAfterSeconds

		model.hexColor = colorFromHex(model.hexColor)

		// Do the same for all related tasks
		if (model.relatedTasks) {
			Object.keys(model.relatedTasks).forEach(relationKind => {
				const tasks = (model.relatedTasks as any)[relationKind]
				if (tasks) {
					(model.relatedTasks as any)[relationKind] = tasks.map((t: ITask) => {
						return this.processModel(t)
					})
				}
			})
		}

		// Process all attachments to prevent parsing errors
		if (model.attachments && model.attachments.length > 0) {
			const attachmentService = new AttachmentService()
			;(model as any).attachments = model.attachments.map((a: IAttachment) => {
				return attachmentService.processModel(a)
			})
		}

		// Preprocess all labels
		if (model.labels && model.labels.length > 0) {
			const labelService = new LabelService()
			;(model as any).labels = model.labels.map((l: ILabel) => labelService.processModel(l))
		}

		const transformed = objectToSnakeCase(model)

		// We can't convert emojis to skane case, hence we add them back again
		transformed.reactions = {}
		Object.keys(updatedModel.reactions || {}).forEach(reaction => {
			transformed.reactions[reaction] = updatedModel.reactions[reaction].map((u: IUser) => objectToSnakeCase(u))
		})

		return transformed as ITask
	}
}

