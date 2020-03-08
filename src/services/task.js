import AbstractService from './abstractService'
import TaskModel from '../models/task'
import AttachmentService from './attachment'
import {formatISO} from 'date-fns'

export default class TaskService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listID}',
			getAll: '/tasks/all',
			get: '/tasks/{id}',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		});
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

	processModel(model) {
		// Ensure the listID is an int
		model.listID = Number(model.listID)

		// Convert dates into an iso string
		model.dueDate = model.dueDate === null ? null : formatISO(new Date(model.dueDate))
		model.startDate = model.startDate === null ? null : formatISO(new Date(model.startDate))
		model.endDate = model.endDate === null ? null : formatISO(new Date(model.endDate))
		model.created = formatISO(model.created)
		model.updated = formatISO(model.updated)

		// remove all nulls, these would create empty reminders
		for (const index in model.reminderDates) {
			if (model.reminderDates[index] === null) {
				model.reminderDates.splice(index, 1)
			}
		}

		// Make normal timestamps from js dates
		if(model.reminderDates.length > 0) {
			model.reminderDates = model.reminderDates.map(r => {
				return formatISO(new Date(r))
			})
		}

		// Make the repeating amount to seconds
		let repeatAfterSeconds = 0
		if (model.repeatAfter.amount !== null || model.repeatAfter.amount !== 0) {
			switch (model.repeatAfter.type) {
				case 'hours':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60
					break
				case 'days':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24
					break
				case 'weeks':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 7
					break
				case 'months':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 30
					break
				case 'years':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 365
					break
			}
		}
		model.repeatAfter = repeatAfterSeconds

		if (model.hexColor.substring(0, 1) === '#') {
			model.hexColor = model.hexColor.substring(1, 7)
		}

		// Do the same for all related tasks
		Object.keys(model.related_tasks).forEach(relationKind  => {
			model.related_tasks[relationKind] = model.related_tasks[relationKind].map(t => {
				return this.processModel(t)
			})
		})

		// Process all attachments to preven parsing errors
		if(model.attachments.length > 0) {
			const attachmentService = new AttachmentService()
			model.attachments.map(a => {
				return attachmentService.processModel(a)
			})
		}

		return model
	}
}