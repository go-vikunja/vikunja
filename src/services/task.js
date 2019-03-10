import AbstractService from './abstractService'
import TaskModel from '../models/task'

export default class TaskService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listID}',
			getAll: '/tasks/all',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		});
	}
	
	modelFactory(data) {
		return new TaskModel(data)
	}

	beforeUpdate(model) {
		// Convert the date in a unix timestamp
		model.dueDate = +new Date(model.dueDate) / 1000
		model.startDate = +new Date(model.startDate) / 1000
		model.endDate = +new Date(model.endDate) / 1000

		// remove all nulls, these would create empty reminders
		for (const index in model.reminderDates) {
			if (model.reminderDates[index] === null) {
				model.reminderDates.splice(index, 1)
			}
		}

		// Make normal timestamps from js dates
		model.reminderDates = model.reminderDates.map(r => {
			return Math.round(+new Date(r) / 1000)
		})

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

		return model
	}
}