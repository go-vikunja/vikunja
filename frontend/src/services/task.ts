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

const parseDate = (date: string | Date | null | undefined) => {
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

		model.title = model.title?.trim()

		// Ensure that projectId is an int
		model.projectId = Number(model.projectId)

		// Convert dates into an iso string
		model.dueDate = parseDate(model.dueDate)
		model.startDate = parseDate(model.startDate)
		model.endDate = parseDate(model.endDate)
		model.doneAt = parseDate(model.doneAt)
		model.created = new Date(model.created || Date.now()).toISOString()
		model.updated = new Date(model.updated || Date.now()).toISOString()

		model.reminderDates = null
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
					r.reminder = new Date(r.reminder).toISOString()
				}
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
		if (model.relatedTasks) {
			Object.keys(model.relatedTasks).forEach(relationKind => {
				model.relatedTasks[relationKind] = model.relatedTasks[relationKind]!.map((t: ITask) => {
					return this.processModel(t)
				})
			})
		}

		// Process all attachments to prevent parsing errors
		if (model.attachments && model.attachments.length > 0) {
			const attachmentService = new AttachmentService()
			model.attachments = model.attachments.map((a: IAttachment) => {
				return attachmentService.processModel(a)
			})
		}

		// Preprocess all labels
		if (model.labels && model.labels.length > 0) {
			const labelService = new LabelService()
			model.labels = model.labels.map((l: ILabel) => labelService.processModel(l))
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

